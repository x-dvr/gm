/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	gmDir     = ".gm"
	workspace = "workspace"
	versions  = "versions"
	current   = "current"
)

var (
	ErrNoPath       = errors.New("environment variable 'PATH' is not set")
	ErrNotInstalled = errors.New("version is not installed")
)

func PathForVersion(version string) (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir of user: %w", err)
	}
	return filepath.Join(homedir, gmDir, versions, version), nil
}

func SetAsCurrent(version string) error {
	versionPath, err := PathForVersion(version)
	if err != nil {
		return fmt.Errorf("get path for version %q: %w", version, err)
	}
	currentPath, err := PathForVersion(current)
	if err != nil {
		return fmt.Errorf("get path for current version: %w", err)
	}

	if _, err := os.Stat(versionPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNotInstalled
		}
		return fmt.Errorf("check installed version: %w", err)
	}

	if err := os.Remove(currentPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reset current version: %w", err)
	}

	return os.Symlink(versionPath, currentPath)
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

	fmt.Printf("export GOPATH=%s\n", goPath)
	fmt.Printf("export GOBIN=%s\n", goBin)
	fmt.Printf("export GOROOT=%s\n", goRoot)
	fmt.Printf("export PATH=\"%s:$PATH\"\n", goSDKBin+":"+goBin)
	return nil
}
