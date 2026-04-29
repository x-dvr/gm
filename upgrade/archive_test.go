/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract_UnsupportedFormat(t *testing.T) {
	err := Extract("foo.7z", t.TempDir(), nopTracker{})
	if !errors.Is(err, ErrUnsupportedArchive) {
		t.Errorf("err = %v, want ErrUnsupportedArchive", err)
	}
}

func TestExtract_TarGz(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "release.tar.gz")
	files := map[string]string{
		"gm":             "binary",
		"docs/README.md": "hello",
	}
	if err := writeTarGz(src, files); err != nil {
		t.Fatalf("makeTarGz: %v", err)
	}

	dst := filepath.Join(dir, "out")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := Extract(src, dst, nopTracker{}); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("file %s = %q, want %q", name, got, want)
		}
	}
}

func TestExtract_Zip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "release.zip")
	files := map[string]string{
		"gm.exe":         "binary",
		"docs/README.md": "hello",
	}
	if err := writeZip(src, files); err != nil {
		t.Fatalf("makeZip: %v", err)
	}

	dst := filepath.Join(dir, "out")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := Extract(src, dst, nopTracker{}); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("read %s: %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("file %s = %q, want %q", name, got, want)
		}
	}
}

// TODO: extract to separate test helper
func writeTarGz(path string, files map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gzw := gzip.NewWriter(f)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for name, contents := range files {
		hdr := &tar.Header{
			Name:     name,
			Mode:     0644,
			Size:     int64(len(contents)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(contents)); err != nil {
			return err
		}
	}
	return nil
}

// TODO: extract to separate test helper
func writeZip(path string, files map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	for name, contents := range files {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(contents)); err != nil {
			return err
		}
	}
	return nil
}
