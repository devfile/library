package tests

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

type GenericCommand struct {
	Id                     string
	Verified               bool
	CommandType            schema.CommandType
	ExecCommandSchema      *schema.ExecCommand
	SchemaCompositeCommand *schema.CompositeCommand
}

func (genericCommand *GenericCommand) setVerified() {
	genericCommand.Verified = true
}

func (genericCommand *GenericCommand) setId(id string) {
	genericCommand.Id = id
}

func (genericCommand *GenericCommand) checkId(command schema.Command) bool {
	return genericCommand.Id == command.Id
}

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

	genericCommand := GenericCommand{}
	genericCommand.CommandType = commandType

	generateCommand(&commands[index], &genericCommand)

	devfile.MapCommand(genericCommand)
	devfile.SchemaDevFile.Commands = commands

	return commands[index].Id

}

func generateCommand(command *schema.Command, genericCommand *GenericCommand) {

	command.Id = GetRandomUniqueString(8, true)
	genericCommand.Id = command.Id
	LogMessage(fmt.Sprintf("   ....... id: %s", command.Id))

	if genericCommand.CommandType == schema.ExecCommandType {
		command.Exec = createExecCommand()
		genericCommand.ExecCommandSchema = command.Exec
	} else if genericCommand.CommandType == schema.CompositeCommandType {
		command.Composite = createCompositeCommand()
		genericCommand.SchemaCompositeCommand = command.Composite
	}
}

func (devfile *TestDevfile) UpdateCommand(command *schema.Command) error {

	errorString := ""
	genericCommand := devfile.GetCommand(command.Id)
	if genericCommand != nil {
		LogMessage(fmt.Sprintf(" ....... Updating command id: %s", command.Id))
		if genericCommand.CommandType == schema.ExecCommandType {
			setExecCommandValues(command.Exec)
			genericCommand.ExecCommandSchema = command.Exec
		} else if genericCommand.CommandType == schema.CompositeCommandType {
			setCompositeCommandValues(command.Composite)
			genericCommand.SchemaCompositeCommand = command.Composite
		}
	} else {
		errorString += LogMessage(fmt.Sprintf(" ....... Command not found in test : %s", command.Id))
	}
	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}

func createExecCommand() *schema.ExecCommand {

	LogMessage("Create a composite command :")

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

func (devfile TestDevfile) VerifyCommands(commands []schema.Command) error {

	LogMessage("Enter VerifyCommands")
	errorString := ""

	if devfile.CommandMap != nil {
		for _, command := range commands {

			if matchedCommand, found := devfile.CommandMap[command.Id]; found {
				matchedCommand.setVerified()
				if matchedCommand.CommandType == schema.ExecCommandType {
					if !cmp.Equal(*command.Exec, *matchedCommand.ExecCommandSchema) {
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Exec Command %s from parser: %v", command.Id, *command.Exec))
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Exec Command %s from tester: %v", matchedCommand.Id, matchedCommand.ExecCommandSchema))
					} else {
						LogMessage(fmt.Sprintf(" --> Exec command structures matched - id : %s ", command.Id))
					}
				}
				if matchedCommand.CommandType == schema.CompositeCommandType {
					if !cmp.Equal(*command.Composite, *matchedCommand.SchemaCompositeCommand) {
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Composite Command from parser: %v", *command.Composite))
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Composite Command from tester: %v", matchedCommand.SchemaCompositeCommand))
					} else {
						LogMessage(fmt.Sprintf(" --> Composite command structures matched - id : %s ", command.Id))
					}
				}

			} else {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command from parser not known to test - id : %s ", command.Id))
			}
		}

		for _, genericCommand := range devfile.CommandMap {
			if !genericCommand.Verified {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Command not returned by parser - id : %s", genericCommand.Id))
			}
		}

	} else {
		if commands != nil {
			errorString += LogMessage(" --> ERROR: Parser returned commands but Test does not include any.")
		}
	}
	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}
