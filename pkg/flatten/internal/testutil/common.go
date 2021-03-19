//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package testutil

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"sigs.k8s.io/yaml"
)

// WorkspaceTemplateDiffOpts are used to compare test output against the expected result.
var WorkspaceTemplateDiffOpts = cmp.Options{
	cmpopts.SortSlices(func(a, b dw.Component) bool {
		return strings.Compare(a.Key(), b.Key()) > 0
	}),
	cmpopts.SortSlices(func(a, b string) bool {
		return strings.Compare(a, b) > 0
	}),
	// TODO: Devfile overriding results in empty []string instead of nil
	cmpopts.IgnoreFields(dw.WorkspaceEvents{}, "PostStart", "PreStop", "PostStop"),
}

// TestCase describes a single test case for the library.
type TestCase struct {
	// Name is a descriptive name of what is being tested
	Name string `json:"name"`
	// Input describes the test inputs
	Input  TestInput  `json:"input"`
	Output TestOutput `json:"output"`
}

// TestInput defines the inputs required for a test case.
type TestInput struct {
	Workspace dw.DevWorkspaceTemplateSpec `json:"workspace,omitempty"`
	// Plugins is a map of plugin "name" to devworkspace template; namespace is ignored.
	Plugins map[string]dw.DevWorkspaceTemplate `json:"plugins,omitempty"`
	// DevfilePlugins is a map of plugin "name" to devfile
	DevfilePlugins map[string]dw.Devfile `json:"devfilePlugins,omitempty"`
	// Errors is a map of plugin name to the error that should be returned when attempting to retrieve it.
	Errors map[string]TestPluginError `json:"errors,omitempty"`
}

// TestPluginError describes an expected error.
type TestPluginError struct {
	// IsNotFound marks this error as a kubernetes NotFoundError
	IsNotFound bool `json:"isNotFound"`
	// StatusCode defines the HTTP response code (if relevant)
	StatusCode int `json:"statusCode"`
	// Message is the error message returned
	Message string `json:"message"`
}

// TestOutput describes expected test outputs. If errRegexp is not empty, it is compared to the returned error as a regular
// expression. Otherwise, the output Workspace is compared with the output of the function call.
type TestOutput struct {
	Workspace *dw.DevWorkspaceTemplateSpec `json:"workspace,omitempty"`
	ErrRegexp *string                      `json:"errRegexp,omitempty"`
}

// LoadTestCaseOrPanic loads the test file at testFilepath.
func LoadTestCaseOrPanic(t *testing.T, testFilepath string) TestCase {
	bytes, err := ioutil.ReadFile(testFilepath)
	if err != nil {
		t.Fatal(err)
	}
	var test TestCase
	if err := yaml.Unmarshal(bytes, &test); err != nil {
		t.Fatal(err)
	}
	return test
}

// LoadAllTestsOrPanic loads all yaml files in fromDir as test cases.
func LoadAllTestsOrPanic(t *testing.T, fromDir string) []TestCase {
	files, err := ioutil.ReadDir(fromDir)
	if err != nil {
		t.Fatal(err)
	}
	var tests []TestCase
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		tests = append(tests, LoadTestCaseOrPanic(t, filepath.Join(fromDir, file.Name())))
	}
	return tests
}
