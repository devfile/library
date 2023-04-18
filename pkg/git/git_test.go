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
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_ParseGitUrl(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantUrl Url
		wantErr string
	}{
		{
			name:    "should fail with empty url",
			url:     "",
			wantErr: "URL is invalid",
		},
		{
			name:    "should fail with invalid git host",
			url:     "https://google.ca/",
			wantErr: "url host should be a valid GitHub, GitLab, or Bitbucket host*",
		},
		// GitHub
		{
			name: "should parse GitHub repo with root path",
			url:  "https://github.com/devfile/library",
			wantUrl: Url{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "",
				Path:     "",
				IsFile:   false,
			},
		},
		{
			name: "should parse GitHub repo with root path and tag",
			url:  "https://github.com/devfile/library/tree/v2.2.0",
			wantUrl: Url{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "v2.2.0",
				Path:     "",
				IsFile:   false,
			},
		},
		{
			name: "should parse GitHub repo with root path and revision",
			url:  "https://github.com/devfile/library/tree/0ce592a416fb185564516353891a45016ac7f671",
			wantUrl: Url{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "0ce592a416fb185564516353891a45016ac7f671",
				Path:     "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only GitHub host",
			url:     "https://github.com/",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name: "should parse GitHub repo with file path",
			url:  "https://github.com/devfile/library/blob/main/devfile.yaml",
			wantUrl: Url{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "main",
				Path:     "devfile.yaml",
				IsFile:   true,
			},
		},
		{
			name: "should parse GitHub repo with raw file path",
			url:  "https://raw.githubusercontent.com/devfile/library/main/devfile.yaml",
			wantUrl: Url{
				Protocol: "https",
				Host:     "raw.githubusercontent.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "main",
				Path:     "devfile.yaml",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing GitHub repo",
			url:     "https://github.com/devfile",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with missing GitHub blob",
			url:     "https://github.com/devfile/library/main/devfile.yaml",
			wantErr: "url path to directory or file should contain*",
		},
		{
			name:    "should fail with missing GitHub tree",
			url:     "https://github.com/devfile/library/main/tests/yamls",
			wantErr: "url path to directory or file should contain*",
		},
		{
			name:    "should fail with just GitHub tree",
			url:     "https://github.com/devfile/library/tree",
			wantErr: "url path should contain <owner>/<repo>/<tree or blob>/<branch>/<path/to/file/or/directory>*",
		},
		{
			name:    "should fail with just GitHub blob",
			url:     "https://github.com/devfile/library/blob",
			wantErr: "url path should contain <owner>/<repo>/<tree or blob>/<branch>/<path/to/file/or/directory>*",
		},
		{
			name:    "should fail with invalid GitHub raw file path",
			url:     "https://raw.githubusercontent.com/devfile/library/devfile.yaml",
			wantErr: "raw url path should contain <owner>/<repo>/<branch>/<path/to/file>*",
		},
		// Gitlab
		{
			name: "should parse GitLab repo with root path",
			url:  "https://gitlab.com/gitlab-org/gitlab-foss",
			wantUrl: Url{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "gitlab-org",
				Repo:     "gitlab-foss",
				Revision: "",
				Path:     "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only GitLab host",
			url:     "https://gitlab.com/",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name: "should parse GitLab repo with file path",
			url:  "https://gitlab.com/gitlab-org/gitlab-foss/-/blob/master/README.md",
			wantUrl: Url{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "gitlab-org",
				Repo:     "gitlab-foss",
				Revision: "master",
				Path:     "README.md",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing GitLab repo",
			url:     "https://gitlab.com/gitlab-org",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with missing GitLab keywords",
			url:     "https://gitlab.com/gitlab-org/gitlab-foss/-/master/directory/README.md",
			wantErr: "url path should contain 'blob' or 'tree' or 'raw'*",
		},
		{
			name:    "should fail with missing GitLab file or directory path",
			url:     "https://gitlab.com/gitlab-org/gitlab-foss/-/tree/master",
			wantErr: "url path to directory or file should contain <blob or tree or raw>/<branch>/<path/to/file/or/directory>*",
		},
		// Bitbucket
		{
			name: "should parse Bitbucket repo with root path",
			url:  "https://bitbucket.org/fake-owner/fake-public-repo",
			wantUrl: Url{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Revision: "",
				Path:     "",
				IsFile:   false,
			},
		},
		{
			name:    "should fail with only Bitbucket host",
			url:     "https://bitbucket.org/",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name: "should parse Bitbucket repo with file path",
			url:  "https://bitbucket.org/fake-owner/fake-public-repo/src/main/README.md",
			wantUrl: Url{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Revision: "main",
				Path:     "README.md",
				IsFile:   true,
			},
		},
		{
			name: "should parse Bitbucket file path with nested path",
			url:  "https://bitbucket.org/fake-owner/fake-public-repo/src/main/directory/test.txt",
			wantUrl: Url{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Revision: "main",
				Path:     "directory/test.txt",
				IsFile:   true,
			},
		},
		{
			name: "should parse Bitbucket repo with raw file path",
			url:  "https://bitbucket.org/fake-owner/fake-public-repo/raw/main/README.md",
			wantUrl: Url{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "fake-owner",
				Repo:     "fake-public-repo",
				Revision: "main",
				Path:     "README.md",
				IsFile:   true,
			},
		},
		{
			name:    "should fail with missing Bitbucket repo",
			url:     "https://bitbucket.org/fake-owner",
			wantErr: "url path should contain <user>/<repo>*",
		},
		{
			name:    "should fail with invalid Bitbucket directory or file path",
			url:     "https://bitbucket.org/fake-owner/fake-public-repo/main/README.md",
			wantErr: "url path should contain path to directory or file*",
		},
		{
			name:    "should fail with missing Bitbucket keywords",
			url:     "https://bitbucket.org/fake-owner/fake-public-repo/main/test/README.md",
			wantErr: "url path should contain 'raw' or 'src'*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitUrl(tt.url)
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

func Test_GetGitRawFileAPI(t *testing.T) {
	tests := []struct {
		name string
		g    Url
		want string
	}{
		{
			name: "Github url",
			g: Url{
				Protocol: "https",
				Host:     "github.com",
				Owner:    "devfile",
				Repo:     "library",
				Revision: "main",
				Path:     "tests/README.md",
			},
			want: "https://raw.githubusercontent.com/devfile/library/main/tests/README.md",
		},
		{
			name: "GitLab url",
			g: Url{
				Protocol: "https",
				Host:     "gitlab.com",
				Owner:    "gitlab-org",
				Repo:     "gitlab",
				Revision: "master",
				Path:     "README.md",
			},
			want: "https://gitlab.com/api/v4/projects/gitlab-org%2Fgitlab/repository/files/README.md/raw",
		},
		{
			name: "Bitbucket url",
			g: Url{
				Protocol: "https",
				Host:     "bitbucket.org",
				Owner:    "owner",
				Repo:     "repo-name",
				Revision: "main",
				Path:     "path/to/file.md",
			},
			want: "https://api.bitbucket.org/2.0/repositories/owner/repo-name/src/main/path/to/file.md",
		},
		{
			name: "Empty GitUrl",
			g:    Url{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.g.GitRawFileAPI()
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("Got: %v, want: %v", result, tt.want)
			}
		})
	}
}

func Test_IsPublic(t *testing.T) {
	publicGitUrl := Url{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "devfile",
		Repo:     "library",
		Revision: "main",
		token:    "fake-token",
	}

	privateGitUrl := Url{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "not",
		Repo:     "a-valid",
		Revision: "none",
		token:    "fake-token",
	}

	httpTimeout := 0

	tests := []struct {
		name string
		g    Url
		want bool
	}{
		{
			name: "should be public",
			g:    publicGitUrl,
			want: true,
		},
		{
			name: "should be private",
			g:    privateGitUrl,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.g.IsPublic(&httpTimeout)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("Got: %t, want: %t", result, tt.want)
			}
		})
	}
}

func Test_CloneGitRepo(t *testing.T) {
	tempInvalidDir := t.TempDir()
	tempDirGitHub := t.TempDir()
	tempDirGitLab := t.TempDir()
	tempDirBitbucket := t.TempDir()

	invalidGitUrl := Url{
		Protocol: "",
		Host:     "",
		Owner:    "nonexistent",
		Repo:     "nonexistent",
		Revision: "nonexistent",
	}

	validPublicGitHubUrl := Url{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "devfile",
		Repo:     "library",
		Revision: "main",
	}

	validPublicGitLabUrl := Url{
		Protocol: "https",
		Host:     "gitlab.com",
		Owner:    "mike-hoang",
		Repo:     "public-testing-repo",
		Revision: "main",
	}

	validPublicBitbucketUrl := Url{
		Protocol: "https",
		Host:     "bitbucket.org",
		Owner:    "mike-hoang",
		Repo:     "public-testing-repo",
		Revision: "master",
	}

	invalidPrivateGitHubRepo := Url{
		Protocol: "https",
		Host:     "github.com",
		Owner:    "fake-owner",
		Repo:     "fake-private-repo",
		Revision: "master",
		token:    "fake-github-token",
	}

	privateRepoBadTokenErr := "failed to clone repo with token*"
	publicRepoInvalidUrlErr := "failed to clone repo without a token"
	missingDestDirErr := "failed to clone repo, destination directory*"

	tests := []struct {
		name    string
		gitUrl  Url
		destDir string
		wantErr string
	}{
		{
			name:    "should fail with invalid destination directory",
			gitUrl:  invalidGitUrl,
			destDir: filepath.Join(os.TempDir(), "nonexistent"),
			wantErr: missingDestDirErr,
		},
		{
			name:    "should fail with invalid git url",
			gitUrl:  invalidGitUrl,
			destDir: tempInvalidDir,
			wantErr: publicRepoInvalidUrlErr,
		},
		{
			name:    "should fail to clone invalid private git url with a bad token",
			gitUrl:  invalidPrivateGitHubRepo,
			destDir: tempInvalidDir,
			wantErr: privateRepoBadTokenErr,
		},
		{
			name:    "should be able to clone valid public github url",
			gitUrl:  validPublicGitHubUrl,
			destDir: tempDirGitHub,
		},
		{
			name:    "should be able to clone valid public gitlab url",
			gitUrl:  validPublicGitLabUrl,
			destDir: tempDirGitLab,
		},
		{
			name:    "should be able to clone valid public bitbucket url",
			gitUrl:  validPublicBitbucketUrl,
			destDir: tempDirBitbucket,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gitUrl.CloneGitRepo(tt.destDir)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("Unxpected error: %t, want: %v", err, tt.wantErr)
			} else if err != nil {
				assert.Regexp(t, tt.wantErr, err.Error(), "Error message should match")
			}
		})
	}
}