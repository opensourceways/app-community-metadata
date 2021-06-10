/*
Copyright 2021 The Opensourceways Group.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitsync

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
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

func PathIncludes(s []string, absPath string) bool {
	for _, a := range s {
		if strings.HasSuffix(absPath, a) {
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
