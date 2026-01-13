/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/x-dvr/gm/progress"
)

func extractZip(src, dest string, tracker progress.IOTracker) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		tracker.Reset(fmt.Sprintf("Extracting %s ...", f.Name))
		fpath := filepath.Join(dest, f.Name)
		fi := f.FileInfo()

		if fi.IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		tracker.SetSize(fi.Size())
		writer := io.MultiWriter(outFile, tracker.Writer())
		_, err = io.Copy(writer, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(src, dest string, tracker progress.IOTracker) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		tracker.Reset(fmt.Sprintf("Extracting %s ...", header.Name))
		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			fi := header.FileInfo()
			mode := fi.Mode()
			os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}

			tracker.SetSize(fi.Size())
			writer := io.MultiWriter(outFile, tracker.Writer())
			if _, err := io.Copy(writer, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
