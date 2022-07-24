package pkg

import "errors"

const GO_DL_SERVER = "https://go.dev/dl"
const DGF_FOLDER = "~/.local/doggofetcher"
const HASHES_FILE = "hashes.txt"
const LTS = "lts"

var (
	ErrReleaseNotFound = errors.New("release not found")
	ErrBadResponse     = errors.New("github API responde with a non success code")
	ErrEmptyTags       = errors.New("no tags found")
	ErrHashNotFound    = errors.New("release hash not found")
	ErrHashInvalid     = errors.New("release hash doesn't match")
)
