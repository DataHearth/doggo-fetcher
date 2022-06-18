package pkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ExtractRelease(releasePath, releaseName string) (string, error) {
	f, err := os.Open(releasePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()

	dst, _ := filepath.Split(releasePath)
	tarReader := tar.NewReader(gzipReader)

	dstFolder := ""
	for {
		h, err := tarReader.Next()

		if err != nil {
			if err == io.EOF {
				extractedRelease := fmt.Sprintf("%s/%s", filepath.Dir(dstFolder), releaseName)
				if err := os.Rename(dstFolder, extractedRelease); err != nil {
					return "", err
				}

				return extractedRelease, os.Remove(releasePath)
			}

			return "", err
		}

		target := filepath.Join(dst, h.Name)

		if dstFolder == "" {
			dstFolder = target
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, h.FileInfo().Mode()); err != nil {
				return "", err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, h.FileInfo().Mode())
			if err != nil {
				return "", err
			}
			defer file.Close()

			if _, err = io.Copy(file, tarReader); err != nil {
				return "", err
			}
		}
	}
}
