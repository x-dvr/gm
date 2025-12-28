// This file is borrowed with minimal modifications from https://github.com/golang/dl/blob/master/internal/version/version.go
// Copyright 2016 The Go Authors. All rights reserved.
// License: BSD-3-Clause (https://raw.githubusercontent.com/golang/dl/refs/heads/master/LICENSE)

package toolchain

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const goDownloadBaseURL = "https://dl.google.com/go"

const installSuccessMarker = ".install-success"

func Install(version, destPath string) error {
	dir, err := os.Stat(filepath.Join(destPath, installSuccessMarker))
	if err == nil && dir.IsDir() {
		slog.Info("Version of Go toolchain is already installed", slog.String("version", version), slog.String("path", destPath))
		return nil
	}

	err = os.MkdirAll(destPath, 0755)
	if err != nil {
		return fmt.Errorf("create destination directory %s: %w", destPath, err)
	}

	goURL := getDownloadURL(version)
	res, err := http.Head(goURL)
	if err != nil {
		return fmt.Errorf("check size of %s: %w", goURL, err)
	}
	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("no binary release of %s for %s/%s at %s", version, runtime.GOOS, runtime.GOARCH, goURL)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s checking size of %s", http.StatusText(res.StatusCode), goURL)
	}

	base := path.Base(goURL)
	archiveFile := filepath.Join(destPath, base)
	if fi, err := os.Stat(archiveFile); err != nil || fi.Size() != res.ContentLength {
		if err != nil && !os.IsNotExist(err) {
			// Something weird. Don't try to download.
			return err
		}
		if err := downloadFromURL(archiveFile, goURL); err != nil {
			return fmt.Errorf("download %s: %w", goURL, err)
		}
		fi, err = os.Stat(archiveFile)
		if err != nil {
			return err
		}
		if fi.Size() != res.ContentLength {
			return fmt.Errorf("downloaded file %s size %d doesn't match server size %d", archiveFile, fi.Size(), res.ContentLength)
		}
	}

	expectedSHA, err := slurpURLToString(goURL + ".sha256")
	if err != nil {
		return err
	}
	if err := verifySHA256(archiveFile, expectedSHA); err != nil {
		return fmt.Errorf("verify SHA256 of %s: %w", archiveFile, err)
	}
	slog.Info(fmt.Sprintf("Unpacking %s ...", archiveFile))
	if err := unpackArchive(destPath, archiveFile); err != nil {
		return fmt.Errorf("extract archive %s: %w", archiveFile, err)
	}
	if err := os.WriteFile(filepath.Join(destPath, installSuccessMarker), nil, 0644); err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Successfully installed Go toolchain version '%s'", version))
	return nil
}

// verifySHA256 reports whether the named file has contents with
// SHA-256 of the given wantHex value.
func verifySHA256(file, wantHex string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}
	if fmt.Sprintf("%x", hash.Sum(nil)) != wantHex {
		return fmt.Errorf("%s corrupt? does not have expected SHA-256 of %s", file, wantHex)
	}
	return nil
}

// unpackArchive unpacks the provided archive zip or tar.gz file to targetDir,
// removing the "go/" prefix from file entries.
func unpackArchive(targetDir, archiveFile string) error {
	switch {
	case strings.HasSuffix(archiveFile, ".zip"):
		return unpackZip(targetDir, archiveFile)
	case strings.HasSuffix(archiveFile, ".tar.gz"):
		return unpackTarGz(targetDir, archiveFile)
	default:
		return errors.New("unsupported archive file")
	}
}

// unpackTarGz is the tar.gz implementation of unpackArchive.
func unpackTarGz(targetDir, archiveFile string) error {
	r, err := os.Open(archiveFile)
	if err != nil {
		return err
	}
	defer r.Close()
	madeDir := map[string]bool{}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tr := tar.NewReader(zr)
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar file contained invalid name %q", f.Name)
		}
		rel := filepath.FromSlash(strings.TrimPrefix(f.Name, "go/"))
		abs := filepath.Join(targetDir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
					return err
				}
				madeDir[dir] = true
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("write to %s: %w", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			if !f.ModTime.IsZero() {
				if err := os.Chtimes(abs, f.ModTime, f.ModTime); err != nil {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					slog.Error("Error changing modtime", slog.String("error", err.Error()))
				}
			}
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
			madeDir[abs] = true
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %d", f.Name, mode)
		}
	}
	return nil
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

// unpackZip is the zip implementation of unpackArchive.
func unpackZip(targetDir, archiveFile string) error {
	zr, err := zip.OpenReader(archiveFile)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		name := strings.TrimPrefix(f.Name, "go/")

		outpath := filepath.Join(targetDir, name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outpath, 0755); err != nil {
				return err
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		// File
		if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		if err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
	}
	return nil
}

func downloadFromURL(dstFile, srcURL string) (err error) {
	f, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(dstFile)
		}
	}()
	c := &http.Client{
		Transport: &userAgentTransport{&http.Transport{
			// It's already compressed. Prefer accurate ContentLength.
			// (Not that GCS would try to compress it, though)
			DisableCompression: true,
			DisableKeepAlives:  true,
			Proxy:              http.ProxyFromEnvironment,
		}},
	}
	res, err := c.Get(srcURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	pw := &progressWriter{w: f, total: res.ContentLength, output: os.Stderr}
	n, err := io.Copy(pw, res.Body)
	if err != nil {
		return err
	}
	if res.ContentLength != -1 && res.ContentLength != n {
		return fmt.Errorf("copied %d bytes; expected %d", n, res.ContentLength)
	}
	pw.update() // 100%
	return f.Close()
}

// slurpURLToString downloads the given URL and returns it as a string.
func slurpURLToString(url_ string) (string, error) {
	res, err := http.Get(url_)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url_, res.Status)
	}
	slurp, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", url_, err)
	}
	return strings.TrimSpace(string(slurp)), nil
}

func getDownloadURL(version string) string {
	goos := runtime.GOOS
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	arch := runtime.GOARCH
	if goos == "linux" && runtime.GOARCH == "arm" {
		arch = "armv6l"
	}
	return fmt.Sprintf("%s/%s.%s-%s%s", goDownloadBaseURL, version, goos, arch, ext)
}

type progressWriter struct {
	w         io.Writer
	n         int64
	total     int64
	last      time.Time
	formatted bool
	output    io.Writer
}

func (p *progressWriter) update() {
	end := " ..."
	if p.n == p.total {
		end = ""
	}
	if p.formatted {
		fmt.Fprintf(p.output, "Downloaded %5.1f%% (%s / %s)%s\n",
			(100.0*float64(p.n))/float64(p.total),
			fmtSize(p.n), fmtSize(p.total), end)
	} else {
		fmt.Fprintf(p.output, "Downloaded %5.1f%% (%*d / %d bytes)%s\n",
			(100.0*float64(p.n))/float64(p.total),
			ndigits(p.total), p.n, p.total, end)
	}
}

func ndigits(i int64) int {
	var n int
	for ; i != 0; i /= 10 {
		n++
	}
	return n
}

func fmtSize(size int64) string {
	const (
		byte_unit = 1 << (10 * iota)
		kilobyte_unit
		megabyte_unit
	)

	unit := "B"
	value := float64(size)

	switch {
	case size >= megabyte_unit:
		unit = "MB"
		value = value / megabyte_unit
	case size >= kilobyte_unit:
		unit = "KB"
		value = value / kilobyte_unit
	}
	formatted := strings.TrimSuffix(strconv.FormatFloat(value, 'f', 1, 64), ".0")
	return fmt.Sprintf("%s %s", formatted, unit)
}

func (p *progressWriter) Write(buf []byte) (n int, err error) {
	n, err = p.w.Write(buf)
	p.n += int64(n)
	if now := time.Now(); now.Unix() != p.last.Unix() {
		p.update()
		p.last = now
	}
	return
}

type userAgentTransport struct {
	rt http.RoundTripper
}

func (uat userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	version := runtime.Version()
	if strings.Contains(version, "devel") {
		version = "devel"
	}
	r.Header.Set("User-Agent", "go-manager/"+version)
	return uat.rt.RoundTrip(r)
}
