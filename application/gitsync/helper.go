package gitsync

import (
	"encoding/json"
	"net/url"
	"strings"
	"errors"
)

func GetRepoLocalName(repo string) string {
	repoUrl, err := url.Parse(repo)
	if err != nil {
		return ""
	}
	names := strings.Split(repoUrl.Path, "/")
	return strings.TrimRight(names[len(names)-1], ".git")
}

func RepoEqualIgnoreSchemaAndLevel(base, compare string) (bool, error) {
	bUrl, err := url.Parse(base)
	if err != nil {
		return false, err
	}
	cUrl, err := url.Parse(compare)
	if err != nil {
		return false, err
	}
	if bUrl.Host == cUrl.Host {
		bPath := strings.Split(strings.TrimRight(bUrl.Path, "/"), "/")
		cPath := strings.Split(strings.TrimRight(cUrl.Path, "/"), "/")
		if len(bPath) == 0 || len(cPath) == 0 {
			return false, nil
		}
		if bPath[len(bPath)-1] == cPath[len(cPath)-1] {
			return true, nil
		}
	}
	return false, nil
}

func GetRepo(s []GitMeta, name string) *GitMeta {
	for _, a := range s {
		if a.Repo == name {
			return &a
		}
	}
	return nil
}

func StringInclude(s []string, name string) bool {
	for _, a := range s {
		if a == name {
			return true
		}
	}
	return false
}

func DeepCopyMap(src map[string][]string, dst map[string][]string) error {
	if src == nil {
		return errors.New("src cannot be nil")
	}
	if dst == nil {
		return errors.New("dst cannot be nil")
	}
	jsonStr, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonStr, &dst)
	if err != nil {
		return err
	}
	return nil
}