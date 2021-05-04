package generator

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/pkg/devfile/parser"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
)

var fakeResources corev1.ResourceRequirements

func init() {
	fakeResources, _ = testingutil.FakeResourceRequirements("0.5m", "300Mi")
}

func TestGetContainers(t *testing.T) {

	containerNames := []string{"testcontainer1", "testcontainer2"}
	containerImages := []string{"image1", "image2"}
	trueMountSources := true
	falseMountSources := false

	project := v1.Project{
		ClonePath: "test-project/",
		Name:      "project0",
		ProjectSource: v1.ProjectSource{
			Git: &v1.GitProjectSource{
				GitLikeProjectSource: v1.GitLikeProjectSource{
					Remotes: map[string]string{
						"origin": "repo",
					},
				},
			},
		},
	}

	tests := []struct {
		name                  string
		containerComponents   []v1.Component
		filterOptions         common.DevfileOptions
		wantContainerName     string
		wantContainerImage    string
		wantContainerEnv      []corev1.EnvVar
		wantContainerVolMount []corev1.VolumeMount
		wantErr               bool
	}{
		{
			name: "Container with default project root",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &trueMountSources,
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{

				{
					Name:  "PROJECTS_ROOT",
					Value: "/projects",
				},
				{
					Name:  "PROJECT_SOURCE",
					Value: "/projects/test-project",
				},
			},
			wantContainerVolMount: []corev1.VolumeMount{
				{
					Name:      "devfile-projects",
					MountPath: "/projects",
				},
			},
		},
		{
			name: "Container with source mapping",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:         containerImages[0],
								MountSources:  &trueMountSources,
								SourceMapping: "/myroot",
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{

				{
					Name:  "PROJECTS_ROOT",
					Value: "/myroot",
				},
				{
					Name:  "PROJECT_SOURCE",
					Value: "/myroot/test-project",
				},
			},
			wantContainerVolMount: []corev1.VolumeMount{
				{
					Name:      "devfile-projects",
					MountPath: "/myroot",
				},
			},
		},
		{
			name: "Container with no mount source",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &falseMountSources,
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
		},
		{
			name: "Filter containers",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &falseMountSources,
							},
						},
					},
				},
				{
					Name: containerNames[1],
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &falseMountSources,
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[1],
			wantContainerImage: containerImages[0],
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: tt.containerComponents,
								Projects: []v1.Project{
									project,
								},
							},
						},
					},
				},
			}

			containers, err := GetContainers(devObj, tt.filterOptions)
			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContainers() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				for _, container := range containers {
					if container.Name != tt.wantContainerName {
						t.Errorf("TestGetContainers error: Name mismatch - got: %s, wanted: %s", container.Name, tt.wantContainerName)
					}
					if container.Image != tt.wantContainerImage {
						t.Errorf("TestGetContainers error: Image mismatch - got: %s, wanted: %s", container.Image, tt.wantContainerImage)
					}
					if len(container.Env) > 0 && !reflect.DeepEqual(container.Env, tt.wantContainerEnv) {
						t.Errorf("TestGetContainers error: Env mismatch - got: %+v, wanted: %+v", container.Env, tt.wantContainerEnv)
					}
					if len(container.VolumeMounts) > 0 && !reflect.DeepEqual(container.VolumeMounts, tt.wantContainerVolMount) {
						t.Errorf("TestGetContainers error: Vol Mount mismatch - got: %+v, wanted: %+v", container.VolumeMounts, tt.wantContainerVolMount)
					}
				}
			}
		})
	}

}

func TestGetVolumesAndVolumeMounts(t *testing.T) {

	type testVolumeMountInfo struct {
		mountPath  string
		volumeName string
	}

	tests := []struct {
		name                string
		components          []v1.Component
		volumeNameToVolInfo map[string]VolumeInfo
		wantContainerToVol  map[string][]testVolumeMountInfo
		wantErr             bool
	}{
		{
			name:       "One volume mounted",
			components: []v1.Component{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"myvolume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"comp1": {
					{
						mountPath:  "/my/volume/mount/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
				"comp2": {
					{
						mountPath:  "/my/volume/mount/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One volume mounted at diff locations",
			components: []v1.Component{
				{
					Name: "container1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path1",
									},
									{
										Name: "volume1",
										Path: "/path2",
									},
								},
							},
						},
					},
				},
			},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"volume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"container1": {
					{
						mountPath:  "/path1",
						volumeName: "volume1-pvc-vol",
					},
					{
						mountPath:  "/path2",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "One volume mounted at diff container components",
			components: []v1.Component{
				{
					Name: "container1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path1",
									},
								},
							},
						},
					},
				},
				{
					Name: "container2",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								VolumeMounts: []v1.VolumeMount{
									{
										Name: "volume1",
										Path: "/path2",
									},
								},
							},
						},
					},
				},
			},
			volumeNameToVolInfo: map[string]VolumeInfo{
				"volume1": {
					PVCName:    "volume1-pvc",
					VolumeName: "volume1-pvc-vol",
				},
			},
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"container1": {
					{
						mountPath:  "/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
				"container2": {
					{
						mountPath:  "/path2",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid case",
			components: []v1.Component{
				{
					Name: "container1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := parser.DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: tt.components,
							},
						},
					},
				},
			}

			containers, err := GetContainers(devObj, common.DevfileOptions{})
			if err != nil {
				t.Errorf("TestGetVolumesAndVolumeMounts error - %v", err)
				return
			}

			var options common.DevfileOptions
			if tt.wantErr {
				options = common.DevfileOptions{
					Filter: map[string]interface{}{
						"firstString": "firstStringValue",
					},
				}
			}

			volumeParams := VolumeParams{
				Containers:             containers,
				VolumeNameToVolumeInfo: tt.volumeNameToVolInfo,
			}

			pvcVols, err := GetVolumesAndVolumeMounts(devObj, volumeParams, options)
			if tt.wantErr == (err == nil) {
				t.Errorf("TestGetVolumesAndVolumeMounts() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				// check if the pvc volumes returned are correct
				for _, volInfo := range tt.volumeNameToVolInfo {
					matched := false
					for _, pvcVol := range pvcVols {
						if volInfo.VolumeName == pvcVol.Name && pvcVol.PersistentVolumeClaim != nil && volInfo.PVCName == pvcVol.PersistentVolumeClaim.ClaimName {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestGetVolumesAndVolumeMounts error - could not find volume details %s in the actual result", volInfo.VolumeName)
					}
				}

				// check the volume mounts of the containers
				for _, container := range containers {
					if volMounts, ok := tt.wantContainerToVol[container.Name]; !ok {
						t.Errorf("TestGetVolumesAndVolumeMounts error - did not find the expected container %s", container.Name)
						return
					} else {
						for _, expectedVolMount := range volMounts {
							matched := false
							for _, actualVolMount := range container.VolumeMounts {
								if expectedVolMount.volumeName == actualVolMount.Name && expectedVolMount.mountPath == actualVolMount.MountPath {
									matched = true
								}
							}

							if !matched {
								t.Errorf("TestGetVolumesAndVolumeMounts error - could not find volume mount details for path %s in the actual result for container %s", expectedVolMount.mountPath, container.Name)
							}
						}
					}
				}
			}
		})
	}
}

func TestGetVolumeMountPath(t *testing.T) {

	tests := []struct {
		name        string
		volumeMount v1.VolumeMount
		wantPath    string
	}{
		{
			name: "Mount Path is present",
			volumeMount: v1.VolumeMount{
				Name: "name1",
				Path: "/path1",
			},
			wantPath: "/path1",
		},
		{
			name: "Mount Path is absent",
			volumeMount: v1.VolumeMount{
				Name: "name1",
			},
			wantPath: "/name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := GetVolumeMountPath(tt.volumeMount)

			if path != tt.wantPath {
				t.Errorf("TestGetVolumeMountPath error: mount path mismatch, expected: %v got: %v", tt.wantPath, path)
			}
		})
	}

}
