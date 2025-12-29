//go:build windows

/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	user32                      = syscall.NewLazyDLL("user32.dll")
	procSetEnvironmentVariableW = kernel32.NewProc("SetEnvironmentVariableW")
	procSendMessageTimeoutW     = user32.NewProc("SendMessageTimeoutW")
)

const (
	HWND_BROADCAST   = uintptr(0xffff)
	WM_SETTINGCHANGE = 0x001A
	SMTO_ABORTIFHUNG = 0x0002
)

// setUserEnv sets a user-level environment variable in the Windows registry
func setUserEnv(name, value string) error {
	key, err := syscall.UTF16PtrFromString(`Environment`)
	if err != nil {
		return err
	}

	var regKey syscall.Handle
	err = syscall.RegOpenKeyEx(
		syscall.HKEY_CURRENT_USER,
		key,
		0,
		syscall.KEY_WRITE,
		&regKey,
	)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer syscall.RegCloseKey(regKey)

	valuePtr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}

	err = syscall.RegSetValueEx(
		regKey,
		namePtr,
		0,
		syscall.REG_EXPAND_SZ,
		(*byte)(unsafe.Pointer(valuePtr)),
		uint32((len(value)+1)*2),
	)
	if err != nil {
		return fmt.Errorf("set registry value: %w", err)
	}

	return nil
}

// broadcastSettingChange notifies the system that environment variables have changed
func broadcastSettingChange() {
	environment, _ := syscall.UTF16PtrFromString("Environment")
	procSendMessageTimeoutW.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(environment)),
		SMTO_ABORTIFHUNG,
		5000,
		0,
	)
}

func SetGoEnvs() error {
	path := os.Getenv("PATH")
	if path == "" {
		return ErrNoPath
	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir of user: %w", err)
	}

	goPath := filepath.Join(homedir, gmDir, workspace)
	goBin := filepath.Join(goPath, "bin")
	goRoot := filepath.Join(homedir, gmDir, versions, current)
	goSDKBin := filepath.Join(goRoot, "bin")

	// Set user-level environment variables
	if err := setUserEnv("GOPATH", goPath); err != nil {
		return fmt.Errorf("set GOPATH: %w", err)
	}
	if err := setUserEnv("GOBIN", goBin); err != nil {
		return fmt.Errorf("set GOBIN: %w", err)
	}
	if err := setUserEnv("GOROOT", goRoot); err != nil {
		return fmt.Errorf("set GOROOT: %w", err)
	}

	// Update PATH - prepend Go SDK and Go bin if not already present
	newPathEntries := []string{goSDKBin, goBin}
	pathParts := strings.Split(path, ";")

	for _, entry := range newPathEntries {
		found := false
		for _, part := range pathParts {
			if strings.EqualFold(strings.TrimSpace(part), entry) {
				found = true
				break
			}
		}
		if !found {
			pathParts = append([]string{entry}, pathParts...)
		}
	}

	newPath := strings.Join(pathParts, ";")
	if err := setUserEnv("PATH", newPath); err != nil {
		return fmt.Errorf("set PATH: %w", err)
	}

	// Notify system of environment change
	broadcastSettingChange()

	fmt.Println("✅ Environment variables set successfully")
	fmt.Printf("   GOPATH: %s\n", goPath)
	fmt.Printf("   GOBIN: %s\n", goBin)
	fmt.Printf("   GOROOT: %s\n", goRoot)
	fmt.Println("\nNote: Restart your terminal for changes to take effect in new sessions")

	return nil
}
