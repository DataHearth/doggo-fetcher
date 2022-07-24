package pkg

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

type ReleaseActions interface {
	DownloadRelease() error
	ExtractRelease() error
	CheckReleaseExists() error
	GetReleaseFolder() string
	downloadFile(*os.File) error
}

type Release struct {
	release       string
	releaseFolder string
	client        *http.Client
	archivePath   string
}

// NewRelease returns a Release object
func NewRelease(release, releaseFolder string) ReleaseActions {
	return &Release{
		release:       release,
		releaseFolder: releaseFolder,
		client:        http.DefaultClient,
	}
}

// DownloadRelease downloads the golang release
func (d *Release) DownloadRelease() error {
	d.archivePath = fmt.Sprintf("%s/%s.tar.gz", os.TempDir(), d.release)
	f, err := os.Create(d.archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return d.downloadFile(f)
}

// downloadFile downloads the file to the given writer
func (d *Release) downloadFile(f *os.File) error {
	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.Suffix = " Downloading release..."
	s.Start()
	defer s.Stop()

	rsp, err := d.client.Get(fmt.Sprintf("%s/%s.%s-amd64.tar.gz", GO_DL_SERVER, d.release, runtime.GOOS))
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return errors.New("golang download server return a non success status code")
	}

	if _, err := io.Copy(f, rsp.Body); err != nil {
		return err
	}

	return nil
}

// ExtractRelease extracts the golang release
func (d *Release) ExtractRelease() error {
	f, err := os.Open(d.archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	if err := os.MkdirAll(d.releaseFolder, 0755); err != nil {
		return err
	}

	var rootFolder string
	init := true
	for {
		h, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return os.Remove(d.archivePath)
			}

			return err
		}

		// skip first folder
		if init {
			rootFolder = h.Name
			init = false
			continue
		}

		target := filepath.Join(d.releaseFolder, strings.Replace(h.Name, rootFolder, "", 1))
		switch h.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, h.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, h.FileInfo().Mode())
			if err != nil {
				return err
			}

			if _, err = io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}

			file.Close()
		}
	}
}

// CheckReleaseExists checks if the release exists in doggofetcher folder
func (d *Release) CheckReleaseExists() error {
	fi, err := os.Stat(d.releaseFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrReleaseNotFound
		}

		return err
	}

	if !fi.IsDir() {
		return ErrReleaseNotFound
	}

	return nil
}

// GetReleaseFolder returns the release folder
func (d *Release) GetReleaseFolder() string {
	return d.releaseFolder
}
