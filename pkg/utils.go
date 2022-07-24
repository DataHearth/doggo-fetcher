package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UpdateRelease(releaseFolder string) error {
	activeReleaseFolder := filepath.Join(strings.Replace(DGF_FOLDER, "~", os.Getenv("HOME"), 1), "go")
	if err := os.RemoveAll(activeReleaseFolder); err != nil {
		return err
	}
	if err := os.MkdirAll(activeReleaseFolder, 0755); err != nil {
		return err
	}

	return CopyFolder(releaseFolder, activeReleaseFolder, true)
}

func CopyFolder(src, dst string, init bool) error {
	items, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, i := range items {
		info, err := i.Info()
		if err != nil {
			return err
		}

		src := filepath.Join(src, i.Name())
		dst := filepath.Join(dst, i.Name())
		if i.IsDir() {
			if err := os.MkdirAll(dst, info.Mode()); err != nil {
				return err
			}

			CopyFolder(src, dst, false)
			continue
		}

		f, err := os.Open(src)
		if err != nil {
			return err
		}

		dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, info.Mode())
		if err != nil {
			f.Close()
			return err
		}

		if _, err := io.Copy(dstFile, f); err != nil {
			f.Close()
			dstFile.Close()
			return err
		}

		f.Close()
		dstFile.Close()
	}

	return nil
}

func Init() error {
	dgfFolder := strings.Replace(DGF_FOLDER, "~", os.Getenv("HOME"), 1)
	if err := os.MkdirAll(dgfFolder, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(getShellConfigFile(), os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	msg := fmt.Sprintln("\n\n# doggofetcher section")
	msg += fmt.Sprintf("export PATH=%s:$PATH\n", filepath.Join(dgfFolder, "go", "bin"))
	if _, err := f.WriteString(msg); err != nil {
		return err
	}

	return nil
}

func getShellConfigFile() string {
	if strings.Contains(os.Getenv("SHELL"), "zsh") {
		return fmt.Sprintf("%s/.zshrc", os.Getenv("HOME"))
	}

	return fmt.Sprintf("%s/.bashrc", os.Getenv("HOME"))
}
