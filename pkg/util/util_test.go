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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// copied from https://github.com/redhat-developer/odo/blob/main/pkg/util/util_test.go#L1618
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
