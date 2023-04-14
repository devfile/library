//
// Copyright 2023 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

type MockGitUrl struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Branch   string // branch name
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

func (m *MockGitUrl) GetToken() string {
	return m.token
}

var mockExecute = func(baseDir string, cmd CommandType, args ...string) ([]byte, error) {
	if cmd == GitCommand {
		u, _ := url.Parse(args[1])
		password, hasPassword := u.User.Password()

		if hasPassword {
			switch password {
			case "valid-token":
				return []byte("test"), nil
			default:
				return []byte(""), fmt.Errorf("not a valid token")
			}
		}

		return []byte("test"), nil
	}

	return []byte(""), fmt.Errorf(unsupportedCmdMsg, string(cmd))
}

func (m *MockGitUrl) CloneGitRepo(destDir string) error {
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := m.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	var repoUrl string
	if m.GetToken() == "" {
		repoUrl = fmt.Sprintf("%s://%s/%s/%s.git", m.Protocol, host, m.Owner, m.Repo)
	} else {
		repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", m.Protocol, m.GetToken(), host, m.Owner, m.Repo)
		if m.Host == BitbucketHost {
			repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", m.Protocol, m.GetToken(), host, m.Owner, m.Repo)
		}
	}

	_, err := mockExecute(destDir, "git", "clone", repoUrl, ".")

	if err != nil {
		if m.GetToken() == "" {
			return fmt.Errorf("failed to clone repo without a token, ensure that a token is set if the repo is private")
		} else {
			return fmt.Errorf("failed to clone repo with token, ensure that the url and token is correct")
		}
	}

	return nil
}

func (m *MockGitUrl) DownloadGitRepoResources(url string, destDir string, httpTimeout *int, token string) error {
	gitUrl := m
	if gitUrl.IsGitProviderRepo() && gitUrl.IsFile {
		stackDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("git-resources"))
		if err != nil {
			return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
		}
		defer os.RemoveAll(stackDir)

		if !gitUrl.IsPublic(httpTimeout) {
			err = m.SetToken(token, httpTimeout)
			if err != nil {
				return err
			}
		}

		err = gitUrl.CloneGitRepo(stackDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MockGitUrl) SetToken(token string, httpTimeout *int) error {
	m.token = token
	return nil
}

func (m *MockGitUrl) IsPublic(httpTimeout *int) bool {
	if *httpTimeout != 0 {
		return false
	}
	return true
}

func (m *MockGitUrl) GitRawFileAPI() string {
	return ""
}

func (m *MockGitUrl) IsGitProviderRepo() bool {
	return true
}
