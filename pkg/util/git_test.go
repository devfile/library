//
// Copyright 2021-2022 Red Hat, Inc.
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

package util

import (
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	githubToken    = "fake-github-token"
	gitlabToken    = "fake-gitlab-token"
	bitbucketToken = "fake-bitbucket-token"
)

type respondWithStatus struct {
	status int
}

func (rs respondWithStatus) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: rs.status,
	}, nil
}

var (
	publicClient  = http.Client{Transport: respondWithStatus{status: http.StatusOK}}
	privateClient = http.Client{Transport: respondWithStatus{status: http.StatusNotFound}}
)

func Test_parseGitUrlWithClient(t *testing.T) {
	defer func() {
		err := os.Unsetenv(githubToken)
		if err != nil {
			t.Errorf("Failed to unset GitHub token")
		}
		err = os.Unsetenv(gitlabToken)
		if err != nil {
			t.Errorf("Failed to unset GitLab token")
		}
		err = os.Unsetenv(bitbucketToken)
		if err != nil {
			t.Errorf("Failed to unset Bitbucket token")
		}
	}()

	err := os.Setenv("GITHUB_TOKEN", githubToken)
	if err != nil {
		t.Errorf("Failed to set GitHub token")
	}
	err = os.Setenv("GITLAB_TOKEN", gitlabToken)
	if err != nil {
		t.Errorf("Failed to set GitLab token")
	}
	err = os.Setenv("BITBUCKET_TOKEN", bitbucketToken)
	if err != nil {
		t.Errorf("Failed to set Bitbucket token")
	}

	tests := []struct {
		name    string
		url     string
		client  http.Client
		wantUrl GitUrl
		wantErr string
	}{
		{
			name:    "should fail with empty url",
			url:     "",
			client:  publicClient,
			wantErr: "URL is invalid",
		},
		{
			name:    "should fail with invalid git host",
			url:     "https://google.ca/",
			client:  publicClient,
			wantErr: "url host should be a valid GitHub, GitLab, or Bitbucket host*",
		},
		// GitHub
		{
			name:   "should parse public GitHub repo with root path",
			url:    "https://github.com/devfile/library",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Branch:   "",
				Path:     "",
				token:    "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only GitHub host",
			url:     "https://github.com/",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:   "should parse public GitHub repo with file path",
			url:    "https://github.com/devfile/library/blob/main/devfile.yaml",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Branch:   "main",
				Path:     "devfile.yaml",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:   "should parse public GitHub repo with raw file path",
			url:    "https://raw.githubusercontent.com/devfile/library/main/devfile.yaml",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "raw.githubusercontent.com",
				Owner:    "devfile",
				Repo:     "library",
				Branch:   "main",
				Path:     "devfile.yaml",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing GitHub repo",
			url:     "https://github.com/devfile",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with invalid GitHub raw file path",
			url:     "https://raw.githubusercontent.com/devfile/library/devfile.yaml",
			client:  publicClient,
			wantErr: "raw url path should contain <owner>/<repo>/<branch>/<path/to/file>*",
		},
		{
			name:   "should parse private GitHub repo with token",
			url:    "https://github.com/fake-owner/fake-private-repo",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "",
				Path:     "",
				token:    "fake-github-token",
				IsFile:   false,
			},
		},
		{
			name:   "should parse private raw GitHub file path with token",
			url:    "https://raw.githubusercontent.com/fake-owner/fake-private-repo/main/README.md",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "raw.githubusercontent.com",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "main",
				Path:     "README.md",
				token:    "fake-github-token",
				IsFile:   true,
			},
		},
		// Gitlab
		{
			name:   "should parse public GitLab repo with root path",
			url:    "https://gitlab.com/gitlab-org/gitlab-foss",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "gitlab-org",
				Repo:     "gitlab-foss",
				Branch:   "",
				Path:     "",
				token:    "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only GitLab host",
			url:     "https://gitlab.com/",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:   "should parse public GitLab repo with file path",
			url:    "https://gitlab.com/gitlab-org/gitlab-foss/-/blob/master/README.md",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "gitlab-org",
				Repo:     "gitlab-foss",
				Branch:   "master",
				Path:     "README.md",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing GitLab repo",
			url:     "https://gitlab.com/gitlab-org",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with missing GitLab keywords",
			url:     "https://gitlab.com/gitlab-org/gitlab-foss/-/master/directory/README.md",
			client:  publicClient,
			wantErr: "url path should contain 'blob' or 'tree' or 'raw'*",
		},
		{
			name:   "should parse private GitLab repo with token",
			url:    "https://gitlab.com/fake-owner/fake-private-repo",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "",
				Path:     "",
				token:    "fake-gitlab-token",
				IsFile:   false,
			},
		},
		{
			name:   "should parse private raw GitLab file path with token",
			url:    "https://gitlab.com/fake-owner/fake-private-repo/-/raw/main/README.md",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "main",
				Path:     "README.md",
				token:    "fake-gitlab-token",
				IsFile:   true,
			},
		},
		// Bitbucket
		{
			name:   "should parse public Bitbucket repo with root path",
			url:    "https://bitbucket.org/fake-owner/fake-public-repo",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Branch:   "",
				Path:     "",
				token:    "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only Bitbucket host",
			url:     "https://bitbucket.org/",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:   "should parse public Bitbucket repo with file path",
			url:    "https://bitbucket.org/fake-owner/fake-public-repo/src/main/README.md",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Branch:   "main",
				Path:     "README.md",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:   "should parse public Bitbucket file path with nested path",
			url:    "https://bitbucket.org/fake-owner/fake-public-repo/src/main/directory/test.txt",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Branch:   "main",
				Path:     "directory/test.txt",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:   "should parse public Bitbucket repo with raw file path",
			url:    "https://bitbucket.org/fake-owner/fake-public-repo/raw/main/README.md",
			client: publicClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Branch:   "main",
				Path:     "README.md",
				token:    "",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing Bitbucket repo",
			url:     "https://bitbucket.org/fake-owner",
			client:  publicClient,
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with invalid Bitbucket directory or file path",
			url:     "https://bitbucket.org/fake-owner/fake-public-repo/main/README.md",
			client:  publicClient,
			wantErr: "url path should contain path to directory or file*",
		},
		{
			name:    "should fail with missing Bitbucket keywords",
			url:     "https://bitbucket.org/fake-owner/fake-public-repo/main/test/README.md",
			client:  publicClient,
			wantErr: "url path should contain 'raw' or 'src'*",
		},
		{
			name:   "should parse private Bitbucket repo with token",
			url:    "https://bitbucket.org/fake-owner/fake-private-repo",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "",
				Path:     "",
				token:    "fake-bitbucket-token",
				IsFile:   false,
			},
		},
		{
			name:   "should parse private raw Bitbucket file path with token",
			url:    "https://bitbucket.org/fake-owner/fake-private-repo/raw/main/README.md",
			client: privateClient,
			wantUrl: GitUrl{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-private-repo",
				Branch:   "main",
				Path:     "README.md",
				token:    "fake-bitbucket-token",
				IsFile:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitUrlWithClient(tt.url, tt.client)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("Unxpected error: %t, want: %v", err, tt.wantUrl)
			} else if err == nil && !reflect.DeepEqual(got, tt.wantUrl) {
				t.Errorf("Expected: %v, received: %v, difference at %v", tt.wantUrl, got, pretty.Compare(tt.wantUrl, got))
			} else if err != nil {
				assert.Regexp(t, tt.wantErr, err.Error(), "Error message should match")
			}
		})
	}
}

func TestCloneGitRepo(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	invalidGitUrl := GitUrl{
		Protocol: "",
		Host:     "",
		Owner:    "nonexistent",
		Repo:     "nonexistent",
		Branch:   "nonexistent",
	}

	validGitHubUrl := GitUrl{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "devfile",
		Repo:     "library",
		Branch:   "main",
	}

	tests := []struct {
		name    string
		gitUrl  GitUrl
		destDir string
		wantErr bool
	}{
		{
			name:    "should fail with invalid git url",
			gitUrl:  invalidGitUrl,
			destDir: filepath.Join(os.TempDir(), "nonexistent"),
			wantErr: true,
		},
		{
			name:    "should be able to clone valid github url",
			gitUrl:  validGitHubUrl,
			destDir: tempDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CloneGitRepo(tt.gitUrl, tt.destDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %t, got error: %t", tt.wantErr, err)
			}
		})
	}
}
