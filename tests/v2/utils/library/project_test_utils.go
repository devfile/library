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
	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	commonUtils "github.com/devfile/api/v2/test/v200/utils/common"
)

// getSchemaProject gets a named Project from the saved devfile schema structure
func getSchemaProject(projects []schema.Project, name string) (*schema.Project, bool) {
	found := false
	var schemaProject schema.Project
	for _, project := range projects {
		if project.Name == name {
			schemaProject = project
			found = true
			break
		}
	}
	return &schemaProject, found
}

// getSchemaStarterProject gets a named Starter Project from the saved devfile schema structure
func getSchemaStarterProject(starterProjects []schema.StarterProject, name string) (*schema.StarterProject, bool) {
	found := false
	var schemaStarterProject schema.StarterProject
	for _, starterProject := range starterProjects {
		if starterProject.Name == name {
			schemaStarterProject = starterProject
			found = true
			break
		}
	}
	return &schemaStarterProject, found
}

// UpdateProject randomly modifies an existing project
func UpdateProject(devfile *commonUtils.TestDevfile, projectName string) error {

	var err error
	testProject, found := getSchemaProject(devfile.SchemaDevFile.Projects, projectName)
	if found {
		commonUtils.LogInfoMessage(fmt.Sprintf("Updating Project : %s", projectName))
		devfile.SetProjectValues(testProject)
	} else {
		err = errors.New(commonUtils.LogErrorMessage(fmt.Sprintf("Project not found in test : %s", projectName)))
	}
	return err

}

// UpdateStarterProject randomly modifies an existing starter project
func UpdateStarterProject(devfile *commonUtils.TestDevfile, projectName string) error {

	var err error
	testStarterProject, found := getSchemaStarterProject(devfile.SchemaDevFile.StarterProjects, projectName)
	if found {
		commonUtils.LogInfoMessage(fmt.Sprintf("Updating Starter Project : %s", projectName))
		devfile.SetStarterProjectValues(testStarterProject)
	} else {
		err = errors.New(commonUtils.LogErrorMessage(fmt.Sprintf("Starter Project not found in test : %s", projectName)))
	}
	return err
}

// VerifyProjects verifies projects returned by the parser are the same as those saved in the devfile schema
func VerifyProjects(devfile *commonUtils.TestDevfile, parserProjects []schema.Project) error {

	commonUtils.LogInfoMessage("Enter VerifyProjects")
	var errorString []string

	// Compare entire array of projects
	if !cmp.Equal(parserProjects, devfile.SchemaDevFile.Projects) {
		// Compare failed so compare each project to find which one(s) don't compare
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Project array compare failed.")))
		for _, project := range parserProjects {
			if testProject, found := getSchemaProject(devfile.SchemaDevFile.Projects, project.Name); found {
				if !cmp.Equal(project, *testProject) {
					// Write out the failing project to a file, once as expected by the test, and a second as returned by the parser
					parserFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+project.Name+"_Parser")
					testFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+project.Name+"_Test")
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", parserFilename))
					c, err := yaml.Marshal(project)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = ioutil.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testProject)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = ioutil.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Project %s did not match, see files : %s and %s", project.Name, parserFilename, testFilename)))
				} else {
					commonUtils.LogInfoMessage(fmt.Sprintf(" --> Project matched : %s", project.Name))
				}
			} else {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Project from parser not known to test - id : %s ", project.Name)))
			}
		}
		// Check test does not include projects which the parser did not return
		for _, project := range devfile.SchemaDevFile.Projects {
			if _, found := getSchemaProject(parserProjects, project.Name); !found {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Project from test not returned by parser : %s ", project.Name)))
			}
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Project structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// VerifyStarterProjects verifies starter projects returned by the parser are the same as those saved in the devfile schema
func VerifyStarterProjects(devfile *commonUtils.TestDevfile, parserStarterProjects []schema.StarterProject) error {

	commonUtils.LogInfoMessage("Enter VerifyStarterProjects")
	var errorString []string

	// Compare entire array of projects
	if !cmp.Equal(parserStarterProjects, devfile.SchemaDevFile.StarterProjects) {
		// Compare failed so compare each project to find which one(s) don't compare
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Starter Project array compare failed.")))
		for _, starterProject := range parserStarterProjects {
			if testStarterProject, found := getSchemaStarterProject(devfile.SchemaDevFile.StarterProjects, starterProject.Name); found {
				if !cmp.Equal(starterProject, *testStarterProject) {
					// Write out the failing starter project to a file, once as expected by the test, and a second as returned by the parser
					parserFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+starterProject.Name+"_Parser")
					testFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+starterProject.Name+"_Test")
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", parserFilename))
					c, err := yaml.Marshal(starterProject)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = ioutil.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testStarterProject)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = ioutil.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Starter Project %s did not match, see files : %s and %s", starterProject.Name, parserFilename, testFilename)))
				} else {
					commonUtils.LogInfoMessage(fmt.Sprintf(" --> Starter Project matched : %s", starterProject.Name))
				}
			} else {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Starter Project from parser not known to test - id : %s ", starterProject.Name)))
			}
		}
		// Check test does not include projects which the parser did not return
		for _, starterProject := range devfile.SchemaDevFile.StarterProjects {
			if _, found := getSchemaStarterProject(parserStarterProjects, starterProject.Name); !found {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Starter Project from test not returned by parser : %s ", starterProject.Name)))
			}
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Starter Project structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}
