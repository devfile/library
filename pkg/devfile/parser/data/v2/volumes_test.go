package v2

import (
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/stretchr/testify/assert"
)

func TestDevfile200_AddVolumeMount(t *testing.T) {
	image0 := "some-image-0"

	container0 := "container0"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	type args struct {
		componentName string
		name          string
		path          string
	}
	tests := []struct {
		name              string
		currentComponents []v1.Component
		wantComponents    []v1.Component
		args              args
		wantErr           bool
	}{
		{
			name: "add the volume mount when other mounts are present",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
			args: args{
				name:          volume0,
				path:          "/path0",
				componentName: container0,
			},
			wantComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
									testingutil.GetFakeVolumeMount(volume0, "/path0"),
								},
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "error out when same path is present in the container",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
			args: args{
				name:          volume0,
				path:          "/data",
				componentName: container0,
			},
			wantErr: true,
		},
		{
			name: "error out when the specified container is not found",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
			},
			args: args{
				name:          volume0,
				path:          "/data",
				componentName: container1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.currentComponents,
						},
					},
				},
			}

			err := d.AddVolumeMount(tt.args.componentName, tt.args.name, tt.args.path)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error from test but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Got unexpected error: %s", err)
			} else if err == nil {
				assert.Equal(t, tt.wantComponents, d.Components, "The two values should be the same.")
			}
		})
	}
}

func TestDevfile200_DeleteVolumeMounts(t *testing.T) {

	tests := []struct {
		name             string
		volMountToDelete string
		components       []v1.Component
		wantComponents   []v1.Component
		wantErr          bool
	}{
		{
			name:             "Volume Component with mounts",
			volMountToDelete: "comp2",
			components: []v1.Component{
				{
					Name: "comp1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp2", "/path"),
									testingutil.GetFakeVolumeMount("comp2", "/path2"),
									testingutil.GetFakeVolumeMount("comp3", "/path"),
								},
							},
						},
					},
				},
				{
					Name: "comp4",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp2", "/path"),
								},
							},
						},
					},
				},
			},
			wantComponents: []v1.Component{
				{
					Name: "comp1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp3", "/path"),
								},
							},
						},
					},
				},
				{
					Name: "comp4",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:             "Missing mount name",
			volMountToDelete: "comp1",
			components: []v1.Component{
				{
					Name: "comp4",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp2", "/path"),
								},
							},
						},
					},
				},
			},
			wantComponents: []v1.Component{
				{
					Name: "comp4",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount("comp2", "/path"),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.components,
						},
					},
				},
			}

			err := d.DeleteVolumeMount(tt.volMountToDelete)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error from test but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Got unexpected error: %s", err)
			} else if err == nil {
				assert.Equal(t, tt.wantComponents, d.Components, "The two values should be the same.")
			}
		})
	}

}

func TestDevfile200_GetVolumeMountPath(t *testing.T) {
	volume1 := "volume1"
	component1 := "component1"

	tests := []struct {
		name              string
		currentComponents []v1.Component
		mountName         string
		componentName     string
		wantPath          string
		wantErr           bool
	}{
		{
			name: "vol is mounted on the specified container component",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/path"),
								},
							},
						},
					},
					Name: component1,
				},
			},
			wantPath:      "/path",
			mountName:     volume1,
			componentName: component1,
			wantErr:       false,
		},
		{
			name: "vol is not mounted on the specified container component",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/path"),
								},
							},
						},
					},
					Name: component1,
				},
			},
			mountName:     "volume2",
			componentName: component1,
			wantErr:       true,
		},
		{
			name: "invalid specified container",
			currentComponents: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/path"),
								},
							},
						},
					},
					Name: component1,
				},
			},
			mountName:     volume1,
			componentName: "component2",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.currentComponents,
						},
					},
				},
			}
			gotPath, err := d.GetVolumeMountPath(tt.mountName, tt.componentName)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error from test but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Got unexpected error: %s", err)
			} else if err == nil {
				assert.Equal(t, tt.wantPath, gotPath, "The two values should be the same.")
			}
		})
	}
}
