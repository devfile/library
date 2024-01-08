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
	"fmt"
	"reflect"
	"testing"

	errPkg "github.com/devfile/library/v2/pkg/devfile/parser/errors"
)

func TestSetDevfileAPIVersion(t *testing.T) {

	const (
		schemaVersion          = "2.2.0"
		validJson              = `{"schemaVersion": "2.2.0"}`
		concreteSchema         = `{"schemaVersion": "2.2.0-latest"}`
		emptyJson              = "{}"
		emptySchemaVersionJson = `{"schemaVersion": ""}`
		badJson                = `{"name": "Joe", "age": null]`
		devfilePath            = "/testpath/devfile.yaml"
		devfileURL             = "http://server/devfile.yaml"
	)

	// test table
	tests := []struct {
		name       string
		devfileCtx DevfileCtx
		want       string
		wantErr    error
	}{
		{
			name:       "valid schemaVersion",
			devfileCtx: DevfileCtx{rawContent: []byte(validJson), absPath: devfilePath},
			want:       schemaVersion,
			wantErr:    nil,
		},
		{
			name:       "concrete schemaVersion",
			devfileCtx: DevfileCtx{rawContent: []byte(concreteSchema), absPath: devfilePath},
			want:       schemaVersion,
			wantErr:    nil,
		},
		{
			name:       "schemaVersion not present",
			devfileCtx: DevfileCtx{rawContent: []byte(emptyJson), absPath: devfilePath},
			want:       "",
			wantErr:    &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("schemaVersion not present in devfile: %s", devfilePath)},
		},
		{
			name:       "schemaVersion empty",
			devfileCtx: DevfileCtx{rawContent: []byte(emptySchemaVersionJson), url: devfileURL},
			want:       "",
			wantErr:    &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("schemaVersion in devfile: %s cannot be empty", devfileURL)},
		},
		{
			name:       "unmarshal error",
			devfileCtx: DevfileCtx{rawContent: []byte(badJson), url: devfileURL},
			want:       "",
			wantErr:    &errPkg.NonCompliantDevfile{Err: "invalid character ']' after object key:value pair"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// new devfile context object
			d := tt.devfileCtx

			// SetDevfileAPIVersion
			gotErr := d.SetDevfileAPIVersion()
			got := d.apiVersion

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("TestSetDevfileAPIVersion() unexpected error: '%v', wantErr: '%v'", gotErr, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("TestSetDevfileAPIVersion() want: '%v', got: '%v'", tt.want, got)
			}
		})
	}
}

func TestGetApiVersion(t *testing.T) {

	const (
		apiVersion = "2.0.0"
	)

	t.Run("get apiVersion", func(t *testing.T) {

		var (
			d    = DevfileCtx{apiVersion: apiVersion}
			want = apiVersion
			got  = d.GetApiVersion()
		)

		if got != want {
			t.Errorf("TestGetApiVersion() want: '%v', got: '%v'", want, got)
		}
	})
}
