package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/parser/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"
)

func TestDevfile200_AddVolume(t *testing.T) {
	image0 := "some-image-0"
	container0 := "container0"

	image1 := "some-image-1"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	type args struct {
		volumeComponent v1.Component
		path            string
	}
	tests := []struct {
		name              string
		currentComponents []v1.Component
		wantComponents    []v1.Component
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: it should add the volume to all the containers",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image1,
							},
						},
					},
				},
			},
			args: args{
				volumeComponent: testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				path:            "/path",
			},
			wantComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume0, "/path"),
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
								Image: image1,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume0, "/path"),
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
		},
		{
			name: "case 2: it should add the volume when other volumes are present",
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
				volumeComponent: testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				path:            "/path",
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
									testingutil.GetFakeVolumeMount(volume0, "/path"),
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
		},
		{
			name: "case 3: error out when same volume is present",
			currentComponents: []v1.Component{
				testingutil.GetFakeVolumeComponent(volume0, "1Gi"),
			},
			args: args{
				volumeComponent: testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				path:            "/path",
			},
			wantErr: true,
		},
		{
			name: "case 4: it should error out when another volume is mounted to the same path",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/path"),
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				volumeComponent: testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				path:            "/path",
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

			err := d.AddVolume(tt.args.volumeComponent, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddVolume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(d.Components, tt.wantComponents) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantComponents, d.Components, pretty.Compare(tt.wantComponents, d.Components))
			}
		})
	}
}

func TestDevfile200_DeleteVolume(t *testing.T) {
	image0 := "some-image-0"
	container0 := "container0"

	image1 := "some-image-1"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	type args struct {
		name string
	}
	tests := []struct {
		name              string
		currentComponents []v1.Component
		wantComponents    []v1.Component
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: volume is present and mounted to multiple components",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume0, "/path"),
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
								Image: image1,
								VolumeMounts: []v1.VolumeMount{
									{
										Name: volume0,
										Path: "/path",
									},
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
			wantComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
							},
						},
					},
				},
				{
					Name: container1,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image1,
							},
						},
					},
				},
			},
			args: args{
				name: volume0,
			},
			wantErr: false,
		},
		{
			name: "case 2: delete only the required volume in case of multiples",
			currentComponents: []v1.Component{
				{
					Name: container0,
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: image0,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume0, "/path"),
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
								Image: image1,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
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
								Image: image1,
								VolumeMounts: []v1.VolumeMount{
									testingutil.GetFakeVolumeMount(volume1, "/data"),
								},
							},
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				name: volume0,
			},
			wantErr: false,
		},
		{
			name: "case 3: volume is not present",
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
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			wantComponents: []v1.Component{},
			args: args{
				name: volume0,
			},
			wantErr: true,
		},
		{
			name: "case 4: volume is present but not mounted to any component",
			currentComponents: []v1.Component{
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
			wantComponents: []v1.Component{},
			args: args{
				name: volume0,
			},
			wantErr: false,
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
			err := d.DeleteVolume(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(d.Components, tt.wantComponents) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantComponents, d.Components, pretty.Compare(tt.wantComponents, d.Components))
			}
		})
	}
}

func TestDevfile200_GetVolumeMountPath(t *testing.T) {
	volume1 := "volume1"

	type args struct {
		name string
	}
	tests := []struct {
		name              string
		currentComponents []v1.Component
		wantPath          string
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: volume is present and mounted on a component",
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
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			wantPath: "/path",
			args: args{
				name: volume1,
			},
			wantErr: false,
		},
		{
			name: "case 2: volume is not present but mounted on a component",
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
				},
			},
			args: args{
				name: volume1,
			},
			wantErr: true,
		},
		{
			name:              "case 3: volume is not present and not mounted on a component",
			currentComponents: []v1.Component{},
			args: args{
				name: volume1,
			},
			wantErr: true,
		},
		{
			name: "case 4: volume is present but not mounted",
			currentComponents: []v1.Component{
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				name: volume1,
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
			got, err := d.GetVolumeMountPath(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVolumeMountPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.wantPath) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantPath, got, pretty.Compare(tt.wantPath, got))
			}
		})
	}
}
