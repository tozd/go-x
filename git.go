package x

import (
	"strings"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/go-git/go-git/v5"
	"gitlab.com/tozd/go/errors"
)

var (
	ErrOpenGitRepository = errors.Base("cannot open git repository")
	ErrObtainGitRemote   = errors.Base("cannot obtain git remote")
	ErrParseGitRemoteURL = errors.Base("cannot parse git remote URL")
)

// InferGitLabProjectID infers a GitLab project ID from "origin" remote of a
// git repository at path.
func InferGitLabProjectID(path string) (string, errors.E) {
	repository, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		errE := errors.WrapWith(err, ErrOpenGitRepository)
		errors.Details(errE)["path"] = path
		return "", errE
	}

	remote, err := repository.Remote("origin")
	if err != nil {
		errE := errors.WrapWith(err, ErrObtainGitRemote)
		errors.Details(errE)["path"] = path
		errors.Details(errE)["remote"] = "origin"
		return "", errE
	}

	url, err := giturls.Parse(remote.Config().URLs[0])
	if err != nil {
		errE := errors.WrapWith(err, ErrParseGitRemoteURL)
		errors.Details(errE)["path"] = path
		errors.Details(errE)["remote"] = "origin"
		errors.Details(errE)["url"] = remote.Config().URLs[0]
		return "", errE
	}

	url.Path = strings.TrimSuffix(url.Path, ".git")
	url.Path = strings.TrimPrefix(url.Path, "/")

	return url.Path, nil
}
