//
// Copyright Red Hat, Inc.
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
	"path/filepath"
	"testing"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/google/go-cmp/cmp"
)

func Test_lookupDevfileFromPath(t *testing.T) {
	type fields struct {
		relPath      string
		fsCustomizer func(filesystem.Filesystem) error
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "invalid relative path",
			fields: fields{
				relPath: "non/existent/relative/path",
			},
			wantErr: true,
		},
		{
			name: "invalid absolute path",
			fields: fields{
				relPath: "/non/existent/absolute/path",
			},
			wantErr: true,
		},
		{
			name: "relative path to file",
			fields: fields{
				relPath: "my-devfile.yaml",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					_, err := fs.Create("my-devfile.yaml")
					return err
				},
			},
			wantErr: false,
			want:    "my-devfile.yaml",
		},
		{
			name: "absolute path to file",
			fields: fields{
				relPath: "/my-absolute-devfile.yaml",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					_, err := fs.Create("/my-absolute-devfile.yaml")
					return err
				},
			},
			wantErr: false,
			want:    "/my-absolute-devfile.yaml",
		},
		{
			name: "empty directory",
			fields: fields{
				relPath: "my-files",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					return fs.MkdirAll("my-files", 0755)
				},
			},
			wantErr: true,
		},
		{
			name: "directory with no devfile filename detected",
			fields: fields{
				relPath: "my-files",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					dir := "my-files"
					err := fs.MkdirAll("my-files", 0755)
					if err != nil {
						return err
					}
					for _, f := range possibleDevfileNames {
						if _, err = fs.Create(filepath.Join(dir, f+".bak")); err != nil {
							return err
						}
					}
					return err
				},
			},
			wantErr: true,
		},
		{
			name: "directory with all possible devfile filenames => priority to devfile.yaml",
			fields: fields{
				relPath: "my-devfiles",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					dir := "my-devfiles"
					err := fs.MkdirAll("my-devfiles", 0755)
					if err != nil {
						return err
					}
					for _, f := range possibleDevfileNames {
						if _, err = fs.Create(filepath.Join(dir, f)); err != nil {
							return err
						}
					}
					return err
				},
			},
			wantErr: false,
			want:    "my-devfiles/devfile.yaml",
		},
		{
			name: "directory with missing devfile.yaml => priority to .devfile.yaml",
			fields: fields{
				relPath: "my-devfiles",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					dir := "my-devfiles"
					err := fs.MkdirAll("my-devfiles", 0755)
					if err != nil {
						return err
					}
					for _, f := range possibleDevfileNames {
						if f == "devfile.yaml" {
							continue
						}
						if _, err = fs.Create(filepath.Join(dir, f)); err != nil {
							return err
						}
					}
					return err
				},
			},
			wantErr: false,
			want:    "my-devfiles/.devfile.yaml",
		},
		{
			name: "directory with missing devfile.yaml and .devfile.yaml => priority to devfile.yml",
			fields: fields{
				relPath: "my-devfiles",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					dir := "my-devfiles"
					err := fs.MkdirAll("my-devfiles", 0755)
					if err != nil {
						return err
					}
					for _, f := range possibleDevfileNames {
						if f == "devfile.yaml" || f == ".devfile.yaml" {
							continue
						}
						if _, err = fs.Create(filepath.Join(dir, f)); err != nil {
							return err
						}
					}
					return err
				},
			},
			wantErr: false,
			want:    "my-devfiles/devfile.yml",
		},
		{
			name: "directory with missing devfile.yaml and .devfile.yaml and devfile.yml => priority to .devfile.yml",
			fields: fields{
				relPath: "my-devfiles",
				fsCustomizer: func(fs filesystem.Filesystem) error {
					dir := "my-devfiles"
					err := fs.MkdirAll("my-devfiles", 0755)
					if err != nil {
						return err
					}
					for _, f := range possibleDevfileNames {
						if f == "devfile.yaml" || f == ".devfile.yaml" || f == "devfile.yml" {
							continue
						}
						if _, err = fs.Create(filepath.Join(dir, f)); err != nil {
							return err
						}
					}
					return err
				},
			},
			wantErr: false,
			want:    "my-devfiles/.devfile.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := filesystem.NewFakeFs()
			var err error
			if tt.fields.fsCustomizer != nil {
				err = tt.fields.fsCustomizer(fs)
			}
			if err != nil {
				t.Fatalf("unexpected error while setting up filesystem: %v", err)
			}
			got, err := lookupDevfileFromPath(fs, tt.fields.relPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("lookupDevfileFromPath(): unexpected error: %v. wantErr=%v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("lookupDevfileFromPath(): mismatch (-want +got): %s\n", diff)
			}
		})
	}
}
