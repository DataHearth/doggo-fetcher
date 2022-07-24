package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v45/github"
)

type TagsAction interface {
	GetRelease(beta, rc bool) (string, error)
	getTagsRef() ([]*github.Reference, error)
	getLatestTag(beta, rc bool) (string, error)
}

type Tags struct {
	release string
	client  github.Client
	ctx     context.Context
}

func NewTags(release string, ctx context.Context) TagsAction {
	return Tags{
		release: release,
		client:  *github.NewClient(nil),
		ctx:     ctx,
	}
}

// GetRelease retrieves tags from "golang/go" and check whether
// the given release exists in it
//
// Returns the found release or an error
func (t Tags) GetRelease(beta, rc bool) (string, error) {
	if t.release == LTS {
		return t.getLatestTag(beta, rc)
	}

	refs, err := t.getTagsRef()
	if err != nil {
		if err == ErrEmptyTags {
			return "", nil
		}
		return "", err
	}
	for i, ref := range refs {
		userRelease := fmt.Sprintf("go%s", t.release)
		tag := strings.Split(*ref.Ref, "/")[2]

		if strings.Contains(tag, userRelease) {
			if (!beta && strings.Contains(tag, "beta")) || (!rc && strings.Contains(tag, "rc")) {
				continue
			}

			if beta && strings.Contains(tag, "beta") {
				if strings.Contains(*refs[i+1].Ref, userRelease) && strings.Contains(*refs[i+1].Ref, "beta") {
					continue
				} else {
					return tag, nil
				}
			}

			if rc && strings.Contains(tag, "rc") {
				if strings.Contains(*refs[i+1].Ref, userRelease) && strings.Contains(*refs[i+1].Ref, "rc") {
					continue
				} else {
					return tag, nil
				}
			}

			if strings.Contains(*refs[i+1].Ref, userRelease) {
				continue
			}

			return tag, nil
		}
	}

	return "", nil
}

// getTagsRef retrieves all tags from golang/go
//
// Returns a list of tags reference if there is as least one
// or an error otherwise.
func (t Tags) getTagsRef() ([]*github.Reference, error) {
	refs, response, err := t.client.Git.ListMatchingRefs(t.ctx, "golang", "go", &github.ReferenceListOptions{
		Ref: "tags/go",
	})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "%s: %d\n", ErrBadResponse.Error(), response.StatusCode)
		return nil, ErrBadResponse
	}
	if len(refs) == 0 {
		return nil, ErrEmptyTags
	}

	return refs, nil
}

// getLatestTag gather find the latest version of golang.
// beta and rc version can be specified.
//
// Returns the latest release or an error
func (t Tags) getLatestTag(beta, rc bool) (string, error) {
	refs, err := t.getTagsRef()
	if err != nil {
		if err == ErrEmptyTags {
			return "", nil
		}
		return "", err
	}

	for i := len(refs) - 1; i >= 0; i-- {
		if (!beta && strings.Contains(*refs[i].Ref, "beta")) || (!rc && strings.Contains(*refs[i].Ref, "rc")) {
			continue
		}

		if beta && strings.Contains(*refs[i].Ref, "beta") {
			return strings.Split(*refs[i].Ref, "/")[2], nil
		}

		if rc && strings.Contains(*refs[i].Ref, "rc") {
			return strings.Split(*refs[i].Ref, "/")[2], nil
		}

		if !strings.Contains(*refs[i].Ref, "beta") && !strings.Contains(*refs[i].Ref, "rc") {
			return strings.Split(*refs[i].Ref, "/")[2], nil
		}
	}

	return "", nil
}
