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
	"fmt"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"testing"
)

func TestNamespaceOpenShiftObject(t *testing.T) {
	tests := []struct {
		testName        string
		componentName   string
		applicationName string
		want            string
		wantErr         bool
	}{
		{
			testName:        "Test namespacing",
			componentName:   "foo",
			applicationName: "bar",
			want:            "foo-bar",
		},
		{
			testName:        "Blank applicationName with namespacing",
			componentName:   "foo",
			applicationName: "",
			wantErr:         true,
		},
		{
			testName:        "Blank componentName with namespacing",
			componentName:   "",
			applicationName: "bar",
			wantErr:         true,
		},
	}

	// Test that it "joins"
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := NamespaceOpenShiftObject(tt.componentName, tt.applicationName)

			if tt.wantErr && err == nil {
				t.Errorf("Expected an error, got success")
			} else if tt.wantErr == false && err != nil {
				t.Errorf("Error with namespacing: %s", err)
			}

			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}
}

func TestExtractComponentType(t *testing.T) {
	tests := []struct {
		testName      string
		componentType string
		want          string
		wantErr       bool
	}{
		{
			testName:      "Test namespacing and versioning",
			componentType: "myproject/foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test versioning",
			componentType: "foo:3.5",
			want:          "foo",
		},
		{
			testName:      "Test plain component type",
			componentType: "foo",
			want:          "foo",
		},
	}

	// Test that it "joins"
	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := ExtractComponentType(tt.componentType)
			if tt.want != name {
				t.Errorf("Expected %s, got %s", tt.want, name)
			}
		})
	}
}

func TestGetRandomName(t *testing.T) {
	type args struct {
		prefix    string
		existList []string
	}
	tests := []struct {
		testName string
		args     args
		// want is regexp if expectConflictResolution is true else it is a full name
		want string
	}{
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is not already used",
			args: args{
				prefix: "odo",
				existList: []string{
					"odo-auth",
					"odo-pqrs",
				},
			},
			want: "odo-[a-z]{4}",
		},
		{
			testName: "Case: Optional suffix passed and prefix-suffix as a name is already used",
			args: args{
				prefix: "nodejs-ex-nodejs",
				existList: []string{
					"nodejs-ex-nodejs-yvrp",
					"nodejs-ex-nodejs-abcd",
				},
			},
			want: "nodejs-ex-nodejs-[a-z]{4}",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name, err := GetRandomName(tt.args.prefix, -1, tt.args.existList, 3)
			if err != nil {
				t.Errorf("failed to generate a random name. Error %v", err)
			}

			r, _ := regexp.Compile(tt.want)
			match := r.MatchString(name)
			if !match {
				t.Errorf("Received name %s which does not match %s", name, tt.want)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		testName  string
		strLength int
	}{
		{
			testName:  "Case: Generate random string of length 4",
			strLength: 4,
		},
		{
			testName:  "Case: Generate random string of length 3",
			strLength: 3,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			name := GenerateRandomString(tt.strLength)
			r, _ := regexp.Compile(fmt.Sprintf("[a-z]{%d}", tt.strLength))
			match := r.MatchString(name)
			if !match {
				t.Errorf("Randomly generated string %s which does not match regexp %s", name, fmt.Sprintf("[a-z]{%d}", tt.strLength))
			}
		})
	}
}

func TestGetAbsPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		absPath string
		wantErr bool
	}{
		{
			name:    "Case 1: Valid abs path resolution of `~`",
			path:    "~",
			wantErr: false,
		},
		{
			name:    "Case 2: Valid abs path resolution of `.`",
			path:    ".",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Log("Running test: ", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			switch tt.path {
			case "~":
				if len(customHomeDir) > 0 {
					tt.absPath = customHomeDir
				} else {
					usr, err := user.Current()
					if err != nil {
						t.Errorf("Failed to get absolute path corresponding to `~`. Error %v", err)
						return
					}
					tt.absPath = usr.HomeDir
				}

			case ".":
				absPath, err := os.Getwd()
				if err != nil {
					t.Errorf("Failed to get absolute path corresponding to `.`. Error %v", err)
					return
				}
				tt.absPath = absPath
			}
			result, err := GetAbsPath(tt.path)
			if result != tt.absPath {
				t.Errorf("Expected %v, got %v", tt.absPath, result)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("Expected error: %v got error %v", tt.wantErr, err)
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

func TestGetHostWithPort(t *testing.T) {

	tests := []struct {
		inputURL string
		want     string
		wantErr  bool
	}{
		{
			inputURL: "https://example.com",
			want:     "example.com:443",
			wantErr:  false,
		},
		{
			inputURL: "https://example.com:8443",
			want:     "example.com:8443",
			wantErr:  false,
		},
		{
			inputURL: "http://example.com",
			want:     "example.com:80",
			wantErr:  false,
		},
		{
			inputURL: "notexisting://example.com",
			want:     "",
			wantErr:  true,
		},
		{
			inputURL: "http://127.0.0.1",
			want:     "127.0.0.1:80",
			wantErr:  false,
		},
		{
			inputURL: "example.com:1234",
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Testing inputURL: %s", tt.inputURL), func(t *testing.T) {
			got, err := GetHostWithPort(tt.inputURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostWithPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostWithPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAbsGlobExps(t *testing.T) {
	tests := []struct {
		testName              string
		directoryName         string
		inputRelativeGlobExps []string
		expectedGlobExps      []string
	}{
		{
			testName:      "test case 1: with a filename",
			directoryName: "/home/redhat/nodejs-ex/",
			inputRelativeGlobExps: []string{
				"example.txt",
			},
			expectedGlobExps: []string{
				"/home/redhat/nodejs-ex/example.txt",
			},
		},
		{
			testName:      "test case 2: with a folder name",
			directoryName: "/home/redhat/nodejs-ex/",
			inputRelativeGlobExps: []string{
				"example/",
			},
			expectedGlobExps: []string{
				"/home/redhat/nodejs-ex/example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			resultExps := GetAbsGlobExps(tt.directoryName, tt.inputRelativeGlobExps)
			if runtime.GOOS == "windows" {
				for index, element := range resultExps {
					resultExps[index] = filepath.ToSlash(element)
				}
			}

			if !reflect.DeepEqual(resultExps, tt.expectedGlobExps) {
				t.Errorf("expected %v, got %v", tt.expectedGlobExps, resultExps)
			}
		})
	}
}

func TestGetSortedKeys(t *testing.T) {
	tests := []struct {
		testName string
		input    map[string]string
		expected []string
	}{
		{
			testName: "default",
			input:    map[string]string{"a": "av", "c": "cv", "b": "bv"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual := GetSortedKeys(tt.input)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, actual)
			}
		})
	}
}

func TestGetSplitValuesFromStr(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected []string
	}{
		{
			testName: "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			testName: "Single value",
			input:    "s1",
			expected: []string{"s1"},
		},
		{
			testName: "Multiple values",
			input:    "s1, s2, s3 ",
			expected: []string{"s1", "s2", "s3"},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual := GetSplitValuesFromStr(tt.input)
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, actual)
			}
		})
	}
}

func TestGetContainerPortsFromStrings(t *testing.T) {
	tests := []struct {
		name           string
		ports          []string
		containerPorts []corev1.ContainerPort
		wantErr        bool
	}{
		{
			name:  "with normal port values and normal protocol values in lowercase",
			ports: []string{"8080/tcp", "9090/udp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:  "with normal port values and normal protocol values in mixed case",
			ports: []string{"8080/TcP", "9090/uDp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:  "with normal port values and with one protocol value not mentioned",
			ports: []string{"8080", "9090/Udp"},
			containerPorts: []corev1.ContainerPort{
				{
					Name:          "8080-tcp",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "9090-udp",
					ContainerPort: 9090,
					Protocol:      corev1.ProtocolUDP,
				},
			},
			wantErr: false,
		},
		{
			name:    "with normal port values and with one invalid protocol value",
			ports:   []string{"8080/blah", "9090/Udp"},
			wantErr: true,
		},
		{
			name:    "with invalid port values and normal protocol",
			ports:   []string{"ads/Tcp", "9090/Udp"},
			wantErr: true,
		},
		{
			name:    "with invalid port values and one missing protocol value",
			ports:   []string{"ads", "9090/Udp"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := GetContainerPortsFromStrings(tt.ports)
			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(tt.containerPorts, ports) {
					t.Errorf("the ports are not matching, expected %#v, got %#v", tt.containerPorts, ports)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestIsGlobExpMatch(t *testing.T) {

	tests := []struct {
		testName   string
		strToMatch string
		globExps   []string
		want       bool
		wantErr    bool
	}{
		{
			testName:   "Test glob matches",
			strToMatch: "/home/redhat/nodejs-ex/.git",
			globExps:   []string{"/home/redhat/nodejs-ex/.git", "/home/redhat/nodejs-ex/tests/"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test glob does not match",
			strToMatch: "/home/redhat/nodejs-ex/gimmt.gimmt",
			globExps:   []string{"/home/redhat/nodejs-ex/.git/", "/home/redhat/nodejs-ex/tests/"},
			want:       false,
			wantErr:    false,
		},
		{
			testName:   "Test glob match files",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/*.json", "/home/redhat/nodejs-ex/tests/"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test '**' glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/openshift/**/*.json"},
			want:       true,
			wantErr:    false,
		},
		{
			testName:   "Test '!' in glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/!*.json", "/home/redhat/nodejs-ex/tests/"},
			want:       false,
			wantErr:    false,
		},
		{
			testName:   "Test [ in glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"/home/redhat/nodejs-ex/["},
			want:       false,
			wantErr:    true,
		},
		{
			testName:   "Test '#' comment glob matches",
			strToMatch: "/home/redhat/nodejs-ex/openshift/templates/example.json",
			globExps:   []string{"#/home/redhat/nodejs-ex/openshift/**/*.json"},
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			matched, err := IsGlobExpMatch(tt.strToMatch, tt.globExps)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != matched {
				t.Errorf("expected %v, got %v", tt.want, matched)
			}
		})
	}
}

func TestRemoveRelativePathFromFiles(t *testing.T) {
	type args struct {
		path   string
		input  []string
		output []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Case 1 - Remove the relative path from a list of files",
			args: args{
				path:   "/foo/bar",
				input:  []string{"/foo/bar/1", "/foo/bar/2", "/foo/bar/3/", "/foo/bar/4/5/foo/bar"},
				output: []string{"1", "2", "3", "4/5/foo/bar"},
			},
			wantErr: false,
		},
		{
			name: "Case 2 - Fail on purpose with an invalid path",
			args: args{
				path:   `..`,
				input:  []string{"foo", "bar"},
				output: []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Run function RemoveRelativePathFromFiles
			output, err := RemoveRelativePathFromFiles(tt.args.input, tt.args.path)
			if runtime.GOOS == "windows" {
				for index, element := range output {
					output[index] = filepath.ToSlash(element)
				}
			}

			// Check for error
			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !(reflect.DeepEqual(output, tt.args.output)) {
				t.Errorf("expected %v, got %v", tt.args.output, output)
			}

		})
	}
}

func TestHTTPGetFreePort(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "case 1: get a valid free port",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTTPGetFreePort()
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPGetFreePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			addressLook := "localhost:" + strconv.Itoa(got)
			listener, err := net.Listen("tcp", addressLook)
			if err != nil {
				t.Errorf("expected a free port, but listening failed cause: %v", err)
			} else {
				_ = listener.Close()
			}
		})
	}
}

func TestGetRemoteFilesMarkedForDeletion(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		remotePath string
		want       []string
	}{
		{
			name:       "case 1: no files",
			files:      []string{},
			remotePath: "/projects",
			want:       nil,
		},
		{
			name:       "case 2: one file",
			files:      []string{"abc.txt"},
			remotePath: "/projects",
			want:       []string{"/projects/abc.txt"},
		},
		{
			name:       "case 3: multiple files",
			files:      []string{"abc.txt", "def.txt", "hello.txt"},
			remotePath: "/projects",
			want:       []string{"/projects/abc.txt", "/projects/def.txt", "/projects/hello.txt"},
		},
		{
			name:       "case 4: remote path multiple folders",
			files:      []string{"abc.txt", "def.txt", "hello.txt"},
			remotePath: "/test/folder",
			want:       []string{"/test/folder/abc.txt", "/test/folder/def.txt", "/test/folder/hello.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteFiles := GetRemoteFilesMarkedForDeletion(tt.files, tt.remotePath)
			if !reflect.DeepEqual(tt.want, remoteFiles) {
				t.Errorf("Expected %s, got %s", tt.want, remoteFiles)
			}
		})
	}
}

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

func TestFilterIgnores(t *testing.T) {
	tests := []struct {
		name             string
		changedFiles     []string
		deletedFiles     []string
		ignoredFiles     []string
		wantChangedFiles []string
		wantDeletedFiles []string
	}{
		{
			name:             "Case 1: No ignored files",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{},
			wantChangedFiles: []string{"hello.txt", "test.abc"},
			wantDeletedFiles: []string{"one.txt", "two.txt"},
		},
		{
			name:             "Case 2: One ignored file",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{"hello.txt"},
			wantChangedFiles: []string{"test.abc"},
			wantDeletedFiles: []string{"one.txt", "two.txt"},
		},
		{
			name:             "Case 3: Multiple ignored file",
			changedFiles:     []string{"hello.txt", "test.abc"},
			deletedFiles:     []string{"one.txt", "two.txt"},
			ignoredFiles:     []string{"hello.txt", "two.txt"},
			wantChangedFiles: []string{"test.abc"},
			wantDeletedFiles: []string{"one.txt"},
		},
		{
			name:             "Case 4: No changed or deleted files",
			changedFiles:     []string{""},
			deletedFiles:     []string{""},
			ignoredFiles:     []string{"hello.txt", "two.txt"},
			wantChangedFiles: []string{""},
			wantDeletedFiles: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterChanged, filterDeleted := FilterIgnores(tt.changedFiles, tt.deletedFiles, tt.ignoredFiles)

			if !reflect.DeepEqual(tt.wantChangedFiles, filterChanged) {
				t.Errorf("Expected %s, got %s", tt.wantChangedFiles, filterChanged)
			}

			if !reflect.DeepEqual(tt.wantDeletedFiles, filterDeleted) {
				t.Errorf("Expected %s, got %s", tt.wantDeletedFiles, filterDeleted)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
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
		name     string
		url      string
		filepath string
		want     []byte
		wantErr  bool
	}{
		{
			name:     "Case 1: Input url is valid",
			url:      server.URL,
			filepath: "./test.yaml",
			// Want(Expected) result is "OK"
			// According to Unicode table: O == 79, K == 75
			want:    []byte{79, 75},
			wantErr: false,
		},
		{
			name:     "Case 2: Input url is invalid",
			url:      "invalid",
			filepath: "./test.yaml",
			want:     []byte{},
			wantErr:  true,
		},
		{
			name:     "Case 3: Input url is an empty string",
			url:      "",
			filepath: "./test.yaml",
			want:     []byte{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			params := DownloadParams{
				Request: HTTPRequestParams{
					URL: tt.url,
				},
				Filepath: tt.filepath,
			}
			err := DownloadFile(params)
			if err != nil {
				gotErr = true
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Error("Failed to get expected error")
			}

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Failed to download file with error %s", err)
				}

				got, err := ioutil.ReadFile(tt.filepath)
				if err != nil {
					t.Errorf("Failed to read file with error %s", err)
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Got: %v, want: %v", got, tt.want)
				}

				// Clean up the file that downloaded in this test case
				err = os.Remove(tt.filepath)
				if err != nil {
					t.Errorf("Failed to delete file with error %s", err)
				}
			}
		})
	}
}

func TestDownloadInMemory(t *testing.T) {
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
		token   string
		want    []byte
		wantErr string
	}{
		{
			name: "Case 1: Input url is valid",
			url:  server.URL,
			want: []byte{79, 75},
		},
		{
			name:    "Case 2: Input url is invalid",
			url:     "invalid",
			wantErr: "unsupported protocol scheme",
		},
		{
			name:    "Case 3: Git provider with invalid url",
			url:     "github.com/mike-hoang/invalid-repo",
			token:   "",
			want:    []byte(nil),
			wantErr: "failed to parse git repo. error:*",
		},
		{
			name:    "Case 4: Public Github repo with missing blob",
			url:     "https://github.com/devfile/library/main/README.md",
			wantErr: "failed to parse git repo. error: url path to directory or file should contain 'tree' or 'blob'*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := DownloadInMemory(HTTPRequestParams{URL: tt.url, Token: tt.token})
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("Failed to download file with error: %s", err)
			} else if err == nil && !reflect.DeepEqual(data, tt.want) {
				t.Errorf("Expected: %v, received: %v, difference at %v", tt.want, string(data[:]), pretty.Compare(tt.want, data))
			} else if err != nil {
				assert.Regexp(t, tt.wantErr, err.Error(), "Error message should match")
			}
		})
	}
}

func TestValidateK8sResourceName(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  bool
	}{
		{
			name:  "Case 1: Resource name is valid",
			key:   "component name",
			value: "good-name",
			want:  true,
		},
		{
			name:  "Case 2: Resource name contains unsupported character",
			key:   "component name",
			value: "BAD@name",
			want:  false,
		},
		{
			name:  "Case 3: Resource name contains all numeric values",
			key:   "component name",
			value: "12345",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateK8sResourceName(tt.key, tt.value)
			got := err == nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got %t, want %t", got, tt.want)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	// Create temp dir and temp file
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		t.Errorf("Failed to create temp file: %s, error: %v", tempFile.Name(), err)
	}
	defer tempFile.Close()

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Case 1: Valid file path",
			filePath: tempFile.Name(),
			wantErr:  false,
		},
		{
			name:     "Case 2: Invalid file path",
			filePath: "!@#",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			err := ValidateFile(tt.filePath)
			if err != nil {
				gotErr = true
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got error: %t, want error: %t", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp dir
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}

	// Create temp file under temp dir as source file
	tempFile, err := ioutil.TempFile(tempDir, "")
	if err != nil {
		t.Errorf("Failed to create temp file: %s, error: %v", tempFile.Name(), err)
	}
	defer tempFile.Close()

	srcPath := tempFile.Name()
	fakePath := "!@#/**"
	dstPath := filepath.Join(tempDir, "dstFile")
	info, _ := os.Stat(srcPath)

	tests := []struct {
		name    string
		srcPath string
		dstPath string
		wantErr bool
	}{
		{
			name:    "should be able to copy file to destination path",
			srcPath: srcPath,
			dstPath: dstPath,
			wantErr: false,
		},
		{
			name:    "should fail if source path is invalid",
			srcPath: fakePath,
			dstPath: dstPath,
			wantErr: true,
		},
		{
			name:    "should fail if destination path is invalid",
			srcPath: srcPath,
			dstPath: fakePath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			err = CopyFile(tt.srcPath, tt.dstPath, info)
			if err != nil {
				gotErr = true
			}

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got error: %t, want error: %t", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCopyAllDirFiles(t *testing.T) {
	fs := filesystem.NewFakeFs()
	fs.MkdirAll("/path/to/src", 0755)
	fs.MkdirAll("/path/to/dest", 0755)
	fs.WriteFile("/path/to/src/devfile.yaml", []byte(""), 0755)
	fs.WriteFile("/path/to/src/file.txt", []byte(""), 0755)
	fs.WriteFile("/path/to/src/subdir/devfile.yaml", []byte(""), 0755)
	fs.WriteFile("/path/to/src/subdir/file.txt", []byte(""), 0755)

	srcDir := "/path/to/src"
	srcSubDir := "/path/to/src/subdir"
	destDir := "/path/to/dest"
	missingDir := "/invalid/path/to/dir"

	tests := []struct {
		name    string
		srcDir  string
		destDir string
		wantErr bool
	}{
		{
			name:    "should be able to copy files to destination path",
			srcDir:  srcDir,
			destDir: destDir,
			wantErr: false,
		},
		{
			name:    "should be able to copy subdir files to destination path",
			srcDir:  srcSubDir,
			destDir: destDir,
			wantErr: false,
		},
		{
			name:    "should fail if source path is invalid",
			srcDir:  missingDir,
			destDir: destDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := copyAllDirFilesOnFS(tt.srcDir, tt.destDir, fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %t, got error: %t", tt.wantErr, err)
			}
		})
	}
}

func TestPathEqual(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Can't get absolute path of current working directory with error: %v", err)
	}
	fileAbsPath := filepath.Join(currentDir, "file")
	fileRelPath := filepath.Join(".", "file")

	tests := []struct {
		name       string
		firstPath  string
		secondPath string
		want       bool
	}{
		{
			name:       "Case 1: Two paths (two absolute paths) are equal",
			firstPath:  fileAbsPath,
			secondPath: fileAbsPath,
			want:       true,
		},
		{
			name:       "Case 2: Two paths (one absolute path, one relative path) are equal",
			firstPath:  fileAbsPath,
			secondPath: fileRelPath,
			want:       true,
		},
		{
			name:       "Case 3: Two paths are not equal",
			firstPath:  fileAbsPath,
			secondPath: filepath.Join(fileAbsPath, "file"),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathEqual(tt.firstPath, tt.secondPath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %t, want %t", got, tt.want)
			}
		})
	}
}
