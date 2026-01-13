/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"

	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/google/go-github/v80/github"
	"github.com/x-dvr/gm/progress"
	"golang.org/x/mod/semver"
)

var (
	ErrNoBuildInfo          = errors.New("build info is not available")
	ErrPlatformNotSupported = errors.New("platform not supported")
	ErrUnsupportedArchive   = errors.New("unsupported archive format")
	ErrUnknownSize          = errors.New("unknown download size")
	ErrChecksumMismatch     = errors.New("checksum verification failed")
	ErrChecksumNotFound     = errors.New("checksum file not found")
)

type Release struct {
	Version   string
	Assets    []Asset
	Checksums string
}

func (r *Release) GetChecksum(assetName string) (string, error) {
	if r.Checksums == "" {
		return "", ErrChecksumNotFound
	}

	lines := strings.Split(r.Checksums, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if parts[1] == assetName {
			return parts[0], nil
		}
	}

	return "", ErrChecksumNotFound
}

func (r *Release) FindAsset(os, arch string) (*Asset, error) {
	idx := slices.IndexFunc(r.Assets, func(a Asset) bool {
		if !strings.HasSuffix(a.Name, ".tar.gz") && !strings.HasSuffix(a.Name, ".zip") {
			return false
		}
		name := strings.TrimSuffix(strings.TrimSuffix(a.Name, ".tar.gz"), ".zip")
		platform := strings.Split(name, "_")[1]
		pparts := strings.Split(platform, ".")
		return os == pparts[0] && arch == pparts[1]
	})
	if idx == -1 {
		return nil, ErrPlatformNotSupported
	}

	return &r.Assets[idx], nil
}

type Asset struct {
	Name string
	URL  string
}

func (a *Asset) Download(tracker progress.IOTracker, expectedChecksum string) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), "gm-up-*."+a.Name)
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}
	defer f.Close()

	tracker.Reset(fmt.Sprintf("Downloading %s ...", a.URL))
	res, err := http.Get(a.URL)
	if err != nil {
		return "", fmt.Errorf("fetch asset: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("wrong HTTP response: %s", res.Status)
	}
	if res.ContentLength == 0 {
		return "", ErrUnknownSize
	}

	tracker.SetSize(res.ContentLength)
	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher, tracker.Writer())

	if _, err := io.Copy(writer, res.Body); err != nil {
		return "", fmt.Errorf("download asset: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		os.Remove(f.Name())
		return "", fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expectedChecksum, actualChecksum)
	}

	return f.Name(), nil
}

func GetUpdate(ctx context.Context) (*Release, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, ErrNoBuildInfo
	}
	module := strings.TrimPrefix(info.Path, "github.com/")
	parts := strings.Split(module, "/")

	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(ctx, parts[0], parts[1], &github.ListOptions{
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, nil
	}

	newVersion := releases[0].GetTagName()
	if semver.Compare(info.Main.Version, newVersion) >= 0 {
		return nil, nil
	}

	return prepare(releases[0]), nil
}

func Extract(src, dest string, tracker progress.IOTracker) error {
	switch {
	case strings.HasSuffix(src, ".zip"):
		return extractZip(src, dest, tracker)
	case strings.HasSuffix(src, ".tar.gz"):
		return extractTarGz(src, dest, tracker)
	default:
		return ErrUnsupportedArchive
	}
}

func prepare(ghr *github.RepositoryRelease) *Release {
	r := Release{
		Version: ghr.GetTagName(),
	}
	for _, a := range ghr.Assets {
		r.Assets = append(r.Assets, Asset{
			Name: a.GetName(),
			URL:  a.GetBrowserDownloadURL(),
		})
		if aName := a.GetName(); strings.HasSuffix(aName, "_checksums.txt") {
			checksums, err := fetchChecksums(a.GetBrowserDownloadURL())
			if err == nil {
				r.Checksums = checksums
			}
		}
	}
	return &r
}

func fetchChecksums(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch checksums: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("wrong HTTP response: %s", res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}

	return string(data), nil
}
