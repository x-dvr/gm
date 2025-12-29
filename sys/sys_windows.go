//go:build windows

/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	// HWND_BROADCAST sends a message to all top-level windows in the system
	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendmessage#hwnd_broadcast
	HWND_BROADCAST = uintptr(0xffff)
	// WM_SETTINGCHANGE notifies applications that a system parameter has changed
	// https://learn.microsoft.com/en-us/windows/win32/winmsg/wm-settingchange
	WM_SETTINGCHANGE = 0x001A
	// SMTO_ABORTIFHUNG returns immediately if the receiving thread is hung
	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendmessagetimeoutw#smto_abortifhung
	SMTO_ABORTIFHUNG = 0x0002
)

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

// setUserEnv sets a user-level environment variable in the Windows registry
func setUserEnv(name, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		return fmt.Errorf("set registry value: %w", err)
	}

	return nil
}

// broadcastSettingChange notifies the system that environment variables have changed
func broadcastSettingChange() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	// SendMessageTimeoutW sends a message with a timeout to prevent indefinite blocking
	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendmessagetimeoutw
	procSendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	// "Environment" indicates that environment variables have changed
	// https://learn.microsoft.com/en-us/windows/win32/winmsg/wm-settingchange
	environment, _ := windows.UTF16PtrFromString("Environment")
	procSendMessageTimeout.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(environment)),
		SMTO_ABORTIFHUNG,
		5000,
		0,
	)
}

func createSymlink(target, link string) error {
	// os.Symlink(target, link) on Windows requires admin privileges
	return createJunctionFallback(target, link)
}

// createJunctionFallback creates a directory junction using cmd.exe mklink
func createJunctionFallback(target, link string) error {
	cmd := exec.Command("cmd.exe", "/C", "mklink", "/J", link, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create junction: %w (output: %s)", err, string(output))
	}
	return nil
}
