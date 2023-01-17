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

package util

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitHubHost    string = "github.com"
	RawGitHubHost string = "raw.githubusercontent.com"
	GitLabHost    string = "gitlab.com"
	BitbucketHost string = "bitbucket.org"

	GitHubToken    string = "GITHUB_TOKEN"
	GitLabToken    string = "GITLAB_TOKEN"
	BitbucketToken string = "BITBUCKET_TOKEN"
)

type GitUrl struct {
	Protocol string // URL scheme
	Host     string // URL domain name
	Owner    string // name of the repo owner
	Repo     string // name of the repo
	Branch   string // branch name
	Path     string // path to a directory or file in the repo
	token    string // used for authenticating a private repo
	IsFile   bool   // defines if the URL points to a file in the repo
}

// ParseGitUrl extracts information from a GitHub, GitLab, or Bitbucket url
func ParseGitUrl(fullUrl string) (GitUrl, error) {
	var g GitUrl

	err := ValidateURL(fullUrl)
	if err != nil {
		return g, err
	}

	parsedUrl, err := url.Parse(fullUrl)
	if err != nil {
		return g, err
	}

	if len(parsedUrl.Path) == 0 {
		return g, fmt.Errorf("url path should not be empty")
	}

	if parsedUrl.Host == RawGitHubHost || parsedUrl.Host == GitHubHost {
		err = g.parseGitHubUrl(parsedUrl)
	} else if parsedUrl.Host == GitLabHost {
		err = g.parseGitLabUrl(parsedUrl)
	} else if parsedUrl.Host == BitbucketHost {
		err = g.parseBitbucketUrl(parsedUrl)
	} else {
		err = fmt.Errorf("url host should be a valid GitHub, GitLab, or Bitbucket host; received: %s", parsedUrl.Host)
	}

	return g, err
}

func (g *GitUrl) parseGitHubUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.token = os.Getenv(GitHubToken)

	if g.Host == RawGitHubHost {
		g.IsFile = true
		// raw GitHub urls don't contain "blob" or "tree"
		splitUrl = strings.SplitN(url.Path[1:], "/", 4)
		if len(splitUrl) == 4 {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]
			g.Branch = splitUrl[2]
			g.Path = splitUrl[3]
		} else {
			err = fmt.Errorf("raw url path should contain <owner>/<repo>/<branch>/<path/to/file>, received: %s", url.Path[1:])
		}
	}

	if g.Host == GitHubHost {
		splitUrl = strings.SplitN(url.Path[1:], "/", 5)
		if len(splitUrl) < 2 {
			err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
		} else {
			g.Owner = splitUrl[0]
			g.Repo = splitUrl[1]

			if len(splitUrl) == 5 {
				switch splitUrl[2] {
				case "tree":
					g.IsFile = false
				case "blob":
					g.IsFile = true
				}
				g.Branch = splitUrl[3]
				g.Path = splitUrl[4]
			}
		}
	}

	return err
}

func (g *GitUrl) parseGitLabUrl(url *url.URL) error {
	var splitFile, splitOrg []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false
	g.token = os.Getenv(GitLabToken)

	// GitLab urls contain a '-' separating the root of the repo
	// and the path to a file or directory
	split := strings.Split(url.Path[1:], "/-/")

	splitOrg = strings.SplitN(split[0], "/", 2)
	if len(split) == 2 {
		splitFile = strings.SplitN(split[1], "/", 3)
	}

	if len(splitOrg) < 2 {
		err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else {
		g.Owner = splitOrg[0]
		g.Repo = splitOrg[1]
	}

	if len(splitFile) == 3 {
		if splitFile[0] == "blob" || splitFile[0] == "tree" || splitFile[0] == "raw" {
			g.Branch = splitFile[1]
			g.Path = splitFile[2]
			ext := filepath.Ext(g.Path)
			if ext != "" {
				g.IsFile = true
			}
		} else {
			err = fmt.Errorf("url path should contain 'blob' or 'tree' or 'raw', received: %s", url.Path[1:])
		}
	}

	return err
}

func (g *GitUrl) parseBitbucketUrl(url *url.URL) error {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false
	g.token = os.Getenv(BitbucketToken)

	splitUrl = strings.SplitN(url.Path[1:], "/", 5)
	if len(splitUrl) < 2 {
		err = fmt.Errorf("url path should contain <user>/<repo>, received: %s", url.Path[1:])
	} else if len(splitUrl) == 2 {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
	} else {
		g.Owner = splitUrl[0]
		g.Repo = splitUrl[1]
		if len(splitUrl) == 5 {
			if splitUrl[2] == "raw" || splitUrl[2] == "src" {
				g.Branch = splitUrl[3]
				g.Path = splitUrl[4]
				ext := filepath.Ext(g.Path)
				if ext != "" {
					g.IsFile = true
				}
			} else {
				err = fmt.Errorf("url path should contain 'raw' or 'src', received: %s", url.Path[1:])
			}
		} else {
			err = fmt.Errorf("url path should contain path to directory or file, received: %s", url.Path[1:])
		}
	}

	return err
}

// ValidateToken makes a http get request to the repo with the GitUrl token
// Returns an error if the get request fails
// If token is empty or invalid and validate succeeds, the repository is public
func (g *GitUrl) ValidateToken(params HTTPRequestParams) error {
	var apiUrl string

	switch g.Host {
	case GitHubHost, RawGitHubHost:
		apiUrl = fmt.Sprintf("https://api.github.com/repos/%s/%s", g.Owner, g.Repo)
	case GitLabHost:
		apiUrl = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s%%2F%s", g.Owner, g.Repo)
	case BitbucketHost:
		apiUrl = fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", g.Owner, g.Repo)
	default:
		apiUrl = fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, g.Host, g.Owner, g.Repo)
	}

	params.URL = apiUrl
	res, err := HTTPGetRequest(params, 0)
	if len(res) == 0 || err != nil {
		return err
	}

	return nil
}

// CloneGitRepo clones a git repo to a destination directory
// Only supports git repositories hosted on GitHub, GitLab, and Bitbucket
func CloneGitRepo(g GitUrl, destDir string, httpTimeout *int) error {
	exist := CheckPathExists(destDir)
	if !exist {
		return fmt.Errorf("failed to clone repo, destination directory: '%s' does not exists", destDir)
	}

	host := g.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	repoUrl := fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, host, g.Owner, g.Repo)

	params := HTTPRequestParams{
		Timeout: httpTimeout,
	}

	// public repos will succeed even if token is invalid or empty
	err := g.ValidateToken(params)

	if err != nil {
		if g.token != "" {
			params.Token = g.token
			err := g.ValidateToken(params)
			if err != nil {
				return fmt.Errorf("failed to validate git url with token, ensure that the url and token is correct. error: %v", err)
			} else {
				repoUrl = fmt.Sprintf("%s://token:%s@%s/%s/%s.git", g.Protocol, g.token, host, g.Owner, g.Repo)
				if g.Host == BitbucketHost {
					repoUrl = fmt.Sprintf("%s://x-token-auth:%s@%s/%s/%s.git", g.Protocol, g.token, host, g.Owner, g.Repo)
				}
			}
		} else {
			return fmt.Errorf("failed to validate git url without a token, ensure that a token is set if the repo is private. error: %v", err)
		}
	}

	/* #nosec G204 -- user input is processed into an expected format for the git clone command */
	c := exec.Command("git", "clone", repoUrl, destDir)
	c.Dir = destDir

	// set env to skip authentication prompt and directly error out
	c.Env = os.Environ()
	c.Env = append(c.Env, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=/bin/echo")

	_, err = c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repo, ensure that the git url is correct. error: %v", err)
	}

	return nil
}
