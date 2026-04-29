/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type nopTracker struct{}

func (nopTracker) Reset(string)      {}
func (nopTracker) SetSize(int64)     {}
func (nopTracker) Writer() io.Writer { return io.Discard }

func TestRelease_GetChecksum(t *testing.T) {
	r := &Release{
		Checksums: `
abc123  gm_Linux_x86_64.tar.gz
def456  gm_Darwin_arm64.tar.gz

ffff    gm_Windows_x86_64.zip
`,
	}

	got, err := r.GetChecksum("gm_Linux_x86_64.tar.gz")
	if err != nil {
		t.Fatalf("GetChecksum: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want abc123", got)
	}

	got, err = r.GetChecksum("gm_Windows_x86_64.zip")
	if err != nil {
		t.Fatalf("GetChecksum: %v", err)
	}
	if got != "ffff" {
		t.Errorf("got %q, want ffff", got)
	}

	if _, err := r.GetChecksum("gm_Unknown.tar.gz"); !errors.Is(err, ErrChecksumNotFound) {
		t.Errorf("missing asset: err = %v, want ErrChecksumNotFound", err)
	}
}

func TestRelease_GetChecksum_EmptyChecksums(t *testing.T) {
	r := &Release{}
	if _, err := r.GetChecksum("anything"); !errors.Is(err, ErrChecksumNotFound) {
		t.Errorf("err = %v, want ErrChecksumNotFound", err)
	}
}

func TestRelease_FindAsset(t *testing.T) {
	r := &Release{
		Assets: []Asset{
			{Name: "gm_Linux.amd64.tar.gz", URL: "https://example.com/linux.tar.gz"},
			{Name: "gm_Darwin.arm64.tar.gz", URL: "https://example.com/darwin.tar.gz"},
			{Name: "gm_Windows.amd64.zip", URL: "https://example.com/win.zip"},
			{Name: "checksums.txt", URL: "https://example.com/checksums.txt"},
		},
	}

	a, err := r.FindAsset("Linux", "amd64")
	if err != nil {
		t.Fatalf("FindAsset: %v", err)
	}
	if a.Name != "gm_Linux.amd64.tar.gz" {
		t.Errorf("got %q, want gm_Linux.amd64.tar.gz", a.Name)
	}

	a, err = r.FindAsset("Windows", "amd64")
	if err != nil {
		t.Fatalf("FindAsset: %v", err)
	}
	if a.Name != "gm_Windows.amd64.zip" {
		t.Errorf("got %q, want gm_Windows.amd64.zip", a.Name)
	}

	if _, err := r.FindAsset("Plan9", "mips"); !errors.Is(err, ErrPlatformNotSupported) {
		t.Errorf("FindAsset for unknown platform: err = %v, want ErrPlatformNotSupported", err)
	}
}

func TestAsset_Download(t *testing.T) {
	payload := []byte("release-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	sum := sha256.Sum256(payload)
	want := hex.EncodeToString(sum[:])

	a := &Asset{Name: "gm.tar.gz", URL: srv.URL + "/gm.tar.gz"}
	_, err := a.Download(nopTracker{}, "wrongchecksum")
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("Download: err = %v, want ErrChecksumMismatch", err)
	}

	path, err := a.Download(nopTracker{}, want)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("downloaded = %q, want %q", got, payload)
	}
}

func TestAsset_Download_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	a := &Asset{Name: "gm.tar.gz", URL: srv.URL + "/gm.tar.gz"}
	if _, err := a.Download(nopTracker{}, "anything"); err == nil {
		t.Error("Download: want error on HTTP 500, got nil")
	}
}

func TestAsset_Download_UnknownSize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	a := &Asset{Name: "gm.tar.gz", URL: srv.URL + "/gm.tar.gz"}
	if _, err := a.Download(nopTracker{}, "anything"); !errors.Is(err, ErrUnknownSize) {
		t.Errorf("err = %v, want ErrUnknownSize", err)
	}
}

func TestFetchChecksums(t *testing.T) {
	body := "abc  file1\ndef  file2\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)

	got, err := fetchChecksums(srv.URL + "/checksums.txt")
	if err != nil {
		t.Fatalf("fetchChecksums: %v", err)
	}
	if got != body {
		t.Errorf("got %q, want %q", got, body)
	}
}

func TestFetchChecksums_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	if _, err := fetchChecksums(srv.URL + "/checksums.txt"); err == nil {
		t.Error("fetchChecksums: want error, got nil")
	} else if !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v, want it to mention 404", err)
	}
}
