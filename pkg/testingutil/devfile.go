package testingutil

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components        []v1.Component
	ExecCommands      []v1.ExecCommand
	CompositeCommands []v1.CompositeCommand
	Commands          []v1.Command
	Events            v1.Events
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []v1.Component {
	return d.Components
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() common.DevfileMetadata {
	return common.DevfileMetadata{}
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() v1.Events {
	return d.Events
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() *v1.Parent {
	return &v1.Parent{}
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []v1.Component {
	var aliasedComponents = []v1.Component{}

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
	if d.Commands == nil {
		var commands []v1.Command

		for i := range d.ExecCommands {
			commands = append(commands, v1.Command{Exec: &d.ExecCommands[i]})
		}

		for i := range d.CompositeCommands {
			commands = append(commands, v1.Command{Composite: &d.CompositeCommands[i]})
		}

		return commands
	} else {
		return d.Commands
	}
}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}

// SetMetadata sets metadata for devfile
func (d TestDevfileData) SetMetadata(name, version string) {}

// SetSchemaVersion sets schema version for devfile
func (d TestDevfileData) SetSchemaVersion(version string) {}

func (d TestDevfileData) AddComponents(components []v1.Component) error { return nil }

func (d TestDevfileData) UpdateComponent(component v1.Component) {}

func (d TestDevfileData) AddCommands(commands []v1.Command) error { return nil }

func (d TestDevfileData) UpdateCommand(command v1.Command) {}

func (d TestDevfileData) SetEvents(events v1.Events) {}

func (d TestDevfileData) AddProjects(projects []v1.Project) error { return nil }

func (d TestDevfileData) UpdateProject(project v1.Project) {}

func (d TestDevfileData) AddEvents(events v1.Events) error { return nil }

func (d TestDevfileData) UpdateEvents(postStart, postStop, preStart, preStop []string) {}

func (d TestDevfileData) SetParent(parent *v1.Parent) {}

// GetFakeContainerComponent returns a fake container component for testing
func GetFakeContainerComponent(name string) v1.Component {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"

	return v1.Component{
		Container: &v1.ContainerComponent{
			Container: v1.Container{
				Name:        name,
				Image:       image,
				Env:         []v1.EnvVar{},
				MemoryLimit: memoryLimit,
				VolumeMounts: []v1.VolumeMount{
					{
						Name: volumeName,
						Path: volumePath,
					},
				},
				MountSources: true,
			},
		},
	}
}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name string) v1.Component {
	size := "4Gi"

	return v1.Component{
		Volume: &v1.VolumeComponent{
			Volume: v1.Volume{
				Name: name,
				Size: size,
			},
		},
	}

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

// GetFakeExecRunCommands returns a fake env for testing
func GetFakeEnv(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

// GetFakeVolumeMount returns a fake volume mount for testing
func GetFakeVolumeMount(name, path string) v1.VolumeMount {
	return v1.VolumeMount{
		Name: name,
		Path: path,
	}
}
