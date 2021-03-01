package v2

import (
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// GetCommands returns the slice of Command objects parsed from the Devfile
func (d *DevfileV2) GetCommands(options common.DevfileOptions) ([]v1.Command, error) {
	if len(options.Filter) == 0 {
		return d.Commands, nil
	}

	var commands []v1.Command
	for _, command := range d.Commands {
		filterIn, err := common.FilterDevfileObject(command.Attributes, options)
		if err != nil {
			return nil, err
		}

		if filterIn {
			command.Id = strings.ToLower(command.Id)
			commands = append(commands, command)
		}
	}

	return commands, nil
}

// AddCommands adds the slice of Command objects to the Devfile's commands
// if a command is already defined, error out
func (d *DevfileV2) AddCommands(commands []v1.Command) error {

	for _, command := range commands {
		for _, devfileCommand := range d.Commands {
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

// DeleteCommand removes the specified command
func (d *DevfileV2) DeleteCommand(id string) error {

	found := false
	for i := len(d.Commands) - 1; i >= 0; i-- {
		if d.Commands[i].Composite != nil && d.Commands[i].Id != id {
			var subCmd []string
			for _, command := range d.Commands[i].Composite.Commands {
				if command != id {
					subCmd = append(subCmd, command)
				}
			}
			d.Commands[i].Composite.Commands = subCmd
		} else if d.Commands[i].Id == id {
			found = true
			d.Commands = append(d.Commands[:i], d.Commands[i+1:]...)
		}
	}

	if !found {
		return &common.FieldNotFoundError{
			Field: "command",
			Name:  id,
		}
	}

	return nil
}
