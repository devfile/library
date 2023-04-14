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

package parser

import (
	"github.com/devfile/library/v2/pkg/git"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPopulateFromBytes(t *testing.T) {
	failedToConvertYamlErr := "failed to convert devfile yaml to json: yaml: mapping values are not allowed in this context"

	tests := []struct {
		name        string
		dataFunc    func() []byte
		expectError *string
	}{
		{
			name:     "valid data passed",
			dataFunc: validJsonRawContent200,
		},
		{
			name:        "invalid data passed",
			dataFunc:    invalidJsonRawContent200,
			expectError: &failedToConvertYamlErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(tt.dataFunc())
				if err != nil {
					t.Error(err)
				}
			}))
			var (
				d = DevfileCtx{
					url: testServer.URL,
				}
			)
			defer testServer.Close()
			err := d.PopulateFromURL()
			if (tt.expectError != nil) != (err != nil) {
				t.Errorf("TestPopulateFromBytes(): unexpected error: %v, wantErr: %v", err, tt.expectError)
			} else if tt.expectError != nil {
				assert.Regexp(t, *tt.expectError, err.Error(), "TestPopulateFromBytes(): Error message should match")
			}
		})
	}
}

func TestPopulateFromInvalidURL(t *testing.T) {
	expectError := ".*invalid URI for request"
	t.Run("Populate from invalid URL", func(t *testing.T) {
		var (
			d = DevfileCtx{
				url: "blah",
			}
		)

		err := d.PopulateFromURL()

		if err == nil {
			t.Errorf("TestPopulateFromInvalidURL(): expected an error, didn't get one")
		} else {
			assert.Regexp(t, expectError, err.Error(), "TestPopulateFromInvalidURL(): Error message should match")
		}
	})
}

func TestNewURLDevfileCtx(t *testing.T) {
	var (
		token = "fake-token"
		url   = "https://github.com/devfile/registry/blob/main/stacks/go/2.0.0/devfile.yaml"
	)

	{
		d := NewPrivateURLDevfileCtx(url, token)
		assert.Equal(t, "https://github.com/devfile/registry/blob/main/stacks/go/2.0.0/devfile.yaml", d.GetURL())
		assert.Equal(t, "fake-token", d.GetToken())
		assert.Equal(t, &git.Url{}, d.GetGit())
	}
	{
		d := NewURLDevfileCtx(url)
		assert.Equal(t, "https://github.com/devfile/registry/blob/main/stacks/go/2.0.0/devfile.yaml", d.GetURL())
		assert.Equal(t, "", d.GetToken())
		assert.Equal(t, &git.Url{}, d.GetGit())
	}
}

func invalidJsonRawContent200() []byte {
	return []byte(InvalidDevfileContent)
}
