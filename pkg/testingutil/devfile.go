package testingutil

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetFakeContainerComponent returns a fake container component for testing.
// Deprecated: use GenerateDummyContainerComponent instead
func GetFakeContainerComponent(name string) v1.Component {
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"
	VolumeMounts := []v1.VolumeMount{
		{
			Name: volumeName,
			Path: volumePath,
		},
	}
	return GenerateDummyContainerComponent(name, VolumeMounts, nil, []v1.EnvVar{}, v1.Annotation{}, nil)
}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) v1.Component {

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Volume: &v1.VolumeComponent{
				Volume: v1.Volume{
					Size: size,
				},
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

// GetFakeEnv returns a fake env for testing
func GetFakeEnv(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

// GetFakeEnvParentOverride returns a fake envParentOverride for testing
func GetFakeEnvParentOverride(name, value string) v1.EnvVarParentOverride {
	return v1.EnvVarParentOverride{
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

// GetFakeVolumeMountParentOverride returns a fake volumeMountParentOverride for testing
func GetFakeVolumeMountParentOverride(name, path string) v1.VolumeMountParentOverride {
	return v1.VolumeMountParentOverride{
		Name: name,
		Path: path,
	}
}

// GenerateDummyContainerComponent returns a dummy container component for testing
func GenerateDummyContainerComponent(name string, volMounts []v1.VolumeMount, endpoints []v1.Endpoint, envs []v1.EnvVar, annotation v1.Annotation, dedicatedPod *bool) v1.Component {
	image := "docker.io/maven:latest"
	mountSources := true

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Container: &v1.ContainerComponent{
				Container: v1.Container{
					Image:        image,
					Annotation:   annotation,
					Env:          envs,
					VolumeMounts: volMounts,
					MountSources: &mountSources,
					DedicatedPod: dedicatedPod,
				},
				Endpoints: endpoints,
			}}}
}
