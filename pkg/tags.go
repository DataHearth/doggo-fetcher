package pkg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	var tags []*github.RepositoryTag
	var response *github.Response
	var err error

	tags, response, err = t.client.Repositories.ListTags(t.ctx, "golang", "go", nil)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", ErrBadResponse
	}
	if len(tags) == 0 {
		return "", ErrEmptyTags
	}

	if t.release == "lts" {
		return *tags[len(tags)-1].Name, nil
	}

	for i, tag := range tags {
		userRelease := fmt.Sprintf("go%s", t.release)
		if strings.Contains(*tag.Name, userRelease) {
			if (!beta && strings.Contains(*tag.Name, "beta")) || (!rc && strings.Contains(*tag.Name, "rc")) {
				continue
			}

			if strings.Contains(*tags[i+1].Name, userRelease) {
				continue
			}

			return "", nil
		}
	}

	return "", nil
}
