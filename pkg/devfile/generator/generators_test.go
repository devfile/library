//
// Copyright 2022 Red Hat, Inc.
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
	"fmt"
	"reflect"
	"strings"
	"testing"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/testingutil"
	"github.com/devfile/library/v2/pkg/util"
	"github.com/golang/mock/gomock"

	corev1 "k8s.io/api/core/v1"
)

var fakeResources corev1.ResourceRequirements

func init() {
	fakeResources, _ = testingutil.FakeResourceRequirements("0.5m", "300Mi")
}

func TestGetContainers(t *testing.T) {

	containerNames := []string{"testcontainer1", "testcontainer2", "testcontainer3"}
	containerImages := []string{"image1", "image2", "image3"}
	defaultPullPolicy := corev1.PullAlways
	defaultEnv := []corev1.EnvVar{
		{Name: "PROJECTS_ROOT", Value: "/projects"},
		{Name: "PROJECT_SOURCE", Value: "/projects/test-project"},
	}

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
		{
			name: "container with container-overrides",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: containerImages[0],
							},
						},
					},
					Attributes: attributes.Attributes{}.FromMap(map[string]interface{}{
						"container-overrides": map[string]interface{}{"securityContext": map[string]int64{"runAsGroup": 3000}},
					}, nil),
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv:   defaultEnv,
			wantContainerOverrideData: &corev1.Container{
				Name:            containerNames[0],
				Image:           containerImages[0],
				Env:             defaultEnv,
				ImagePullPolicy: defaultPullPolicy,
				SecurityContext: &corev1.SecurityContext{
					RunAsGroup: pointer.Int64(3000),
				},
			},
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
	containers := []corev1.Container{
		{
			Name: "container1",
		},
		{
			Name: "container2",
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
		attributes          attributes.Attributes
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
				Containers: containers,
				Replicas:   pointer.Int32Ptr(1),
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
					Template: corev1.PodTemplateSpec{
						ObjectMeta: objectMetaDedicatedPod,
						Spec: corev1.PodSpec{
							Containers: containers,
						},
					},
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
				Containers: containers,
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
					Template: corev1.PodTemplateSpec{
						ObjectMeta: objectMeta,
						Spec: corev1.PodSpec{
							Containers: containers,
						},
					},
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
			attributes: attributes.Attributes{
				PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": {\"serviceAccountName\": \"new-service-account\"}}")},
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				Containers: containers,
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
					Template: corev1.PodTemplateSpec{
						ObjectMeta: objectMeta,
						Spec: corev1.PodSpec{
							Containers:         containers,
							ServiceAccountName: "new-service-account",
						},
					},
				},
			},
		},
		{
			name: "pod has an invalid pod-overrides attribute that throws error",
			containerComponents: []v1.Component{
				testingutil.GenerateDummyContainerComponent("container2", nil, nil, nil, v1.Annotation{
					Deployment: map[string]string{
						"key2": "value2",
					},
				}, nil),
			},
			attributes: attributes.Attributes{
				PodOverridesAttribute: apiext.JSON{Raw: []byte("{\"spec\": \"serviceAccountName\": \"new-service-account\"}}")},
			},
			deploymentParams: DeploymentParams{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"preserved-key": "preserved-value",
					},
				},
				Containers: containers,
			},
			expected: nil,
			wantErr:  trueBool,
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
				Containers: containers,
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
					Template: corev1.PodTemplateSpec{
						ObjectMeta: objectMeta,
						Spec: corev1.PodSpec{
							Containers: containers,
						},
					},
				},
			},
			attributes: nil,
			wantErr:    false,
			devObj: func(ctrl *gomock.Controller, containerComponents []v1.Component) parser.DevfileObj {
				mockDevfileData := data.NewMockDevfileData(ctrl)

				options := common.DevfileOptions{
					ComponentOptions: common.ComponentOptions{
						ComponentType: v1.ContainerComponentType,
					},
				}
				// set up the mock data
				mockDevfileData.EXPECT().GetSchemaVersion().Return("2.0.0")
				mockDevfileData.EXPECT().GetDevfileContainerComponents(common.DevfileOptions{}).Return(containerComponents, nil).AnyTimes()
				mockDevfileData.EXPECT().GetComponents(options).Return(containerComponents, nil).AnyTimes()

				devObj := parser.DevfileObj{
					Data: mockDevfileData,
				}
				return devObj
			},
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
				mockDevfileData.EXPECT().GetSchemaVersion().Return("2.1.0")
				mockDevfileData.EXPECT().GetAttributes().Return(tt.attributes, nil).AnyTimes()
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
