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
	"os"
	"testing"

	"github.com/devfile/library/pkg/testingutil/filesystem"
)

const (
	TempJSONDevfilePrefix = "odo-devfile.*.json"
	InvalidDevfileContent = ":: invalid :: content"
)

func TestSetDevfileContent(t *testing.T) {

	const (
		InvalidDevfilePath = "/invalid/path"
	)

	// createTempDevfile helper creates temp devfile
	createTempDevfile := func(t *testing.T, content []byte, fakeFs filesystem.Filesystem) (f filesystem.File) {

		t.Helper()

		// Create tempfile
		f, err := fakeFs.TempFile(os.TempDir(), TempJSONDevfilePrefix)
		if err != nil {
			t.Errorf("failed to create temp devfile, %v", err)
			return f
		}

		// Write content to devfile
		if _, err := f.Write(content); err != nil {
			t.Errorf("failed to write to temp devfile")
			return f
		}

		// Successful
		return f
	}

	t.Run("valid file", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, validJsonRawContent200(), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContent()

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid content", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, []byte(InvalidDevfileContent), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContent()

		if err == nil {
			t.Errorf("expected error, didn't get one ")
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid filepath", func(t *testing.T) {

		var (
			fakeFs = filesystem.NewFakeFs()
			d      = DevfileCtx{
				absPath: InvalidDevfilePath,
				fs:      fakeFs,
			}
		)

		err := d.SetDevfileContent()

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}

func TestSetDevfileContentFromBytes(t *testing.T) {

	// createTempDevfile helper creates temp devfile
	createTempDevfile := func(t *testing.T, content []byte, fakeFs filesystem.Filesystem) (f filesystem.File) {

		t.Helper()

		// Create tempfile
		f, err := fakeFs.TempFile(os.TempDir(), TempJSONDevfilePrefix)
		if err != nil {
			t.Errorf("failed to create temp devfile, %v", err)
			return f
		}

		// Write content to devfile
		if _, err := f.Write(content); err != nil {
			t.Errorf("failed to write to temp devfile")
			return f
		}

		// Successful
		return f
	}

	t.Run("valid data passed", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, validJsonRawContent200(), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				fs:      fakeFs,
			}
		)

		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContentFromBytes(validJsonRawContent200())

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid data passed", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, []byte(validJsonRawContent200()), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContentFromBytes([]byte(InvalidDevfileContent))

		if err == nil {
			t.Errorf("expected error, didn't get one ")
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})
}
