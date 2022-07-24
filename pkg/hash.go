package pkg

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type HashActions interface {
	GetFolderHash(path string) (string, error)
	CompareReleaseHash(path string, hash string) error
	AddHash(path, hash string) error
	ReplaceHash(path, hash string) error
}

type Hash struct {
	hashTable     map[string]string
	hashTablePath string
}

// NewHash returns a new Hash object. It reads the hashes from the ~/.local/doggofetcher/hashes.txt file then loads
// them into the hash table.
func NewHash(localFolder string) (HashActions, error) {
	hashTablePath := filepath.Join(localFolder, HASHES_FILE)
	f, err := os.OpenFile(hashTablePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return &Hash{}, err
	}
	defer f.Close()

	hashTable := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.Split(sc.Text(), " ")
		hashTable[l[0]] = l[1]
	}

	return &Hash{
		hashTable:     hashTable,
		hashTablePath: hashTablePath,
	}, nil
}

// GetFolderHash returns the hash of the folder by using the Merkle tree.
func (h *Hash) GetFolderHash(path string) (string, error) {
	hashes := [][]byte{}
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.Type().IsRegular() {
			return nil
		}
		if d.IsDir() {
			hash, err := h.GetFolderHash(path)
			if err != nil {
				return err
			}

			hashes = append(hashes, []byte(hash))
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		sha := sha256.New()
		buf := make([]byte, 10*1024)
		for {
			n, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			sha.Write(buf[:n])
		}

		hashes = append(hashes, sha.Sum(nil))
		return nil
	})
	if err != nil {
		return "", err
	}

	sha := sha256.New()
	for _, h := range hashes {
		if _, err := sha.Write(h); err != nil {
			return "", err
		}
	}

	return string(sha.Sum(nil)), nil
}

// CompareReleaseHash compares the hash of the release with the hash in the hash table.
func (h *Hash) CompareReleaseHash(path, hash string) error {
	if h, ok := h.hashTable[path]; !ok {
		return ErrHashNotFound
	} else if h != hash {
		return ErrHashInvalid
	} else {
		return nil
	}
}

// AddHash adds a hash to the hash table by writing the "hashTable" property and the file.
func (h *Hash) AddHash(path, hash string) error {
	h.hashTable[path] = hash

	f, err := os.OpenFile(h.hashTablePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s %s\n", path, hash)); err != nil {
		return err
	}

	return nil
}

// ReplaceHash replaces the hash in the hash table with the new hash.
func (h *Hash) ReplaceHash(path, hash string) error {
	h.hashTable[path] = hash

	d, err := os.ReadFile(h.hashTablePath)
	if err != nil {
		return err
	}

	// todo: find a way to write only the relevent line
	var out []byte
	for _, l := range strings.Split(string(d), "\n") {
		if strings.Contains(l, path) {
			out = append(out, []byte(strings.Replace(l, strings.Split(l, " ")[1], hash, 1))...)
		}
	}

	if err := os.WriteFile(h.hashTablePath, out, 0644); err != nil {
		return err
	}

	return nil
}
