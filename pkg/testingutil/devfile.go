package testingutil

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
	// versionsCommon "github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components   []common.DevfileComponent
	ExecCommands []v1.ExecCommand
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []common.DevfileComponent {
	var components = []common.DevfileComponent{}
	for _, comp := range d.Components {
		if comp.Container != nil {
			if comp.Container.Name != "" {
				components = append(components, comp)
			}
		}
	}
	return components
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() v1.WorkspaceEvents {
	return v1.WorkspaceEvents{}
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() common.DevfileMetadata {
	return common.DevfileMetadata{}
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() v1.Parent {
	return v1.Parent{}
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
// func (d TestDevfileData) GetAliasedComponents() []common.DevfileComponent {
// 	var aliasedComponents = []common.DevfileComponent{}

// 	for _, comp := range d.Components {
// 		if comp.Container != nil {
// 			if comp.Container.Name != "" {
// 				aliasedComponents = append(aliasedComponents, comp)
// 			}
// 		}
// 	}
// 	return aliasedComponents

// }

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []v1.Project {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"/test-project", "/anotherproject"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}

	project1 := v1.Project{
		ClonePath: clonePath[0],
		Name:      projectName[0],
		ProjectSource: v1.ProjectSource{
			Git: &v1.GitProjectSource{
				GitLikeProjectSource: v1.GitLikeProjectSource{
					Remotes: map[string]string{
						"origin": sourceLocation[0],
					},
				},
			},
		},
	}

	project2 := v1.Project{
		ClonePath: clonePath[1],
		Name:      projectName[1],
		ProjectSource: v1.ProjectSource{
			Git: &v1.GitProjectSource{
				GitLikeProjectSource: v1.GitLikeProjectSource{
					Remotes: map[string]string{
						"origin": sourceLocation[1],
					},
				},
			},
		},
	}
	return []v1.Project{project1, project2}

}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands() []v1.Command {

	var commands []v1.Command

	for i := range d.ExecCommands {
		commands = append(commands, v1.Command{Exec: &d.ExecCommands[i]})
	}

	return commands

}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}

// GetFakeComponent returns fake component for testing
func GetFakeComponent(name string) common.DevfileComponent {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"

	return common.DevfileComponent{
		Container: &v1.Container{
			Name:        name,
			Image:       image,
			Env:         []v1.EnvVar{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []v1.VolumeMount{{
				Name: volumeName,
				Path: volumePath,
			}},
			MountSources: true,
		}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []v1.ExecCommand {
	return []v1.ExecCommand{
		{
			CommandLine: "ls -a",
			Component:   "alias1",
			LabeledCommand: v1.LabeledCommand{
				BaseCommand: v1.BaseCommand{
					Group: &v1.CommandGroup{
						Kind: v1.RunCommandGroupKind,
					},
				},
			},
			WorkingDir: "/root",
		},
	}
}
