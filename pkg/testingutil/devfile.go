package testingutil

import (
	apiComp "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
	// versionsCommon "github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components   []common.DevfileComponent
	ExecCommands []apiComp.ExecCommand
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []common.DevfileComponent {
	return d.GetAliasedComponents()
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() apiComp.WorkspaceEvents {
	return apiComp.WorkspaceEvents{}
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() common.DevfileMetadata {
	return common.DevfileMetadata{}
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() apiComp.Parent {
	return apiComp.Parent{}
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []common.DevfileComponent {
	var aliasedComponents = []common.DevfileComponent{}

	for _, comp := range d.Components {
		if comp.Container != nil {
			if comp.Container.Name != "" {
				aliasedComponents = append(aliasedComponents, comp)
			}
		}
	}
	return aliasedComponents

}

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []apiComp.Project {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"/test-project", "/anotherproject"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}

	project1 := apiComp.Project{
		ClonePath: clonePath[0],
		Name:      projectName[0],
		ProjectSource: apiComp.ProjectSource{
			Git: &apiComp.GitProjectSource{
				GitLikeProjectSource: apiComp.GitLikeProjectSource{
					CommonProjectSource: apiComp.CommonProjectSource{
						Location: sourceLocation[0],
					},
				},
			},
		},
	}

	project2 := apiComp.Project{
		ClonePath: clonePath[1],
		Name:      projectName[1],
		ProjectSource: apiComp.ProjectSource{
			Git: &apiComp.GitProjectSource{
				GitLikeProjectSource: apiComp.GitLikeProjectSource{
					CommonProjectSource: apiComp.CommonProjectSource{
						Location: sourceLocation[1],
					},
				},
			},
		},
	}
	return []apiComp.Project{project1, project2}

}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands() []common.DevfileCommand {

	var commands []common.DevfileCommand

	for i := range d.ExecCommands {
		commands = append(commands, common.DevfileCommand{Exec: &d.ExecCommands[i]})
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
		Container: &apiComp.Container{
			Name:        name,
			Image:       image,
			Env:         []apiComp.EnvVar{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []apiComp.VolumeMount{{
				Name: volumeName,
				Path: volumePath,
			}},
			MountSources: true,
		}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []apiComp.ExecCommand {
	return []apiComp.ExecCommand{
		{
			CommandLine: "ls -a",
			Component:   "alias1",
			LabeledCommand: apiComp.LabeledCommand{
				BaseCommand: apiComp.BaseCommand{
					Group: &apiComp.CommandGroup{
						Kind: apiComp.RunCommandGroupKind,
					},
				},
			},
			WorkingDir: "/root",
		},
	}
}
