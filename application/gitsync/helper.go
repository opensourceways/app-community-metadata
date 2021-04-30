package gitsync

import (
	"net/url"
	"strings"
)

func GetRepoLocalName(repo *GitMeta) (string, error) {
	repoUrl, err := url.Parse(repo.Repo)
	if err != nil {
		return "", err
	}
	names := strings.Split(repoUrl.Path, "/")
	return strings.TrimRight(names[len(names)-1], ".git"), nil
}

func RepoEqual(base, compare string) (bool, error) {
	bUrl, err := url.Parse(base)
	if err != nil {
		return false, err
	}
	cUrl, err := url.Parse(compare)
	if err != nil {
		return false, err
	}
	if bUrl.Scheme == cUrl.Scheme && bUrl.Host == cUrl.Scheme &&
			strings.TrimRight(bUrl.Path, "/") == strings.TrimRight(cUrl.Path, "/") {
		return true, nil
	}
	return false, nil
}