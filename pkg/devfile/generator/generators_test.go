//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/testingutil"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/devfile/library/v2/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/pod-security-admission/api"
	"k8s.io/utils/pointer"
)

var fakeResources corev1.ResourceRequirements

func init() {
	fakeResources, _ = testingutil.FakeResourceRequirements("0.5m", "300Mi")
}

func TestGetContainers(t *testing.T) {
	containerNames := []string{"testcontainer1", "testcontainer2", "testcontainer3"}
	containerImages := []string{"image1", "image2", "image3"}

	trueMountSources := true
	falseMountSources := false

	projects := []v1.Project{
		{
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
		},
	}

	applyCommands := []v1.Command{
		{
			Id: "apply1",
			CommandUnion: v1.CommandUnion{
				Apply: &v1.ApplyCommand{
					Component: containerNames[1],
				},
			},
		},
		{
			Id: "apply2",
			CommandUnion: v1.CommandUnion{
				Apply: &v1.ApplyCommand{
					Component: containerNames[2],
				},
			},
		},
	}

	errMatches := "an expected error"

	type EventCommands struct {
		preStart []string
		postStop []string
	}

	tests := []struct {
		name                      string
		eventCommands             EventCommands
		containerComponents       []v1.Component
		filteredComponents        []v1.Component
		filterOptions             common.DevfileOptions
		wantContainerName         string
		wantContainerImage        string
		wantContainerEnv          []corev1.EnvVar
		wantContainerVolMount     []corev1.VolumeMount
		wantContainerOverrideData *corev1.Container
		wantErr                   *string
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
								Env: []v1.EnvVar{
									{
										Name:  "testVar1",
										Value: "testVal1",
									},
									{
										Name:  "testVar2",
										Value: "testVal2",
									},
								},
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{
				{
					Name:  "PROJECT_SOURCE",
					Value: "/projects/test-project",
				},
				{
					Name:  "PROJECTS_ROOT",
					Value: "/projects",
				},
				{
					Name:  "testVar1",
					Value: "testVal1",
				},
				{
					Name:  "testVar2",
					Value: "testVal2",
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
								Env: []v1.EnvVar{
									{
										Name:  "testVar1",
										Value: "testVal1",
									},
									{
										Name:  "testVar2",
										Value: "testVal2",
									},
								},
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{
				{
					Name:  "PROJECT_SOURCE",
					Value: "/myroot/test-project",
				},
				{
					Name:  "PROJECTS_ROOT",
					Value: "/myroot",
				},
				{
					Name:  "testVar1",
					Value: "testVal1",
				},
				{
					Name:  "testVar2",
					Value: "testVal2",
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
			filteredComponents: []v1.Component{
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
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
		},
		{
			name: "should not return containers for preStart and postStop events",
			eventCommands: EventCommands{
				preStart: []string{"apply1"},
				postStop: []string{"apply2"},
			},
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
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[1],
								MountSources: &falseMountSources,
							},
						},
					},
				},
				{
					Name: containerNames[2],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[2],
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
			name:    "Simulating error case, check if error matches",
			wantErr: &errMatches,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDevfileData := data.NewMockDevfileData(ctrl)

			tt.filterOptions.ComponentOptions = common.ComponentOptions{
				ComponentType: v1.ContainerComponentType,
			}
			mockGetComponents := mockDevfileData.EXPECT().GetComponents(tt.filterOptions)

			// set up the mock data
			if len(tt.filterOptions.Filter) == 0 {
				mockGetComponents.Return(tt.containerComponents, nil).AnyTimes()
			} else {
				mockGetComponents.Return(tt.filteredComponents, nil).AnyTimes()
			}
			if tt.wantErr != nil {
				mockGetComponents.Return(nil, fmt.Errorf(*tt.wantErr))
			}
			mockDevfileData.EXPECT().GetProjects(common.DevfileOptions{}).Return(projects, nil).AnyTimes()

			// to set up the prestartevent and apply command for init container
			mockGetCommands := mockDevfileData.EXPECT().GetCommands(common.DevfileOptions{})
			mockGetCommands.Return(applyCommands, nil).AnyTimes()
			events := v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: tt.eventCommands.preStart,
					PostStop: tt.eventCommands.postStop,
				},
			}
			mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()

			devObj := parser.DevfileObj{
				Data: mockDevfileData,
			}

			containers, err := GetContainers(devObj, tt.filterOptions)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetContainers() error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				for _, container := range containers {
					if container.Name != tt.wantContainerName {
						t.Errorf("TestGetContainers() error: Name mismatch - got: %s, wanted: %s", container.Name, tt.wantContainerName)
					}
					if container.Image != tt.wantContainerImage {
						t.Errorf("TestGetContainers() error: Image mismatch - got: %s, wanted: %s", container.Image, tt.wantContainerImage)
					}
					if len(container.Env) > 0 && !reflect.DeepEqual(container.Env, tt.wantContainerEnv) {
						t.Errorf("TestGetContainers() error: Env mismatch - got: %+v, wanted: %+v", container.Env, tt.wantContainerEnv)
					}
					if len(container.VolumeMounts) > 0 && !reflect.DeepEqual(container.VolumeMounts, tt.wantContainerVolMount) {
						t.Errorf("TestGetContainers() error: Vol Mount mismatch - got: %+v, wanted: %+v", container.VolumeMounts, tt.wantContainerVolMount)
					}
					if tt.wantContainerOverrideData != nil && !reflect.DeepEqual(container, *tt.wantContainerOverrideData) {
						t.Errorf("TestGetContainers() error: Container override mismatch - got: %+v, wanted: %+v", container, *tt.wantContainerOverrideData)
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetContainers(): Error message does not match")
			}
		})
	}
}

func TestGetVolumesAndVolumeMounts(t *testing.T) {

	type testVolumeMountInfo struct {
		mountPath  string
		volumeName string
	}

	errMatches := "an expected error"
	trueEphemeral := true

	tests := []struct {
		name                string
		containerComponents []v1.Component
		volumeComponents    []v1.Component
		volumeNameToVolInfo map[string]VolumeInfo
		wantContainerToVol  map[string][]testVolumeMountInfo
		ephemeralVol        bool
		wantErr             *string
	}{
		{
			name:                "One volume mounted",
			containerComponents: []v1.Component{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
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
		},
		{
			name: "One volume mounted at diff locations",
			containerComponents: []v1.Component{
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
		},
		{
			name: "One volume mounted at diff container components",
			containerComponents: []v1.Component{
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
		},
		{
			name: "Ephemeral volume",
			containerComponents: []v1.Component{
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
			},
			volumeComponents: []v1.Component{
				{
					Name: "volume1",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{
							Volume: v1.Volume{
								Ephemeral: &trueEphemeral,
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
			ephemeralVol: true,
			wantContainerToVol: map[string][]testVolumeMountInfo{
				"container1": {
					{
						mountPath:  "/path1",
						volumeName: "volume1-pvc-vol",
					},
				},
			},
		},
		{
			name:    "Simulating error case, check if error matches",
			wantErr: &errMatches,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDevfileData := data.NewMockDevfileData(ctrl)

			mockGetContainerComponents := mockDevfileData.EXPECT().GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.ContainerComponentType,
				},
			})

			mockGetVolumeComponents := mockDevfileData.EXPECT().GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.VolumeComponentType,
				},
			})

			// set up the mock data
			mockGetContainerComponents.Return(tt.containerComponents, nil).AnyTimes()
			mockGetVolumeComponents.Return(tt.volumeComponents, nil).AnyTimes()
			mockDevfileData.EXPECT().GetProjects(common.DevfileOptions{}).Return(nil, nil).AnyTimes()

			devObj := parser.DevfileObj{
				Data: mockDevfileData,
			}

			containers, err := getAllContainers(devObj, common.DevfileOptions{})
			if err != nil {
				t.Errorf("TestGetVolumesAndVolumeMounts error - %v", err)
				return
			}

			if tt.wantErr != nil {
				// simulate error condition
				mockGetContainerComponents.Return(nil, fmt.Errorf(*tt.wantErr))

			}

			volumeParams := VolumeParams{
				Containers:             containers,
				VolumeNameToVolumeInfo: tt.volumeNameToVolInfo,
			}

			pvcVols, err := GetVolumesAndVolumeMounts(devObj, volumeParams, common.DevfileOptions{})
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetVolumesAndVolumeMounts() error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				// check if the pvc volumes returned are correct
				for _, volInfo := range tt.volumeNameToVolInfo {
					matched := false
					for _, pvcVol := range pvcVols {
						emptyDirVolCondition := tt.ephemeralVol && reflect.DeepEqual(pvcVol.EmptyDir, &corev1.EmptyDirVolumeSource{})
						pvcCondition := pvcVol.PersistentVolumeClaim != nil && volInfo.PVCName == pvcVol.PersistentVolumeClaim.ClaimName
						if volInfo.VolumeName == pvcVol.Name && (emptyDirVolCondition || pvcCondition) {
							matched = true
						}
					}

					if !matched {
						t.Errorf("TestGetVolumesAndVolumeMounts() error: could not find volume details %s in the actual result", volInfo.VolumeName)
					}
				}

				// check the volume mounts of the containers
				for _, container := range containers {
					if volMounts, ok := tt.wantContainerToVol[container.Name]; !ok {
						t.Errorf("TestGetVolumesAndVolumeMounts() error: did not find the expected container %s", container.Name)
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
								t.Errorf("TestGetVolumesAndVolumeMounts() error: could not find volume mount details for path %s in the actual result for container %s", expectedVolMount.mountPath, container.Name)
							}
						}
					}
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetVolumesAndVolumeMounts(): Error message does not match")
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
				t.Errorf("TestGetVolumeMountPath() error: mount path mismatch, expected: %v got: %v", tt.wantPath, path)
			}
		})
	}

}

func TestGetInitContainers(t *testing.T) {
	shellExecutable := "/bin/sh"
	containers := []v1.Component{
		{
			Name: "container1",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image:   "container1",
						Command: []string{shellExecutable, "-c", "cd execworkdir1 && execcommand1"},
					},
				},
			},
		},
		{
			Name: "container2",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image:   "container2",
						Command: []string{shellExecutable, "-c", "cd execworkdir3 && execcommand3"},
					},
				},
			},
		},
	}

	applyCommands := []v1.Command{
		{
			Id: "apply1",
			CommandUnion: v1.CommandUnion{
				Apply: &v1.ApplyCommand{
					Component: "container1",
				},
			},
		},
		{
			Id: "apply2",
			CommandUnion: v1.CommandUnion{
				Apply: &v1.ApplyCommand{
					Component: "container1",
				},
			},
		},
		{
			Id: "apply3",
			CommandUnion: v1.CommandUnion{
				Apply: &v1.ApplyCommand{
					Component: "container2",
				},
			},
		},
	}

	compCommands := []v1.Command{
		{
			Id: "comp1",
			CommandUnion: v1.CommandUnion{
				Composite: &v1.CompositeCommand{
					Commands: []string{
						"apply1",
						"apply3",
					},
				},
			},
		},
	}

	longContainerName := "thisisaverylongcontainerandkuberneteshasalimitforanamesize-exec2"
	trimmedLongContainerName := util.TruncateString(longContainerName, containerNameMaxLen)

	errMatches := "an expected error"

	tests := []struct {
		name              string
		eventCommands     []string
		wantInitContainer map[string]corev1.Container
		longName          bool
		wantErr           *string
	}{
		{
			name: "Composite and Exec events",
			eventCommands: []string{
				"apply1",
				"apply3",
				"apply2",
			},
			wantInitContainer: map[string]corev1.Container{
				"container1-apply1": {
					Command: []string{shellExecutable, "-c", "cd execworkdir1 && execcommand1"},
				},
				"container1-apply2": {
					Command: []string{shellExecutable, "-c", "cd execworkdir1 && execcommand1"},
				},
				"container2-apply3": {
					Command: []string{shellExecutable, "-c", "cd execworkdir3 && execcommand3"},
				},
			},
		},
		{
			name: "Simulate error case, check if error matches",
			eventCommands: []string{
				"apply1",
				"apply3",
				"apply2",
			},
			wantErr: &errMatches,
		},
		{
			name: "Long Container Name",
			eventCommands: []string{
				"apply2",
			},
			wantInitContainer: map[string]corev1.Container{
				trimmedLongContainerName: {
					Command: []string{shellExecutable, "-c", "cd execworkdir1 && execcommand1"},
				},
			},
			longName: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			preStartEvents := v1.Events{
				DevWorkspaceEvents: v1.DevWorkspaceEvents{
					PreStart: tt.eventCommands,
				},
			}

			if tt.longName {
				containers[0].Name = longContainerName
				applyCommands[1].Apply.Component = longContainerName
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDevfileData := data.NewMockDevfileData(ctrl)

			mockGetCommands := mockDevfileData.EXPECT().GetCommands(common.DevfileOptions{})

			// set up the mock data
			mockDevfileData.EXPECT().GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.ContainerComponentType,
				},
			}).Return(containers, nil).AnyTimes()
			mockDevfileData.EXPECT().GetProjects(common.DevfileOptions{}).Return(nil, nil).AnyTimes()
			mockDevfileData.EXPECT().GetEvents().Return(preStartEvents).AnyTimes()
			mockGetCommands.Return(append(applyCommands, compCommands...), nil).AnyTimes()

			if tt.wantErr != nil {
				mockGetCommands.Return(nil, fmt.Errorf(*tt.wantErr)).AnyTimes()
			}

			devObj := parser.DevfileObj{
				Data: mockDevfileData,
			}

			initContainers, err := GetInitContainers(devObj)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetInitContainers() error: %v, wantErr %v", err, tt.wantErr)
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetInitContainers: Error message does not match")
				return
			}

			if len(tt.wantInitContainer) != len(initContainers) {
				t.Errorf("TestGetInitContainers() error: init container length mismatch, wanted %v got %v", len(tt.wantInitContainer), len(initContainers))
			}

			for _, initContainer := range initContainers {
				nameMatched := false
				commandMatched := false
				for containerName, container := range tt.wantInitContainer {
					if strings.Contains(initContainer.Name, containerName) {
						nameMatched = true
					}

					if reflect.DeepEqual(initContainer.Command, container.Command) {
						commandMatched = true
					}
				}

				if !nameMatched {
					t.Errorf("TestGetInitContainers() error: init container name mismatch, container name not present in %v", initContainer.Name)
				}

				if !commandMatched {
					t.Errorf("TestGetInitContainers() error: init container command mismatch, command not found in %v", initContainer.Command)
				}
			}
		})
	}

}

func TestGetService(t *testing.T) {
	trueBool := true

	serviceParams := ServiceParams{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"preserved-key": "preserved-value",
			},
		},
	}

	tests := []struct {
		name                string
		containerComponents []v1.Component
		expected            corev1.Service
	}{
		{
			// Currently dedicatedPod can only filter out annotations
			// ToDo: dedicatedPod support: https://github.com/devfile/api/issues/670
			name: "has dedicated pod",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Service: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Service: map[string]string{
						"key2": "value2",
					},
				}, &trueBool),
			},
			expected: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
						"key1":          "value1",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "http-8080",
							Port:       8080,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			},
		},
		{
			name: "no dedicated pod",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Service: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Service: map[string]string{
						"key2": "value2",
					},
				}, nil),
			},
			expected: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
						"key1":          "value1",
						"key2":          "value2",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "http-8080",
							Port:       8080,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDevfileData := data.NewMockDevfileData(ctrl)

			options := common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.ContainerComponentType,
				},
			}
			// set up the mock data
			mockGetComponents := mockDevfileData.EXPECT().GetComponents(options)
			mockGetComponents.Return(tt.containerComponents, nil).AnyTimes()
			mockDevfileData.EXPECT().GetProjects(common.DevfileOptions{}).Return(nil, nil).AnyTimes()
			mockDevfileData.EXPECT().GetEvents().Return(v1.Events{}).AnyTimes()

			devObj := parser.DevfileObj{
				Data: mockDevfileData,
			}
			svc, err := GetService(devObj, serviceParams, common.DevfileOptions{})
			// Checks for unexpected error cases
			if err != nil {
				t.Errorf("TestGetService(): unexpected error %v", err)
			}
			assert.Equal(t, tt.expected, *svc, "TestGetService(): The two values should be the same.")

		})
	}
}

func TestGetDeployment(t *testing.T) {
	trueBool := true
	podTemplateSpec := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
				},
				{
					Name: "container2",
				},
			},
		},
	}

	objectMeta := metav1.ObjectMeta{
		Annotations: map[string]string{
			"preserved-key": "preserved-value",
			"key1":          "value1",
			"key2":          "value2",
		},
	}

	objectMetaDedicatedPod := metav1.ObjectMeta{
		Annotations: map[string]string{
			"preserved-key": "preserved-value",
			"key1":          "value1",
		},
	}

	tests := []struct {
		name                string
		containerComponents []v1.Component
		deploymentParams    DeploymentParams
		expected            *appsv1.Deployment
		wantErr             bool
		devObj              func(ctrl *gomock.Controller, containerComponents []v1.Component) parser.DevfileObj
	}{
		{
			// Currently dedicatedPod can only filter out annotations
			// ToDo: dedicatedPod support: https://github.com/devfile/api/issues/670
			name: "has dedicated pod",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Deployment: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Deployment: map[string]string{
						"key2": "value2",
					},
				}, &trueBool),
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				PodTemplateSpec: &podTemplateSpec,
				Replicas:        pointer.Int32Ptr(1),
			},
			expected: &appsv1.Deployment{
				ObjectMeta: objectMetaDedicatedPod,
				Spec: appsv1.DeploymentSpec{
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: nil,
					},
					Template: podTemplateSpec,
					Replicas: pointer.Int32Ptr(1),
				},
			},
		},
		{
			name: "no dedicated pod",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Deployment: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Deployment: map[string]string{
						"key2": "value2",
					},
				}, nil),
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				PodTemplateSpec: &podTemplateSpec,
			},
			expected: &appsv1.Deployment{
				ObjectMeta: objectMeta,
				Spec: appsv1.DeploymentSpec{
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: nil,
					},
					Template: podTemplateSpec,
				},
			},
		},
		{
			name: "pod should have pod-overrides attribute",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Deployment: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Deployment: map[string]string{
						"key2": "value2",
					},
				}, nil),
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				PodTemplateSpec: &podTemplateSpec,
			},
			expected: &appsv1.Deployment{
				ObjectMeta: objectMeta,
				Spec: appsv1.DeploymentSpec{
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: nil,
					},
					Template: podTemplateSpec,
				},
			},
		},
		{
			name: "skip getting global attributes for SchemaVersion less than 2.1.0",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Deployment: map[string]string{
						"key1": "value1",
					},
				}, nil),
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Deployment: map[string]string{
						"key2": "value2",
					},
				}, nil),
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				PodTemplateSpec: &podTemplateSpec,
			},
			expected: &appsv1.Deployment{
				ObjectMeta: objectMeta,
				Spec: appsv1.DeploymentSpec{
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: nil,
					},
					Template: podTemplateSpec,
				},
			},
			wantErr: false,
			devObj: func(ctrl *gomock.Controller, containerComponents []v1.Component) parser.DevfileObj {
				mockDevfileData := data.NewMockDevfileData(ctrl)

				options := common.DevfileOptions{
					ComponentOptions: common.ComponentOptions{
						ComponentType: v1.ContainerComponentType,
					},
				}
				// set up the mock data
				mockDevfileData.EXPECT().GetDevfileContainerComponents(common.DevfileOptions{}).Return(containerComponents, nil).AnyTimes()
				mockDevfileData.EXPECT().GetComponents(options).Return(containerComponents, nil).AnyTimes()

				devObj := parser.DevfileObj{
					Data: mockDevfileData,
				}
				return devObj
			},
		},
		{
			name: "both container and podtemplatespec are passed",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container1", nil, []v1.Endpoint{
					{
						Name:       "http-8080",
						TargetPort: 8080,
					},
				}, nil, v1.Annotation{
					Deployment: map[string]string{
						"key1": "value1",
					},
				}, nil),
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				PodTemplateSpec: &podTemplateSpec,
				Containers: []corev1.Container{
					{
						Name: "container1",
					},
				},
				Replicas: pointer.Int32Ptr(1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var devObj parser.DevfileObj
			if tt.devObj != nil {
				devObj = tt.devObj(ctrl, tt.containerComponents)
			} else {
				mockDevfileData := data.NewMockDevfileData(ctrl)

				options := common.DevfileOptions{
					ComponentOptions: common.ComponentOptions{
						ComponentType: v1.ContainerComponentType,
					},
				}
				// set up the mock data
				mockDevfileData.EXPECT().GetDevfileContainerComponents(common.DevfileOptions{}).Return(tt.containerComponents, nil).AnyTimes()
				mockDevfileData.EXPECT().GetComponents(options).Return(tt.containerComponents, nil).AnyTimes()

				devObj = parser.DevfileObj{
					Data: mockDevfileData,
				}
			}
			deploy, err := GetDeployment(devObj, tt.deploymentParams)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetDeployment(): unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expected, deploy, "TestGetDeployment(): The two values should be the same.")

		})
	}
}

func TestGetPodTemplateSpec(t *testing.T) {
	type args struct {
		devfileObj        func(ctrl *gomock.Controller) parser.DevfileObj
		podTemplateParams PodTemplateParams
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.PodTemplateSpec
		wantErr bool
	}{
		{
			name: "Devfile with wrong container-override",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{
								ContainerOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": \"serviceAccountName\": \"new-service-account\"}")},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "GetContainers returns err",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(nil, errors.New("an error")).AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "GetDevfileContainerComponents returns err",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{
								ContainerOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": \"serviceAccountName\": \"new-service-account\"}")},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(nil, errors.New("an error")).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			wantErr: true,
		},
		{
			name: "Devfile with local container-override on the first container only",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
								"container-overrides": map[string]interface{}{"securityContext": map[string]int64{"runAsGroup": 3000}},
							}, nil),
						},
						{
							Name: "tools",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "a-tool-image",
									},
								},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{
						PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")}}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "new-service-account",
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: &corev1.SecurityContext{
								RunAsGroup: pointer.Int64(3000),
							},
						},
						{
							Name:  "tools",
							Image: "a-tool-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Devfile with local container-override and global pod-override",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
								"container-overrides": map[string]interface{}{"securityContext": map[string]int64{"runAsGroup": 3000}},
							}, nil),
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{
						PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")}}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "new-service-account",
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: &corev1.SecurityContext{
								RunAsGroup: pointer.Int64(3000),
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Devfile with local container-override and local pod-override",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{
								ContainerOverridesAttribute: apiext.JSON{Raw: []byte("{\"securityContext\": {\"runAsGroup\": 3000}}")},
								PodOverridesAttribute:       apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "new-service-account",
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: &corev1.SecurityContext{
								RunAsGroup: pointer.Int64(3000),
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Devfile with pod-override at local ang global level",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{
								PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"schedulerName\": \"new-scheduler\"}}")},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{
						PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")},
					}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "new-service-account",
					SchedulerName:      "new-scheduler",
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Devfile with global container-override and pod-override",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
							Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
								"container-overrides": map[string]interface{}{"securityContext": map[string]int64{"runAsGroup": 3000}},
							}, nil),
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{
						PodOverridesAttribute:       apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")},
						ContainerOverridesAttribute: apiext.JSON{Raw: []byte("{\"securityContext\": {\"runAsGroup\": 3000}}")},
					}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "new-service-account",
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							SecurityContext: &corev1.SecurityContext{
								RunAsGroup: pointer.Int64(3000),
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Restricted policy",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{
						PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"securityContext\": {\"seccompProfile\": {\"type\": \"Localhost\"}}}}")},
					}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
				podTemplateParams: PodTemplateParams{
					PodSecurityAdmissionPolicy: api.Policy{
						Enforce: api.LevelVersion{
							Level:   api.LevelRestricted,
							Version: api.MajorMinorVersion(1, 25),
						},
					},
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: pointer.Bool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: "Localhost",
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Restricted policy and pod override",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
				podTemplateParams: PodTemplateParams{
					PodSecurityAdmissionPolicy: api.Policy{
						Enforce: api.LevelVersion{
							Level:   api.LevelRestricted,
							Version: api.MajorMinorVersion(1, 25),
						},
					},
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: pointer.Bool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: "RuntimeDefault",
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Baseline policy",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
						},
					}
					events := v1alpha2.Events{}
					mockDevfileData := data.NewMockDevfileData(ctrl)
					mockDevfileData.EXPECT().GetComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetDevfileContainerComponents(gomock.Any()).Return(containers, nil).AnyTimes()
					mockDevfileData.EXPECT().GetEvents().Return(events).AnyTimes()
					mockDevfileData.EXPECT().GetProjects(gomock.Any()).Return(nil, nil).AnyTimes()
					mockDevfileData.EXPECT().GetAttributes().Return(attributes.Attributes{}, nil)
					mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0").AnyTimes()
					return parser.DevfileObj{
						Data: mockDevfileData,
					}
				},
				podTemplateParams: PodTemplateParams{
					PodSecurityAdmissionPolicy: api.Policy{
						Enforce: api.LevelVersion{
							Level:   api.LevelBaseline,
							Version: api.MajorMinorVersion(1, 25),
						},
					},
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "an-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
		{
			name: "Filter components by name",
			args: args{
				devfileObj: func(ctrl *gomock.Controller) parser.DevfileObj {
					containers := []v1alpha2.Component{
						{
							Name: "main",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "an-image",
									},
								},
							},
						},
						{
							Name: "tools",
							ComponentUnion: v1.ComponentUnion{
								Container: &v1.ContainerComponent{
									Container: v1.Container{
										Image: "a-tool-image",
									},
								},
							},
						},
					}
					parserArgs := parser.ParserArgs{
						Data: []byte(`schemaVersion: 2.2.0`),
					}
					var err error
					devfile, _, err := devfile.ParseDevfileAndValidate(parserArgs)
					if err != nil {
						t.Errorf("error creating devfile: %v", err)
					}
					devfile.Ctx = context.FakeContext(filesystem.NewFakeFs(), "/devfile.yaml")
					devfile.Data.AddComponents(containers)
					return devfile
				},
				podTemplateParams: PodTemplateParams{
					Options: common.DevfileOptions{
						FilterByName: "tools",
					},
				},
			},
			want: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "tools",
							Image: "a-tool-image",
							Env: []corev1.EnvVar{
								{Name: "PROJECT_SOURCE", Value: "/projects"},
								{Name: "PROJECTS_ROOT", Value: "/projects"},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports:           []corev1.ContainerPort{},
						},
					},
					InitContainers: []corev1.Container{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			devObj := tt.args.devfileObj(ctrl)

			got, err := GetPodTemplateSpec(devObj, tt.args.podTemplateParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodTemplateSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetPodTemplateSpec()  mismatch (-want +got): %s\n", diff)
			}
		})
	}
}
