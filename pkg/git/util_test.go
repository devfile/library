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

package git

import (
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHTTPGetRequest(t *testing.T) {
	invalidHTTPTimeout := -1
	validHTTPTimeout := 20

	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		want    []byte
		timeout *int
	}{
		{
			name: "Case 1: Input url is valid",
			url:  server.URL,
			// Want(Expected) result is "OK"
			// According to Unicode table: O == 79, K == 75
			want: []byte{79, 75},
		},
		{
			name: "Case 2: Input url is invalid",
			url:  "invalid",
			want: nil,
		},
		{
			name:    "Case 3: Test invalid httpTimeout, default timeout will be used",
			url:     server.URL,
			timeout: &invalidHTTPTimeout,
			want:    []byte{79, 75},
		},
		{
			name:    "Case 4: Test valid httpTimeout",
			url:     server.URL,
			timeout: &validHTTPTimeout,
			want:    []byte{79, 75},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := HTTPRequestParams{
				URL:     tt.url,
				Timeout: tt.timeout,
			}
			got, err := HTTPGetRequest(request, 0)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want: %v", got, tt.want)
				t.Logf("Error message is: %v", err)
			}
		})
	}
}

func TestCheckPathExists(t *testing.T) {
	fs := filesystem.NewFakeFs()
	fs.MkdirAll("/path/to/devfile", 0755)
	fs.WriteFile("/path/to/devfile/devfile.yaml", []byte(""), 0755)

	file := "/path/to/devfile/devfile.yaml"
	missingFile := "/path/to/not/devfile"

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "should be able to get file that exists",
			filePath: file,
			want:     true,
		},
		{
			name:     "should fail if file does not exist",
			filePath: missingFile,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkPathExistsOnFS(tt.filePath, fs)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("Got error: %t, want error: %t", result, tt.want)
			}
		})
	}
}
