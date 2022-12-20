//
// Copyright 2022 Red Hat, Inc.
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
	gitpkg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"net/http"
	"net/url"
	"os"
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
	Protocol string
	Host     string
	Owner    string
	Repo     string
	Branch   string
	Path     string
	token    string
	IsFile   bool
}

// ParseGitUrl extracts information from a GitHub, GitLab, or Bitbucket url
// A client is used to check whether the url is private or public, and sets
// the providers personal access token from the environment variable
func ParseGitUrl(fullUrl string) (GitUrl, error) {
	var c = http.Client{
		Timeout: HTTPRequestResponseTimeout,
	}
	return parseGitUrlWithClient(fullUrl, c)
}

func parseGitUrlWithClient(fullUrl string, c http.Client) (GitUrl, error) {
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
		g, err = parseGitHubUrl(g, parsedUrl, c)
	} else if parsedUrl.Host == GitLabHost {
		g, err = parseGitLabUrl(g, parsedUrl, c)
	} else if parsedUrl.Host == BitbucketHost {
		g, err = parseBitbucketUrl(g, parsedUrl, c)
	} else {
		err = fmt.Errorf("url host should be a valid GitHub, GitLab, or Bitbucket host; received: %s", parsedUrl.Host)
	}

	return g, err
}

func parseGitHubUrl(g GitUrl, url *url.URL, c http.Client) (GitUrl, error) {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host

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

	if !isGitUrlPublic(g, c) {
		g.token = os.Getenv(GitHubToken)
	}

	return g, err
}

func parseGitLabUrl(g GitUrl, url *url.URL, c http.Client) (GitUrl, error) {
	var splitFile, splitOrg []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

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

	if !isGitUrlPublic(g, c) {
		g.token = os.Getenv(GitLabToken)
	}

	return g, err
}

func parseBitbucketUrl(g GitUrl, url *url.URL, c http.Client) (GitUrl, error) {
	var splitUrl []string
	var err error

	g.Protocol = url.Scheme
	g.Host = url.Host
	g.IsFile = false

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

	if !isGitUrlPublic(g, c) {
		g.token = os.Getenv(BitbucketToken)
	}

	return g, err
}

func isGitUrlPublic(g GitUrl, c http.Client) bool {
	host := g.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	repo := fmt.Sprintf("%s://%s/%s/%s", g.Protocol, host, g.Owner, g.Repo)

	if res, err := c.Get(repo); err != nil {
		return false
	} else if res.StatusCode == http.StatusOK {
		return true
	}
	return false
}

// CloneGitRepo clones a GitHub Repo to a destination directory
func CloneGitRepo(g GitUrl, destDir string) error {
	var cloneOptions *gitpkg.CloneOptions

	host := g.Host
	if host == RawGitHubHost {
		host = GitHubHost
	}

	repoUrl := fmt.Sprintf("%s://%s/%s/%s.git", g.Protocol, host, g.Owner, g.Repo)
	branch := fmt.Sprintf("refs/heads/%s", g.Branch)

	cloneOptions = &gitpkg.CloneOptions{
		URL:           repoUrl,
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
	}

	if g.token != "" {
		cloneOptions.Auth = &githttp.BasicAuth{
			// go-git auth allows username to be anything except
			// an empty string for GitHub and GitLab, however requires
			// for Bitbucket to be "x-token-auth"
			Username: "x-token-auth",
			Password: g.token,
		}
	}

	_, err := gitpkg.PlainClone(destDir, false, cloneOptions)
	if err != nil {
		return err
	}

	return nil
}
