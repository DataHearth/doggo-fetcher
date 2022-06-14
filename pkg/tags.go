package pkg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v45/github"
)

var (
	ErrBadResponse = errors.New("github API responde with a non success code")
	ErrEmptyTags   = errors.New("no tags found")
)

type Tags struct {
	release string
	client  github.Client
	ctx     context.Context
}

func NewTags(release string, ctx context.Context) Tags {
	return Tags{
		release: release,
		client:  *github.NewClient(nil),
		ctx:     ctx,
	}
}

// checkReleaseExists retrieves tags from "golang/go" and check whether
// the given release exists in it
func (t Tags) CheckReleaseExists(beta, rc bool) (string, error) {
	if t.release == "lts" {
		return t.getLatestRelease(beta, rc)
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

// getTags retrieves 100 tags from parameter PAGE
//
// It returns a list of tags reference if there is as least one tag in the result and an error otherwise
func (t Tags) getTagsRef() ([]*github.Reference, error) {
	refs, response, err := t.client.Git.ListMatchingRefs(t.ctx, "golang", "go", &github.ReferenceListOptions{
		Ref: "tags/go",
	})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "non success github status code %d\n", response.StatusCode)
		return nil, ErrBadResponse
	}
	if len(refs) == 0 {
		return nil, ErrEmptyTags
	}

	return refs, nil
}

func (t Tags) getLatestRelease(beta, rc bool) (string, error) {
	refs, err := t.getTagsRef()
	if err != nil {
		if err == ErrEmptyTags {
			return "", nil
		}
		return "", err
	}

	fmt.Printf("len(tags): %v\n", len(refs))
	for i := len(refs) - 1; i >= 0; i-- {
		fmt.Printf("[%d] tags[i].Name: %v\n", i, *refs[i].Ref)
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
