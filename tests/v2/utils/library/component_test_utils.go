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

package utils

import (
	"errors"
	"fmt"
	"os"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	commonUtils "github.com/devfile/api/v2/test/v200/utils/common"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

// getSchemaComponent returns a named component from an array of components
func getSchemaComponent(components []schema.Component, name string) (*schema.Component, bool) {
	found := false
	var schemaComponent schema.Component
	for _, component := range components {
		if component.Name == name {
			schemaComponent = component
			found = true
			break
		}
	}
	return &schemaComponent, found
}

// UpdateComponent randomly updates the attribute values of a specified component
func UpdateComponent(devfile *commonUtils.TestDevfile, componentName string) error {

	var errorString []string
	testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, componentName)
	if found {
		commonUtils.LogInfoMessage(fmt.Sprintf("....... Updating component name: %s", componentName))
		if testComponent.Container != nil {
			devfile.SetContainerComponentValues(testComponent)
		} else if testComponent.Kubernetes != nil {
			devfile.SetK8sComponentValues(testComponent)
		} else if testComponent.Openshift != nil {
			devfile.SetK8sComponentValues(testComponent)
		} else if testComponent.Volume != nil {
			devfile.SetVolumeComponentValues(testComponent)
		} else {
			errorString = append(errorString, commonUtils.LogInfoMessage(fmt.Sprintf("....... Component is not of expected type.")))
		}
	} else {
		errorString = append(errorString, commonUtils.LogInfoMessage(fmt.Sprintf("....... Component not found in test : %s", componentName)))
	}
	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// VerifyComponents verifies components returned by the parser are the same as those saved in the devfile schema
func VerifyComponents(devfile *commonUtils.TestDevfile, parserComponents []schema.Component) error {

	commonUtils.LogInfoMessage("Enter VerifyComponents")
	var errorString []string

	// Compare entire array of components
	if !cmp.Equal(parserComponents, devfile.SchemaDevFile.Components) {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Component array compare failed.")))
		for _, component := range parserComponents {
			if testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, component.Name); found {
				if !cmp.Equal(component, *testComponent) {
					parserFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Parser")
					testFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Test")
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", parserFilename))
					c, err := yaml.Marshal(component)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = os.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testComponent)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = os.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Component %s did not match, see files : %s and %s", component.Name, parserFilename, testFilename)))
				} else {
					commonUtils.LogInfoMessage(fmt.Sprintf(" --> Component matched : %s", component.Name))
				}
			} else {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Component from parser not known to test - id : %s ", component.Name)))
			}
		}
		for _, component := range devfile.SchemaDevFile.Components {
			if _, found := getSchemaComponent(parserComponents, component.Name); !found {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Component from test not returned by parser : %s ", component.Name)))
			}
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Component structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}
