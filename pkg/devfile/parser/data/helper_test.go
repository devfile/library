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

package data

import (
	"reflect"
	"strings"
	"testing"

	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	v200 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.0.0"
)

func TestNewDevfileData(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version  = APISchemaVersion200
			want     = reflect.TypeOf(&v2.DevfileV2{})
			obj, err = NewDevfileData(string(version))
			got      = reflect.TypeOf(obj)
		)

		// got and want should be equal
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: '%v', want: '%s'", got, want)
		}

		// no error should be received
		if err != nil {
			t.Errorf("did not expect an error '%v'", err)
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			_, err  = NewDevfileData(string(version))
		)

		// no error should be received
		if err == nil {
			t.Errorf("did not expect an error '%v'", err)
		}
	})
}

func TestGetDevfileJSONSchema(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version  = APISchemaVersion200
			want     = v200.JsonSchema200
			got, err = GetDevfileJSONSchema(string(version))
		)

		if err != nil {
			t.Errorf("did not expect an error '%v'", err)
		}

		if strings.Compare(got, want) != 0 {
			t.Errorf("incorrect json schema")
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			_, err  = GetDevfileJSONSchema(string(version))
		)

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}

func TestIsApiVersionSupported(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version = APISchemaVersion200
			want    = true
			got     = IsApiVersionSupported(string(version))
		)

		if got != want {
			t.Errorf("want: '%t', got: '%t'", want, got)
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			want    = false
			got     = IsApiVersionSupported(string(version))
		)

		if got != want {
			t.Errorf("expected an error, didn't get one")
		}
	})
}
