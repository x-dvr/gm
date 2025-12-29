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
	advapi32                    = syscall.NewLazyDLL("advapi32.dll")
	procSetEnvironmentVariableW = kernel32.NewProc("SetEnvironmentVariableW")
	procSendMessageTimeoutW     = user32.NewProc("SendMessageTimeoutW")
	procRegOpenKeyExW           = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW          = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey             = advapi32.NewProc("RegCloseKey")
)

const (
	HWND_BROADCAST   = uintptr(0xffff)
	WM_SETTINGCHANGE = 0x001A
	SMTO_ABORTIFHUNG = 0x0002

	HKEY_CURRENT_USER = 0x80000001
	KEY_WRITE         = 0x20006
	REG_EXPAND_SZ     = 2
)

// setUserEnv sets a user-level environment variable in the Windows registry
func setUserEnv(name, value string) error {
	key, err := syscall.UTF16PtrFromString(`Environment`)
	if err != nil {
		return err
	}

	var regKey uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		uintptr(HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(key)),
		0,
		KEY_WRITE,
		uintptr(unsafe.Pointer(&regKey)),
	)
	if ret != 0 {
		return fmt.Errorf("open registry key: error code %d", ret)
	}
	defer procRegCloseKey.Call(regKey)

	valuePtr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return err
	}
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}

	valueBytes := (*byte)(unsafe.Pointer(valuePtr))
	valueLen := uint32((len(value) + 1) * 2)

	ret, _, _ = procRegSetValueExW.Call(
		regKey,
		uintptr(unsafe.Pointer(namePtr)),
		0,
		REG_EXPAND_SZ,
		uintptr(unsafe.Pointer(valueBytes)),
		uintptr(valueLen),
	)
	if ret != 0 {
		return fmt.Errorf("set registry value: error code %d", ret)
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
