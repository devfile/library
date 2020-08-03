package version200

import (
	"strings"

	apiComp "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// GetComponents returns the slice of DevfileComponent objects parsed from the Devfile
func (d *Devfile200) GetComponents() []common.DevfileComponent {
	return d.Components
}

// GetCommands returns the slice of DevfileCommand objects parsed from the Devfile
func (d *Devfile200) GetCommands() []common.DevfileCommand {
	var commands []common.DevfileCommand

	for _, command := range d.Commands {
		// we convert devfile command id to lowercase so that we can handle
		// cases efficiently without being error prone
		// we also convert the odo push commands from build-command and run-command flags
		command.Exec.Id = strings.ToLower(command.Exec.Id)
		commands = append(commands, command)
	}

	return commands
}

// GetParent returns the  DevfileParent object parsed from devfile
func (d *Devfile200) GetParent() apiComp.Parent {
	return d.Parent
}

// GetProjects returns the DevfileProject Object parsed from devfile
func (d *Devfile200) GetProjects() []apiComp.Project {
	return d.Projects
}

// GetMetadata returns the DevfileMetadata Object parsed from devfile
func (d *Devfile200) GetMetadata() common.DevfileMetadata {
	return d.Metadata
}

// GetEvents returns the Events Object parsed from devfile
func (d *Devfile200) GetEvents() apiComp.WorkspaceEvents {
	return d.Events
}

// GetAliasedComponents returns the slice of DevfileComponent objects that each have an alias
func (d *Devfile200) GetAliasedComponents() []common.DevfileComponent {
	// V2 has name required in jsonSchema
	return d.Components
}
