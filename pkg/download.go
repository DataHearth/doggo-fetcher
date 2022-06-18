package pkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
)

const GO_DL_SERVER = "https://dl.google.com/go"
const DGF_LOCAL_FOLDER = ".local/doggo-fetcher"

var (
	ErrNonSuccessStatusCode = errors.New("golang download server return a non success status code")
)

type DownloadAction interface {
	DownloadRelease() (string, error)
}

type Download struct {
	release string
	keep    bool
	client  *http.Client
}

func NewDownload(release string, keep bool) DownloadAction {
	return Download{
		release: release,
		keep:    keep,
		client:  http.DefaultClient,
	}
}

func (d Download) DownloadRelease() (string, error) {
	var dstDir string
	osName := runtime.GOOS

	if d.keep {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dstDir = fmt.Sprintf("%s/%s", homeDir, DGF_LOCAL_FOLDER)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return "", err
		}
	} else {
		dstDir = os.TempDir()
	}

	releasePath := fmt.Sprintf("%s/%s.%s-amd64.tar.gz", dstDir, d.release, osName)
	if d.keep {
		fi, err := os.Stat(releasePath)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("error type: %T\n", err)
			return "", err
		} else if os.IsNotExist(err) {
			return releasePath, d.writeRelease(releasePath)
		}

		if fi.IsDir() {
			if err := os.RemoveAll(releasePath); err != nil {
				return "", err
			}

			return releasePath, d.writeRelease(releasePath)
		}

		rsp, err := d.client.Head(fmt.Sprintf("%s/%s.%s-amd64.tar.gz", GO_DL_SERVER, d.release, osName))
		if err != nil {
			return "", err
		}
		if rsp.StatusCode != 200 {
			fmt.Printf("rsp.Request.URL: %v\n", rsp.Request.URL)
			fmt.Fprintf(os.Stderr, "%s: %d\n", ErrNonSuccessStatusCode.Error(), rsp.StatusCode)
			return "", ErrNonSuccessStatusCode
		}

		if rsp.Header.Get("content-length") != "" {
			contentLength, err := strconv.ParseInt(rsp.Header.Get("content-length"), 10, 64)
			if err != nil {
				return "", err
			}

			if contentLength == 0 || contentLength != fi.Size() {
				return releasePath, d.writeRelease(releasePath)
			}

			fmt.Printf("release \"%s\" already downloaded and has a valid size. Skipping...\n", d.release)
			return releasePath, nil
		}

		return releasePath, d.writeRelease(releasePath)
	}

	return releasePath, d.writeRelease(releasePath)
}

func (d Download) writeRelease(releasePath string) error {
	f, err := os.Create(releasePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return d.downloadFile(f)
}

func (d Download) downloadFile(f *os.File) error {
	osName := runtime.GOOS
	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.FinalMSG = "Release downloaded!\n"
	s.Suffix = " Downloading release..."
	s.Start()
	defer s.Stop()

	rsp, err := d.client.Get(fmt.Sprintf("%s/%s.%s-amd64.tar.gz", GO_DL_SERVER, d.release, osName))
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		fmt.Printf("rsp.Request.URL: %v\n", rsp.Request.URL)
		fmt.Fprintf(os.Stderr, "%s: %d\n", ErrNonSuccessStatusCode.Error(), rsp.StatusCode)
		return ErrNonSuccessStatusCode
	}

	if _, err := io.Copy(f, rsp.Body); err != nil {
		return err
	}

	return nil
}
