package parserV200

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func AddEnv(numEnv int) []schema.EnvVar {
	commandEnvs := make([]schema.EnvVar, numEnv)
	for i := 0; i < numEnv; i++ {
		commandEnvs[i].Name = "Name_" + GetRandomString(5, false)
		commandEnvs[i].Value = "Value_" + GetRandomString(5, false)
		LogMessage(fmt.Sprintf("   ....... Add Env: %s", commandEnvs[i]))
	}
	return commandEnvs
}

func AddAttributes(numAtrributes int) map[string]string {
	attributes := make(map[string]string)
	for i := 0; i < numAtrributes; i++ {
		AttributeName := "Name_" + GetRandomString(6, false)
		attributes[AttributeName] = "Value_" + GetRandomString(6, false)
		LogMessage(fmt.Sprintf("   ....... Add attribute : %s = %s", AttributeName, attributes[AttributeName]))
	}
	return attributes
}

func addGroup() *schema.CommandGroup {

	commandGroup := schema.CommandGroup{}

	commandGroup.Kind = GetRandomGroupKind()
	LogMessage(fmt.Sprintf("   ....... group Kind: %s", commandGroup.Kind))
	commandGroup.IsDefault = GetBinaryDecision()
	LogMessage(fmt.Sprintf("   ....... group isDefault: %t", commandGroup.IsDefault))

	return &commandGroup
}

func (devfile *TestDevfile) addCommand(commandType schema.CommandType) string {

	var commands []schema.Command
	index := 0
	if devfile.SchemaDevFile.Commands != nil {
		// Commands already exist so expand the current slice
		currentSize := len(devfile.SchemaDevFile.Commands)
		commands = make([]schema.Command, currentSize+1)
		for _, command := range devfile.SchemaDevFile.Commands {
			commands[index] = command
			index++
		}
	} else {
		commands = make([]schema.Command, 1)
	}

	generateCommand(&commands[index], commandType)
	devfile.SchemaDevFile.Commands = commands

	return commands[index].Id

}

func generateCommand(command *schema.Command, commandType schema.CommandType) {
	command.Id = GetRandomUniqueString(8, true)
	LogMessage(fmt.Sprintf("   ....... id: %s", command.Id))

	if commandType == schema.ExecCommandType {
		command.Exec = createExecCommand()
	} else if commandType == schema.CompositeCommandType {
		command.Composite = createCompositeCommand()
	}
}

func (devfile *TestDevfile) UpdateCommand(parserCommand *schema.Command) error {

	errorString := ""
	testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, parserCommand.Id)
	if found {
		LogMessage(fmt.Sprintf(" ....... Updating command id: %s", parserCommand.Id))
		if testCommand.Exec != nil {
			setExecCommandValues(parserCommand.Exec)
		} else if testCommand.Composite != nil {
			setCompositeCommandValues(parserCommand.Composite)
		}
		devfile.replaceSchemaCommand(*parserCommand)
	} else {
		errorString += LogMessage(fmt.Sprintf(" ....... Command not found in test : %s", parserCommand.Id))
	}

	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}

func createExecCommand() *schema.ExecCommand {

	LogMessage("Create an exec command :")

	execCommand := schema.ExecCommand{}

	setExecCommandValues(&execCommand)

	return &execCommand

}

func setExecCommandValues(execCommand *schema.ExecCommand) {

	execCommand.Component = GetRandomString(8, false)
	LogMessage(fmt.Sprintf("   ....... component: %s", execCommand.Component))

	execCommand.CommandLine = GetRandomString(4, false) + " " + GetRandomString(4, false)
	LogMessage(fmt.Sprintf("   ....... commandLine: %s", execCommand.CommandLine))

	if GetRandomDecision(2, 1) {
		execCommand.Group = addGroup()
	} else {
		execCommand.Group = nil
	}

	if GetBinaryDecision() {
		execCommand.Label = GetRandomString(12, false)
		LogMessage(fmt.Sprintf("   ....... label: %s", execCommand.Label))
	} else {
		execCommand.Label = ""
	}

	if GetBinaryDecision() {
		execCommand.WorkingDir = "./tmp"
		LogMessage(fmt.Sprintf("   ....... WorkingDir: %s", execCommand.WorkingDir))
	} else {
		execCommand.WorkingDir = ""
	}

	execCommand.HotReloadCapable = GetBinaryDecision()
	LogMessage(fmt.Sprintf("   ....... HotReloadCapable: %t", execCommand.HotReloadCapable))

	if GetBinaryDecision() {
		execCommand.Env = AddEnv(GetRandomNumber(4))
	} else {
		execCommand.Env = nil
	}

}

func (devfile TestDevfile) replaceSchemaCommand(command schema.Command) {
	for i := 0; i < len(devfile.SchemaDevFile.Commands); i++ {
		if devfile.SchemaDevFile.Commands[i].Id == command.Id {
			devfile.SchemaDevFile.Commands[i] = command
			break
		}
	}
}

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

func createCompositeCommand() *schema.CompositeCommand {

	compositeCommand := schema.CompositeCommand{}

	setCompositeCommandValues(&compositeCommand)

	return &compositeCommand
}

func setCompositeCommandValues(compositeCommand *schema.CompositeCommand) {
	numCommands := GetRandomNumber(3)

	compositeCommand.Commands = make([]string, numCommands)
	for i := 0; i < numCommands; i++ {
		compositeCommand.Commands[i] = GetRandomUniqueString(8, false)
		LogMessage(fmt.Sprintf("   ....... command %d of %d : %s", i, numCommands, compositeCommand.Commands[i]))
	}

	if GetRandomDecision(2, 1) {
		compositeCommand.Group = addGroup()
	}

	if GetBinaryDecision() {
		compositeCommand.Label = GetRandomString(12, false)
		LogMessage(fmt.Sprintf("   ....... label: %s", compositeCommand.Label))
	}

	if GetBinaryDecision() {
		compositeCommand.Parallel = true
		LogMessage(fmt.Sprintf("   ....... Parallel: %t", compositeCommand.Parallel))
	}
}

func (devfile TestDevfile) VerifyCommands(parserCommands []schema.Command) error {

	LogMessage("Enter VerifyCommands")
	errorString := ""

	// Compare entire array of commands
	if !cmp.Equal(parserCommands, devfile.SchemaDevFile.Commands) {
		errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command array compare failed."))
		// Array compare failed. Narrow down by comparing indivdual commands
		for _, command := range parserCommands {
			if testCommand, found := getSchemaCommand(devfile.SchemaDevFile.Commands, command.Id); found {
				if !cmp.Equal(command, *testCommand) {
					parserFilename := AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Parser")
					testFilename := AddSuffixToFileName(devfile.FileName, "_"+command.Id+"_Test")
					LogMessage(fmt.Sprintf("   .......marshall and write devfile %s", devfile.FileName))
					c, err := yaml.Marshal(command)
					if err == nil {
						err = ioutil.WriteFile(parserFilename, c, 0644)
					}
					LogMessage(fmt.Sprintf("   .......marshall and write devfile %s", devfile.FileName))
					c, err = yaml.Marshal(testCommand)
					if err == nil {
						err = ioutil.WriteFile(testFilename, c, 0644)
					}
					errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command %s did not match, see files : %s and %s", command.Id, parserFilename, testFilename))
				} else {
					LogMessage(fmt.Sprintf(" --> Command  matched : %s", command.Id))
				}
			} else {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command from parser not known to test - id : %s ", command.Id))
			}

		}
		for _, command := range devfile.SchemaDevFile.Commands {
			if _, found := getSchemaCommand(parserCommands, command.Id); !found {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command from test not returned by parser : %s ", command.Id))
			}
		}
	} else {
		LogMessage(fmt.Sprintf(" --> Command structures matched"))
	}

	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}
