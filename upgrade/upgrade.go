/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/google/go-github/v80/github"
	"golang.org/x/mod/semver"
)

var (
	ErrNoBuildInfo          = errors.New("build info is not available")
	ErrPlatformNotSupported = errors.New("platform not supported")
	ErrUnsupportedArchive   = errors.New("unsupported archive format")
)

type Release struct {
	Version string
	Assets  []Asset
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

func (a *Asset) Download() (string, error) {
	f, err := os.CreateTemp(os.TempDir(), "gm-up-*."+a.Name)
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}
	defer f.Close()

	res, err := http.Get(a.URL)
	if err != nil {
		return "", fmt.Errorf("fetch asset: %w", err)
	}
	defer res.Body.Close()
	// size, err := strconv.Atoi(res.Header.Get("Content-Length"))
	// if err != nil {
	// 	return "", fmt.Errorf("get download size: %w", err)
	// }
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("wrong HTTP response: %s", res.Status)
	}

	if _, err := io.Copy(f, res.Body); err != nil {
		return "", fmt.Errorf("download asset: %w", err)
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

func Extract(src, dest string) error {
	switch {
	case strings.HasSuffix(src, ".zip"):
		return extractZip(src, dest)
	case strings.HasSuffix(src, ".tar.gz"):
		return extractTarGz(src, dest)
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
	}
	return &r
}
