package v2

import (
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetCommands returns the slice of Command objects parsed from the Devfile
func (d *DevfileV2) GetCommands(options common.DevfileOptions) []v1.Command {
	if len(options.Filter) == 0 {
		return d.Commands
	}

	var commands []v1.Command
	for _, command := range d.Commands {
		filterIn, _ := common.FilterDevfileObject(command.Attributes, options)

		if filterIn {
			command.Id = strings.ToLower(command.Id)
			commands = append(commands, command)
		}
	}

	return commands
}

// AddCommands adds the slice of Command objects to the Devfile's commands
// if a command is already defined, error out
func (d *DevfileV2) AddCommands(commands ...v1.Command) error {
	devfileCommands := d.GetCommands(common.DevfileOptions{})

	for _, command := range commands {
		for _, devfileCommand := range devfileCommands {
			if command.Id == devfileCommand.Id {
				return &common.FieldAlreadyExistError{Name: command.Id, Field: "command"}
			}
		}
		d.Commands = append(d.Commands, command)
	}
	return nil
}

// UpdateCommand updates the command with the given id
func (d *DevfileV2) UpdateCommand(command v1.Command) {
	for i := range d.Commands {
		if strings.ToLower(d.Commands[i].Id) == strings.ToLower(command.Id) {
			d.Commands[i] = command
			d.Commands[i].Id = strings.ToLower(d.Commands[i].Id)
		}
	}
}
