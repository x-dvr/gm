/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package toolchain

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nopTracker satisfies progress.IOTracker with no-op behavior for tests.
type nopTracker struct{}

func (nopTracker) Reset(string)      {}
func (nopTracker) SetSize(int64)     {}
func (nopTracker) Writer() io.Writer { return io.Discard }

func TestSlurpURLToString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			fmt.Fprint(w, "  hello\n")
		case "/notfound":
			http.Error(w, "nope", http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	got, err := slurpURLToString(srv.URL + "/ok")
	if err != nil {
		t.Fatalf("slurpURLToString: %v", err)
	}
	if got != "hello" {
		t.Errorf("slurpURLToString = %q, want %q", got, "hello")
	}

	if _, err := slurpURLToString(srv.URL + "/notfound"); err == nil {
		t.Error("expected error for 404, got nil")
	}
}

func TestDownloadFromURL(t *testing.T) {
	payload := []byte("the-payload-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	dst := filepath.Join(t.TempDir(), "out.bin")
	if err := downloadFromURL(dst, srv.URL+"/file", nopTracker{}); err != nil {
		t.Fatalf("downloadFromURL: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("downloaded bytes = %q, want %q", got, payload)
	}
}

func TestDownloadFromURL_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	dst := filepath.Join(t.TempDir(), "out.bin")
	err := downloadFromURL(dst, srv.URL, nopTracker{})
	if err == nil {
		t.Fatal("downloadFromURL: want error, got nil")
	}
	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Errorf("dst file should be cleaned up on error, stat err = %v", statErr)
	}
}

func TestUnpackArchive_TarGz(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "go.tar.gz")

	files := map[string]string{
		"go/bin/go":      "binary-data",
		"go/src/main.go": "package main",
		"go/VERSION":     "go1.22.0",
	}
	if err := writeTarGz(archive, files); err != nil {
		t.Fatalf("write tar.gz: %v", err)
	}

	out := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := unpackArchive(out, archive, nopTracker{}); err != nil {
		t.Fatalf("unpackArchive: %v", err)
	}

	for name, want := range files {
		rel := strings.TrimPrefix(name, "go/")
		got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("read %s: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("file %s contents = %q, want %q", rel, got, want)
		}
	}
}

func TestUnpackArchive_Zip(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "go.zip")

	files := map[string]string{
		"go/bin/go.exe": "windows-binary",
		"go/README":     "hello",
	}
	if err := writeZip(archive, files); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	out := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := unpackArchive(out, archive, nopTracker{}); err != nil {
		t.Fatalf("unpackArchive: %v", err)
	}

	for name, want := range files {
		rel := strings.TrimPrefix(name, "go/")
		got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("read %s: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("file %s contents = %q, want %q", rel, got, want)
		}
	}
}

func TestUnpackArchive_UnsupportedExtension(t *testing.T) {
	err := unpackArchive(t.TempDir(), "foo.rar", nopTracker{})
	if err == nil {
		t.Error("unpackArchive: want error for unsupported extension, got nil")
	}
}

func TestUnpackArchive_TarGzRejectsInvalidPath(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bad.tar.gz")
	if err := writeTarGz(archive, map[string]string{
		"../escape": "evil",
	}); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	out := filepath.Join(dir, "out")
	os.MkdirAll(out, 0755)
	if err := unpackArchive(out, archive, nopTracker{}); err == nil {
		t.Error("unpackArchive: want error for entry with parent traversal, got nil")
	}
}

// writeTarGz creates a tar.gz archive at path containing the given regular
// files (path -> contents). Intermediate directory entries are also written.
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
			Name: name,
			Mode: 0644,
			Size: int64(len(contents)),
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
