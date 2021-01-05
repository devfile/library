package tests

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// Return a specifed number of env attributes in a schema structure
func AddEnv(numEnv int) []schema.EnvVar {
	commandEnvs := make([]schema.EnvVar, numEnv)
	for i := 0; i < numEnv; i++ {
		commandEnvs[i].Name = "Name_" + GetRandomString(5, false)
		commandEnvs[i].Value = "Value_" + GetRandomString(5, false)
		LogInfoMessage(fmt.Sprintf("Add Env: %s", commandEnvs[i]))
	}
	return commandEnvs
}

// Return a specifed number of attributes in a schema structure
func AddAttributes(numAtrributes int) map[string]string {
	attributes := make(map[string]string)
	for i := 0; i < numAtrributes; i++ {
		AttributeName := "Name_" + GetRandomString(6, false)
		attributes[AttributeName] = "Value_" + GetRandomString(6, false)
		LogInfoMessage(fmt.Sprintf("Add attribute : %s = %s", AttributeName, attributes[AttributeName]))
	}
	return attributes
}

// Create and return a group in a schema structure
func addGroup() *schema.CommandGroup {

	commandGroup := schema.CommandGroup{}
	commandGroup.Kind = GetRandomGroupKind()
	LogInfoMessage(fmt.Sprintf("group Kind: %s", commandGroup.Kind))
	commandGroup.IsDefault = GetBinaryDecision()
	LogInfoMessage(fmt.Sprintf("group isDefault: %t", commandGroup.IsDefault))
	return &commandGroup
}

// Add a command of the specified type to the schema
func (devfile *TestDevfile) addCommand(commandType schema.CommandType) string {
	command := generateCommand(commandType)
	devfile.SchemaDevFile.Commands = append(devfile.SchemaDevFile.Commands, command)
	return command.Id
}

// Create  a command of a specified type in a schema structure
func generateCommand(commandType schema.CommandType) schema.Command {
	command := schema.Command{}
	command.Id = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("command Id: %s", command.Id))

	if commandType == schema.ExecCommandType {
		command.Exec = createExecCommand()
	} else if commandType == schema.CompositeCommandType {
		command.Composite = createCompositeCommand()
	}
	return command
}

// Update the values of a specified command
func (devfile *TestDevfile) UpdateCommand(parserCommand *schema.Command) error {

	var errorString []string
	testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, parserCommand.Id)
	if found {
		LogInfoMessage(fmt.Sprintf("Updating command id: %s", parserCommand.Id))
		if testCommand.Exec != nil {
			setExecCommandValues(parserCommand.Exec)
		} else if testCommand.Composite != nil {
			setCompositeCommandValues(parserCommand.Composite)
		}
		devfile.replaceSchemaCommand(*parserCommand)
	} else {
		errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Command not found in test : %s", parserCommand.Id)))
	}

	var returnError error
	if len(errorString) > 0 {
		returnError = errors.New(fmt.Sprint(errorString))
	}
	return returnError
}

// Create and return an exec command in a schema structure
func createExecCommand() *schema.ExecCommand {

	LogInfoMessage("Create an exec command :")
	execCommand := schema.ExecCommand{}
	setExecCommandValues(&execCommand)
	return &execCommand

}

// Set the attribute values of an exec command
func setExecCommandValues(execCommand *schema.ExecCommand) {

	execCommand.Component = GetRandomString(8, false)
	LogInfoMessage(fmt.Sprintf("....... component: %s", execCommand.Component))

	execCommand.CommandLine = GetRandomString(4, false) + " " + GetRandomString(4, false)
	LogInfoMessage(fmt.Sprintf("....... commandLine: %s", execCommand.CommandLine))

	if GetRandomDecision(2, 1) {
		execCommand.Group = addGroup()
	} else {
		execCommand.Group = nil
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
		execCommand.Env = AddEnv(GetRandomNumber(4))
	} else {
		execCommand.Env = nil
	}

}

// Use the specified command to replace the command in the schema structure with the same Id.
func (devfile TestDevfile) replaceSchemaCommand(command schema.Command) {
	for i := 0; i < len(devfile.SchemaDevFile.Commands); i++ {
		if devfile.SchemaDevFile.Commands[i].Id == command.Id {
			devfile.SchemaDevFile.Commands[i] = command
			break
		}
	}
}

// Get a command from the schema structure
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

// Create a composite command in a schema structure
func createCompositeCommand() *schema.CompositeCommand {

	LogInfoMessage("Create a composite command :")
	compositeCommand := schema.CompositeCommand{}
	setCompositeCommandValues(&compositeCommand)
	return &compositeCommand
}

// Set the attribute values for a composite command
func setCompositeCommandValues(compositeCommand *schema.CompositeCommand) {
	numCommands := GetRandomNumber(3)

	compositeCommand.Commands = make([]string, numCommands)
	for i := 0; i < numCommands; i++ {
		compositeCommand.Commands[i] = GetRandomUniqueString(8, false)
		LogInfoMessage(fmt.Sprintf("....... command %d of %d : %s", i, numCommands, compositeCommand.Commands[i]))
	}

	if GetRandomDecision(2, 1) {
		compositeCommand.Group = addGroup()
	}

	if GetBinaryDecision() {
		compositeCommand.Label = GetRandomString(12, false)
		LogInfoMessage(fmt.Sprintf("....... label: %s", compositeCommand.Label))
	}

	if GetBinaryDecision() {
		compositeCommand.Parallel = true
		LogInfoMessage(fmt.Sprintf("....... Parallel: %t", compositeCommand.Parallel))
	}
}

// Verify commands returned by the parser match with those saved in the schema
func (devfile TestDevfile) VerifyCommands(parserCommands []schema.Command) error {

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
