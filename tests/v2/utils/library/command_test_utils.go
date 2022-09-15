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

// getSchemaCommand get a specified command from the devfile schema structure
func getSchemaCommand(commands []schema.Command, id string) (*schema.Command, bool) {
	found := false
	var schemaCommand schema.Command
	for _, command := range commands {
		if command.Id == id {
			schemaCommand = command
			found = true
			break
		}
	}
	return &schemaCommand, found
}

// UpdateCommand randomly updates attribute values of a specified command in the devfile schema
func UpdateCommand(devfile *commonUtils.TestDevfile, commandId string) error {

	var err error
	testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, commandId)
	if found {
		commonUtils.LogInfoMessage(fmt.Sprintf("Updating command id: %s", commandId))
		if testCommand.Exec != nil {
			devfile.SetExecCommandValues(testCommand)
		} else if testCommand.Composite != nil {
			devfile.SetCompositeCommandValues(testCommand)
		} else if testCommand.Apply != nil {
			devfile.SetApplyCommandValues(testCommand)
		}
	} else {
		err = errors.New(commonUtils.LogErrorMessage(fmt.Sprintf("Command not found in test : %s", commandId)))
	}
	return err
}

// VerifyCommands verifies commands returned by the parser are the same as those saved in the devfile schema
func VerifyCommands(devfile *commonUtils.TestDevfile, parserCommands []schema.Command) error {

	commonUtils.LogInfoMessage("Enter VerifyCommands")
	var errorString []string

	// Compare entire array of commands
	if !cmp.Equal(parserCommands, devfile.SchemaDevFile.Commands) {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Command array compare failed.")))
		// Array compare failed. Narrow down by comparing indivdual commands
		for _, command := range parserCommands {
			if testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, command.Id); found {
				if !cmp.Equal(command, *testCommand) {
					parserFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Parser")
					testFilename := commonUtils.AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Test")
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", devfile.FileName))
					c, err := yaml.Marshal(command)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = ioutil.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					commonUtils.LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testCommand)
					if err != nil {
						errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = ioutil.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, commonUtils.LogInfoMessage(fmt.Sprintf("Command %s did not match, see files : %s and %s", command.Id, parserFilename, testFilename)))
				} else {
					commonUtils.LogInfoMessage(fmt.Sprintf(" --> Command  matched : %s", command.Id))
				}
			} else {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Command from parser not known to test - id : %s ", command.Id)))
			}

		}
		for _, command := range devfile.SchemaDevFile.Commands {
			if _, found := getSchemaCommand(parserCommands, command.Id); !found {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Command from test not returned by parser : %s ", command.Id)))
			}
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf(" --> Command structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}
