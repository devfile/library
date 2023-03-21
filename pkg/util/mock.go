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

import "net/http"

var (
	GetDoFunc            func(req *http.Request) (*http.Response, error)
	GetParseGitUrlFunc   func(url string) error
	GetGitRawFileAPIFunc func() string
	GetSetTokenFunc      func(token string, httpTimeout *int) error
	GetIsPublicFunc      func(httpTimeout *int) bool
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

type MockGitUrl struct {
	ParseGitUrlFunc      func(fullUrl string) error
	GetGitRawFileAPIFunc func(url string) string
	SetTokenFunc         func(token string, httpTimeout *int) error
	IsPublicFunc         func(httpTimeout *int) bool
}

func (m *MockGitUrl) ParseGitUrl(fullUrl string) error {
	return GetParseGitUrlFunc(fullUrl)
}

func (m *MockGitUrl) GitRawFileAPI() string {
	return GetGitRawFileAPIFunc()
}

func (m *MockGitUrl) SetToken(token string, httpTimeout *int) error {
	return GetSetTokenFunc(token, httpTimeout)
}

func (m *MockGitUrl) IsPublic(httpTimeout *int) bool {
	return GetIsPublicFunc(httpTimeout)
}
