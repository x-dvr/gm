/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

type Toolchain struct {
	Version string
	Path    string
}

func PathForVersion(version string) (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir of user: %w", err)
	}
	return filepath.Join(homedir, gmDir, versions, version), nil
}

func ListInstalledVersions() ([]Toolchain, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir of user: %w", err)
	}

	versionsPath := filepath.Join(homedir, gmDir, versions)

	entries, err := os.ReadDir(versionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Toolchain{}, nil
		}
		return nil, fmt.Errorf("read versions directory: %w", err)
	}

	var installed []Toolchain
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "go") {
			installed = append(installed, Toolchain{
				Path:    filepath.Join(versionsPath, entry.Name()),
				Version: strings.TrimPrefix(entry.Name(), "go"),
			})
		}
	}
	return installed, nil
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

	return createSymlink(versionPath, currentPath)
}

func GetCurrentVersion() (*Toolchain, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir of user: %w", err)
	}

	currentPath := filepath.Join(homedir, gmDir, versions, current)
	target, err := os.Readlink(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read current version symlink: %w", err)
	}

	t := Toolchain{
		Path:    target,
		Version: strings.TrimPrefix(filepath.Base(target), "go"),
	}

	return &t, nil
}
