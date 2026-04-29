/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// setHome overrides the user's home directory for the duration of the test.
// os.UserHomeDir consults HOME on unix and USERPROFILE on Windows.
func setHome(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}
}

func TestPathForVersion(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	got, err := PathForVersion("go1.22.0")
	if err != nil {
		t.Fatalf("PathForVersion: %v", err)
	}
	want := filepath.Join(home, gmDir, versions, "go1.22.0")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestListInstalledVersions_NoDirectory(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	got, err := ListInstalledVersions()
	if err != nil {
		t.Fatalf("ListInstalledVersions: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want empty slice when versions dir does not exist", got)
	}
}

func TestListInstalledVersions(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	versionsDir := filepath.Join(home, gmDir, versions)
	for _, d := range []string{"go1.21.0", "go1.22.0", "current", "not-a-version"} {
		if err := os.MkdirAll(filepath.Join(versionsDir, d), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	// A file (not a directory) should be ignored even if it has the prefix.
	if err := os.WriteFile(filepath.Join(versionsDir, "go-not-dir"), nil, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := ListInstalledVersions()
	if err != nil {
		t.Fatalf("ListInstalledVersions: %v", err)
	}

	wantVersions := map[string]bool{"1.21.0": false, "1.22.0": false}
	for _, tc := range got {
		if _, ok := wantVersions[tc.Version]; ok {
			wantVersions[tc.Version] = true
		} else if tc.Version != "" {
			// "current" trims to "" because it has no "go" prefix and was filtered out.
			t.Errorf("unexpected version %q in result", tc.Version)
		}
		expectedPath := filepath.Join(versionsDir, "go"+tc.Version)
		if tc.Path != expectedPath {
			t.Errorf("path for %q = %q, want %q", tc.Version, tc.Path, expectedPath)
		}
	}
	for v, seen := range wantVersions {
		if !seen {
			t.Errorf("version %q missing from result", v)
		}
	}
}

func TestGetCurrentVersion_NoSymlink(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	got, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion: %v", err)
	}
	if got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestSetAsCurrentAndGetCurrent(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	versionsDir := filepath.Join(home, gmDir, versions)
	if err := os.MkdirAll(filepath.Join(versionsDir, "go1.22.0"), 0755); err != nil {
		t.Fatalf("mkdir version: %v", err)
	}

	// No current version yet.
	tc, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion (none): %v", err)
	}
	if tc != nil {
		t.Errorf("got %+v, want nil before SetAsCurrent", tc)
	}

	if err := SetAsCurrent("go1.22.0"); err != nil {
		t.Fatalf("SetAsCurrent: %v", err)
	}

	tc, err = GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion: %v", err)
	}
	if tc == nil {
		t.Fatal("GetCurrentVersion returned nil after SetAsCurrent")
	}
	if tc.Version != "1.22.0" {
		t.Errorf("Version = %q, want %q", tc.Version, "1.22.0")
	}
	if tc.Path != filepath.Join(versionsDir, "go1.22.0") {
		t.Errorf("Path = %q, want %q", tc.Path, filepath.Join(versionsDir, "go1.22.0"))
	}

	// Switching to another installed version should overwrite the symlink.
	if err := os.MkdirAll(filepath.Join(versionsDir, "go1.21.0"), 0755); err != nil {
		t.Fatalf("mkdir version: %v", err)
	}
	if err := SetAsCurrent("go1.21.0"); err != nil {
		t.Fatalf("SetAsCurrent (switch): %v", err)
	}
	tc, err = GetCurrentVersion()
	if err != nil {
		t.Fatalf("GetCurrentVersion after switch: %v", err)
	}
	if tc == nil || tc.Version != "1.21.0" {
		t.Errorf("after switch, Version = %v, want 1.21.0", tc)
	}
}

func TestSetAsCurrent_NotInstalled(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	if err := os.MkdirAll(filepath.Join(home, gmDir, versions), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	err := SetAsCurrent("go9.9.9")
	if !errors.Is(err, ErrNotInstalled) {
		t.Errorf("err = %v, want ErrNotInstalled", err)
	}
}
