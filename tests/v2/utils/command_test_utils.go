package utils

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// commandAdded adds a new command to the test schema data and to the parser data
func (devfile *TestDevfile) commandAdded(command schema.Command) {
	LogInfoMessage(fmt.Sprintf("command added Id: %s", command.Id))
	devfile.SchemaDevFile.Commands = append(devfile.SchemaDevFile.Commands, command)
	devfile.ParserData.AddCommands([]schema.Command{command})
}

// commandUpdated updates a command in the parser data
func (devfile *TestDevfile) commandUpdated(command schema.Command) {
	LogInfoMessage(fmt.Sprintf("command updated Id: %s", command.Id))
	devfile.ParserData.UpdateCommand(command)
}

// addEnv creates and returns a specifed number of env attributes in a schema structure
func addEnv(numEnv int) []schema.EnvVar {
	commandEnvs := make([]schema.EnvVar, numEnv)
	for i := 0; i < numEnv; i++ {
		commandEnvs[i].Name = "Name_" + GetRandomString(5, false)
		commandEnvs[i].Value = "Value_" + GetRandomString(5, false)
		LogInfoMessage(fmt.Sprintf("Add Env: %s", commandEnvs[i]))
	}
	return commandEnvs
}

// addAttributes creates returns a specifed number of attributes in a schema structure
func addAttributes(numAtrributes int) map[string]string {
	attributes := make(map[string]string)
	for i := 0; i < numAtrributes; i++ {
		AttributeName := "Name_" + GetRandomString(6, false)
		attributes[AttributeName] = "Value_" + GetRandomString(6, false)
		LogInfoMessage(fmt.Sprintf("Add attribute : %s = %s", AttributeName, attributes[AttributeName]))
	}
	return attributes
}

// addGroup creates and returns a group in a schema structure
func (devfile *TestDevfile) addGroup() *schema.CommandGroup {

	commandGroup := schema.CommandGroup{}
	commandGroup.Kind = GetRandomGroupKind()
	LogInfoMessage(fmt.Sprintf("group Kind: %s, default already set %t", commandGroup.Kind, devfile.GroupDefaults[commandGroup.Kind]))
	// Ensure only one and at least one of each type are labelled as default
	if !devfile.GroupDefaults[commandGroup.Kind] {
		devfile.GroupDefaults[commandGroup.Kind] = true
		commandGroup.IsDefault = true
	} else {
		commandGroup.IsDefault = false
	}
	LogInfoMessage(fmt.Sprintf("group isDefault: %t", commandGroup.IsDefault))
	return &commandGroup
}

// AddCommand creates a command of a specified type in a schema structure and pupulates it with random attributes
func (devfile *TestDevfile) AddCommand(commandType schema.CommandType) schema.Command {

	var command *schema.Command
	if commandType == schema.ExecCommandType {
		command = devfile.createExecCommand()
		devfile.setExecCommandValues(command)
		// command must be mentioned by a container component
		command.Exec.Component = devfile.GetContainerName()
	} else if commandType == schema.CompositeCommandType {
		command = devfile.createCompositeCommand()
		devfile.setCompositeCommandValues(command)
	}
	return *command
}

// UpdateCommand randomly updates attribute values of a specified command in the devfile schema
func (devfile *TestDevfile) UpdateCommand(commandId string) error {

	var err error
	testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, commandId)
	if found {
		LogInfoMessage(fmt.Sprintf("Updating command id: %s", commandId))
		if testCommand.Exec != nil {
			devfile.setExecCommandValues(testCommand)
		} else if testCommand.Composite != nil {
			devfile.setCompositeCommandValues(testCommand)
		}
	} else {
		err = errors.New(LogErrorMessage(fmt.Sprintf("Command not found in test : %s", commandId)))
	}
	return err
}

// createExecCommand creates and returns an empty exec command in a schema structure
func (devfile *TestDevfile) createExecCommand() *schema.Command {

	LogInfoMessage("Create an exec command :")
	command := schema.Command{}
	command.Id = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("command Id: %s", command.Id))
	command.Exec = &schema.ExecCommand{}
	devfile.commandAdded(command)
	return &command

}

// setExecCommandValues randomly sets exec command attribute to random values
func (devfile *TestDevfile) setExecCommandValues(command *schema.Command) {

	execCommand := command.Exec
	execCommand.CommandLine = GetRandomString(4, false) + " " + GetRandomString(4, false)
	LogInfoMessage(fmt.Sprintf("....... commandLine: %s", execCommand.CommandLine))

	// If group already leave it to make sure defaults are not deleted or added
	if execCommand.Group == nil {
		if GetRandomDecision(2, 1) {
			execCommand.Group = devfile.addGroup()
		}
	}

	if GetBinaryDecision() {
		execCommand.Label = GetRandomString(12, false)
		LogInfoMessage(fmt.Sprintf("....... label: %s", execCommand.Label))
	} else {
		execCommand.Label = ""
	}

	if GetBinaryDecision() {
		execCommand.WorkingDir = "./tmp"
		LogInfoMessage(fmt.Sprintf("....... WorkingDir: %s", execCommand.WorkingDir))
	} else {
		execCommand.WorkingDir = ""
	}

	execCommand.HotReloadCapable = GetBinaryDecision()
	LogInfoMessage(fmt.Sprintf("....... HotReloadCapable: %t", execCommand.HotReloadCapable))

	if GetBinaryDecision() {
		execCommand.Env = addEnv(GetRandomNumber(4))
	} else {
		execCommand.Env = nil
	}
	devfile.commandUpdated(*command)

}

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

// createCompositeCommand creates an empty composite command in a schema structure
func (devfile *TestDevfile) createCompositeCommand() *schema.Command {

	LogInfoMessage("Create a composite command :")
	command := schema.Command{}
	command.Id = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("command Id: %s", command.Id))
	command.Composite = &schema.CompositeCommand{}
	devfile.commandAdded(command)

	return &command
}

// setCompositeCommandValues randomly sets composite command attribute to random values
func (devfile *TestDevfile) setCompositeCommandValues(command *schema.Command) {

	compositeCommand := command.Composite
	numCommands := GetRandomNumber(3)

	for i := 0; i < numCommands; i++ {
		execCommand := devfile.AddCommand(schema.ExecCommandType)
		compositeCommand.Commands = append(compositeCommand.Commands, execCommand.Id)
		LogInfoMessage(fmt.Sprintf("....... command %d of %d : %s", i, numCommands, execCommand.Id))
	}

	// If group already exists - leave it to make sure defaults are not deleted or added
	if compositeCommand.Group == nil {
		if GetRandomDecision(2, 1) {
			compositeCommand.Group = devfile.addGroup()
		}
	}

	if GetBinaryDecision() {
		compositeCommand.Label = GetRandomString(12, false)
		LogInfoMessage(fmt.Sprintf("....... label: %s", compositeCommand.Label))
	}

	if GetBinaryDecision() {
		compositeCommand.Parallel = true
		LogInfoMessage(fmt.Sprintf("....... Parallel: %t", compositeCommand.Parallel))
	}

	devfile.commandUpdated(*command)
}

// VerifyCommands verifies commands returned by the parser are the same as those saved in the devfile schema
func (devfile *TestDevfile) VerifyCommands(parserCommands []schema.Command) error {

	LogInfoMessage("Enter VerifyCommands")
	var errorString []string

	// Compare entire array of commands
	if !cmp.Equal(parserCommands, devfile.SchemaDevFile.Commands) {
		errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Command array compare failed.")))
		// Array compare failed. Narrow down by comparing indivdual commands
		for _, command := range parserCommands {
			if testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, command.Id); found {
				if !cmp.Equal(command, *testCommand) {
					parserFilename := AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Parser")
					testFilename := AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Test")
					LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", devfile.FileName))
					c, err := yaml.Marshal(command)
					if err != nil {
						errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = ioutil.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testCommand)
					if err != nil {
						errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = ioutil.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, LogInfoMessage(fmt.Sprintf("Command %s did not match, see files : %s and %s", command.Id, parserFilename, testFilename)))
				} else {
					LogInfoMessage(fmt.Sprintf(" --> Command  matched : %s", command.Id))
				}
			} else {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Command from parser not known to test - id : %s ", command.Id)))
			}

		}
		for _, command := range devfile.SchemaDevFile.Commands {
			if _, found := getSchemaCommand(parserCommands, command.Id); !found {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Command from test not returned by parser : %s ", command.Id)))
			}
		}
	} else {
		LogInfoMessage(fmt.Sprintf(" --> Command structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}
