package testingutil

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

var (
	isFalse = false
	isTrue  = true
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
					Annotation:   &annotation,
					Env:          envs,
					VolumeMounts: volMounts,
					MountSources: &mountSources,
					DedicatedPod: dedicatedPod,
				},
				Endpoints: endpoints,
			}}}
}

//DockerImageValues struct can be used to set override or main component struct values
type DockerImageValues struct {
	//maps to Image.ImageName
	ImageName string
	//maps to Image.Dockerfile.DockerfileSrc.Uri
	Uri string
	//maps to Image.Dockerfile.BuildContext
	BuildContext string
	//maps to Image.Dockerfile.RootRequired
	RootRequired *bool
}

//GetDockerImageTestComponent returns a docker image component that is used for testing.
//The parameters allow customization of the content.  If they are set to nil, then the properties will not be set
func GetDockerImageTestComponent(div DockerImageValues, attr attributes.Attributes) v1.Component {
	comp := v1.Component{
		Name: "image",
		ComponentUnion: v1.ComponentUnion{
			Image: &v1.ImageComponent{
				Image: v1.Image{
					ImageName: div.ImageName,
					ImageUnion: v1.ImageUnion{
						Dockerfile: &v1.DockerfileImage{
							DockerfileSrc: v1.DockerfileSrc{
								Uri: div.Uri,
							},
							Dockerfile: v1.Dockerfile{
								BuildContext: div.BuildContext,
							},
						},
					},
				},
			},
		},
	}

	if div.RootRequired != nil {
		comp.Image.Dockerfile.RootRequired = div.RootRequired
	}

	if attr != nil {
		comp.Attributes = attr
	}

	return comp
}

//GetDockerImageTestComponentParentOverride returns a docker image parent override component that is used for testing.
//The parameters allow customization of the content.  If they are set to nil, then the properties will not be set
func GetDockerImageTestComponentParentOverride(div DockerImageValues) v1.ComponentParentOverride {
	comp := v1.ComponentParentOverride{
		Name: "image",
		ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
			Image: &v1.ImageComponentParentOverride{
				ImageParentOverride: v1.ImageParentOverride{
					ImageName: div.ImageName,
					ImageUnionParentOverride: v1.ImageUnionParentOverride{
						Dockerfile: &v1.DockerfileImageParentOverride{
							DockerfileSrcParentOverride: v1.DockerfileSrcParentOverride{
								Uri: div.Uri,
							},
							DockerfileParentOverride: v1.DockerfileParentOverride{
								BuildContext: div.BuildContext,
							},
						},
					},
				},
			},
		},
	}

	if div.RootRequired != nil {
		comp.Image.Dockerfile.RootRequired = div.RootRequired
	}

	return comp
}

//GetDockerImageTestComponentPluginOverride returns a docker image parent override component that is used for testing.
//The parameters allow customization of the content.  If they are set to nil, then the properties will not be set
func GetDockerImageTestComponentPluginOverride(div DockerImageValues) v1.ComponentPluginOverride {
	comp := v1.ComponentPluginOverride{
		Name: "image",
		ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
			Image: &v1.ImageComponentPluginOverride{
				ImagePluginOverride: v1.ImagePluginOverride{
					ImageName: div.ImageName,
					ImageUnionPluginOverride: v1.ImageUnionPluginOverride{
						Dockerfile: &v1.DockerfileImagePluginOverride{
							DockerfileSrcPluginOverride: v1.DockerfileSrcPluginOverride{
								Uri: div.Uri,
							},
							DockerfilePluginOverride: v1.DockerfilePluginOverride{
								BuildContext: div.BuildContext,
							},
						},
					},
				},
			},
		},
	}

	if div.RootRequired != nil {
		comp.Image.Dockerfile.RootRequired = div.RootRequired
	}

	return comp
}

// GetUnsetBooleanDevfileObj returns a DevfileData object that contains unset boolean properties
func GetUnsetBooleanDevfileTestData(apiVersion string) (devfileData data.DevfileData, err error) {
	devfileData = &v2.DevfileV2{
		Devfile: v1.Devfile{
			DevfileHeader: devfilepkg.DevfileHeader{
				SchemaVersion: apiVersion,
			},
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Commands: []v1.Command{
						{
							Id: "devrun",
							CommandUnion: v1.CommandUnion{
								Exec: &v1.ExecCommand{
									CommandLine: "npm run",
									WorkingDir:  "/projects/nodejs-starter",
								},
							},
						},
						{
							Id: "testrun",
							CommandUnion: v1.CommandUnion{
								Apply: &v1.ApplyCommand{
									LabeledCommand: v1.LabeledCommand{
										BaseCommand: v1.BaseCommand{
											Group: &v1.CommandGroup{
												Kind: v1.BuildCommandGroupKind,
											},
										},
									},
								},
							},
						},
						{
							Id: "allcmds",
							CommandUnion: v1.CommandUnion{
								Composite: &v1.CompositeCommand{
									Commands: []string{"testrun", "devrun"},
								},
							},
						},
					},
					Components: []v1.Component{
						{
							Name: "nodejs",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Annotation: &v1.Annotation{
											Deployment: map[string]string{
												"deploy-key1": "deploy-value1",
											},
											Service: map[string]string{
												"svc-key1": "svc-value1",
												"svc-key2": "svc-value3",
											},
										},
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []v1.Endpoint{
										{
											Name:       "log",
											TargetPort: 443,
											Annotations: map[string]string{
												"ingress-key1": "ingress-value1",
												"ingress-key2": "ingress-value3",
											},
										},
									},
								},
							},
						},
						GetFakeVolumeComponent("volume", "2Gi"),
						{
							Name: "openshift",
							ComponentUnion: v1.ComponentUnion{
								Openshift: &v1.OpenshiftComponent{
									K8sLikeComponent: v1.K8sLikeComponent{
										K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
											Uri: "https://xyz.com/dir/file.yaml",
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "metrics",
												TargetPort: 8080,
											},
										},
									},
								},
							},
						},
					},
					Events: &v1.Events{
						DevWorkspaceEvents: v1.DevWorkspaceEvents{
							PostStart: []string{"post-start-0"},
							PostStop:  []string{"post-stop"},
							PreStop:   []string{},
							PreStart:  []string{},
						},
					},
				},
			},
		},
	}

	if apiVersion != string(data.APISchemaVersion200) && apiVersion != string(data.APISchemaVersion210) {
		comp := []v1.Component{GetDockerImageTestComponent(DockerImageValues{}, nil)}
		err = devfileData.AddComponents(comp)
	}

	return devfileData, err

}

//GetBooleanDevfileTestData returns a DevfileData object that contains set values for the boolean properties.  If setDefault is true, an object with the default boolean values will be returned
func GetBooleanDevfileTestData(apiVersion string, setDefault bool) (devfileData data.DevfileData, err error) {

	type boolValues struct {
		hotReloadCapable *bool
		secure           *bool
		parallel         *bool
		dedicatedPod     *bool
		mountSources     *bool
		isDefault        *bool
		rootRequired     *bool
		ephemeral        *bool
	}

	//default values according to spec
	defaultBools := boolValues{&isFalse, &isFalse, &isFalse, &isFalse, &isTrue, &isFalse, &isFalse, &isFalse}
	//set values will be a mix of default and inverse values
	setBools := boolValues{&isTrue, &isTrue, &isFalse, &isTrue, &isFalse, &isFalse, &isTrue, &isFalse}

	var values boolValues

	if setDefault {
		values = defaultBools
	} else {
		values = setBools
	}

	devfileData = &v2.DevfileV2{
		Devfile: v1.Devfile{
			DevfileHeader: devfilepkg.DevfileHeader{
				SchemaVersion: apiVersion,
			},
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Commands: []v1.Command{
						{
							Id: "devrun",
							CommandUnion: v1.CommandUnion{
								Exec: &v1.ExecCommand{
									CommandLine:      "npm run",
									WorkingDir:       "/projects/nodejs-starter",
									HotReloadCapable: values.hotReloadCapable,
								},
							},
						},
						{
							Id: "testrun",
							CommandUnion: v1.CommandUnion{
								Apply: &v1.ApplyCommand{
									LabeledCommand: v1.LabeledCommand{
										BaseCommand: v1.BaseCommand{
											Group: &v1.CommandGroup{
												Kind:      v1.BuildCommandGroupKind,
												IsDefault: values.isDefault,
											},
										},
									},
								},
							},
						},
						{
							Id: "allcmds",
							CommandUnion: v1.CommandUnion{
								Composite: &v1.CompositeCommand{
									Commands: []string{"testrun", "devrun"},
									Parallel: values.parallel,
								},
							},
						},
					},
					Components: []v1.Component{
						{
							Name: "nodejs",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Annotation: &v1.Annotation{
											Deployment: map[string]string{
												"deploy-key1": "deploy-value1",
											},
											Service: map[string]string{
												"svc-key1": "svc-value1",
												"svc-key2": "svc-value3",
											},
										},
										Image:        "quay.io/nodejs-12",
										DedicatedPod: values.dedicatedPod,
										MountSources: values.mountSources,
									},
									Endpoints: []v1.Endpoint{
										{
											Name:       "log",
											TargetPort: 443,
											Annotations: map[string]string{
												"ingress-key1": "ingress-value1",
												"ingress-key2": "ingress-value3",
											},
											Secure: values.secure,
										},
									},
								},
							},
						},
						GetFakeVolumeComponent("volume", "2Gi"),
						{
							Name: "openshift",
							ComponentUnion: v1.ComponentUnion{
								Openshift: &v1.OpenshiftComponent{
									K8sLikeComponent: v1.K8sLikeComponent{
										K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
											Uri: "https://xyz.com/dir/file.yaml",
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "metrics",
												TargetPort: 8080,
												Secure:     values.secure,
											},
										},
									},
								},
							},
						},
					},
					Events: &v1.Events{
						DevWorkspaceEvents: v1.DevWorkspaceEvents{
							PostStart: []string{"post-start-0"},
							PostStop:  []string{"post-stop"},
							PreStop:   []string{},
							PreStart:  []string{},
						},
					},
				},
			},
		},
	}

	if apiVersion != string(data.APISchemaVersion200) {
		volComponent, _ := devfileData.GetComponents(common.DevfileOptions{ComponentOptions: common.ComponentOptions{
			ComponentType: v1.VolumeComponentType,
		}})

		volComponent[0].Volume.Ephemeral = values.ephemeral
	}

	if apiVersion != string(data.APISchemaVersion200) && apiVersion != string(data.APISchemaVersion210) {
		comp := []v1.Component{GetDockerImageTestComponent(DockerImageValues{RootRequired: values.rootRequired}, nil)}
		err = devfileData.AddComponents(comp)
	}

	return devfileData, err
}
