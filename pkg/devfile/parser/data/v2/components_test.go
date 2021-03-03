package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/stretchr/testify/assert"
)

func TestDevfile200_AddComponent(t *testing.T) {

	tests := []struct {
		name              string
		currentComponents []v1.Component
		newComponents     []v1.Component
		wantErr           bool
	}{
		{
			name: "case 1: successfully add the component",
			currentComponents: []v1.Component{
				{
					Name: "component1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
				{
					Name: "component2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			newComponents: []v1.Component{
				{
					Name: "component3",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: error out on duplicate component",
			currentComponents: []v1.Component{
				{
					Name: "component1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
				{
					Name: "component2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			newComponents: []v1.Component{
				{
					Name: "component1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
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
							Components: tt.currentComponents,
						},
					},
				},
			}

			got := d.AddComponents(tt.newComponents)

			if !tt.wantErr && got != nil {
				t.Errorf("TestDevfile200_AddComponents() unexpected error - %+v", got)
			} else if tt.wantErr && got == nil {
				t.Errorf("TestDevfile200_AddComponents() expected error but got nil")
			}

		})
	}
}

func TestDevfile200_UpdateComponent(t *testing.T) {

	tests := []struct {
		name              string
		currentComponents []v1.Component
		newComponent      v1.Component
	}{
		{
			name: "case 1: successfully update the component",
			currentComponents: []v1.Component{
				{
					Name: "Component1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image1",
							},
						},
					},
				},
				{
					Name: "component2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			newComponent: v1.Component{
				Name: "Component1",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{
						Container: v1.Container{
							Image: "image2",
						},
					},
				},
			},
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

			d.UpdateComponent(tt.newComponent)

			components, err := d.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Errorf("TestDevfile200_UpdateComponent() unxpected error %v", err)
				return
			}

			matched := false
			for _, component := range components {
				if reflect.DeepEqual(component, tt.newComponent) {
					matched = true
					break
				}
			}

			if !matched {
				t.Error("TestDevfile200_UpdateComponent() error updating the component")
			}
		})
	}
}

func TestGetDevfileContainerComponents(t *testing.T) {

	tests := []struct {
		name                 string
		component            []v1.Component
		expectedMatchesCount int
		filterOptions        common.DevfileOptions
		wantErr              bool
	}{
		{
			name:                 "Case 1: Invalid devfile",
			component:            []v1.Component{},
			expectedMatchesCount: 0,
		},
		{
			name: "Case 2: Valid devfile with wrong component type (Openshift)",
			component: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Openshift: &v1.OpenshiftComponent{},
					},
				},
			},
			expectedMatchesCount: 0,
		},
		{
			name: "Case 3 : Valid devfile with correct component type (Container)",
			component: []v1.Component{
				testingutil.GetFakeContainerComponent("comp1"),
				testingutil.GetFakeContainerComponent("comp2"),
			},
			expectedMatchesCount: 2,
			filterOptions:        common.DevfileOptions{},
		},
		{
			name: "Case 4 : Get Container component with the specified filter",
			component: []v1.Component{
				{
					Name: "comp1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
				{
					Name: "comp2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString":  "firstStringValue",
					"secondString": "secondStringValue",
				},
			},
			expectedMatchesCount: 1,
		},
		{
			name: "Case 5 : Get Container component with the wrong specified filter",
			component: []v1.Component{
				{
					Name: "comp1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
				{
					Name: "comp2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstStringIsWrong": "firstStringValue",
				},
			},
			expectedMatchesCount: 0,
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.component,
						},
					},
				},
			}

			devfileComponents, err := d.GetDevfileContainerComponents(tt.filterOptions)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetDevfileContainerComponents unexpected error: %v", err)
			} else if tt.wantErr && err == nil {
				t.Errorf("TestGetDevfileContainerComponents expected error but got nil")
			} else if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileContainerComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
			}
		})
	}

}

func TestGetDevfileVolumeComponents(t *testing.T) {

	tests := []struct {
		name                 string
		component            []v1.Component
		expectedMatchesCount int
		filterOptions        common.DevfileOptions
		wantErr              bool
	}{
		{
			name:                 "Case 1: Invalid devfile",
			component:            []v1.Component{},
			expectedMatchesCount: 0,
		},
		{
			name: "Case 2: Valid devfile with wrong component type (Kubernetes)",
			component: []v1.Component{
				{
					ComponentUnion: v1.ComponentUnion{
						Kubernetes: &v1.KubernetesComponent{},
					},
				},
			},
			expectedMatchesCount: 0,
		},
		{
			name: "Case 3: Valid devfile with correct component type (Volume)",
			component: []v1.Component{
				testingutil.GetFakeContainerComponent("comp1"),
				testingutil.GetFakeVolumeComponent("myvol", "4Gi"),
			},
			expectedMatchesCount: 1,
		},
		{
			name: "Case 4 : Get Container component with the specified filter",
			component: []v1.Component{
				{
					Name: "comp1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
				{
					Name: "comp2",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString": "firstStringValue",
						"thirdString": "thirdStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstString": "firstStringValue",
				},
			},
			expectedMatchesCount: 2,
		},
		{
			name: "Case 5 : Get Container component with the wrong specified filter",
			component: []v1.Component{
				{
					Name: "comp1",
					Attributes: attributes.Attributes{}.FromStringMap(map[string]string{
						"firstString":  "firstStringValue",
						"secondString": "secondStringValue",
					}),
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			filterOptions: common.DevfileOptions{
				Filter: map[string]interface{}{
					"firstStringIsWrong": "firstStringValue",
				},
			},
			expectedMatchesCount: 0,
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: tt.component,
						},
					},
				},
			}
			devfileComponents, err := d.GetDevfileVolumeComponents(tt.filterOptions)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetDevfileVolumeComponents unexpected error: %v", err)
			} else if tt.wantErr && err == nil {
				t.Errorf("TestGetDevfileVolumeComponents expected error but got nil")
			} else if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileVolumeComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
			}
		})
	}

}

func TestDeleteComponents(t *testing.T) {

	tests := []struct {
		name              string
		componentToDelete string
		components        []v1.Component
		wantComponents    []v1.Component
		wantErr           bool
	}{
		{
			name:              "Volume Component with mounts",
			componentToDelete: "comp3",
			components: []v1.Component{
				{
					Name: "comp2",
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
					Name: "comp2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
				{
					Name: "comp3",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			wantComponents: []v1.Component{
				{
					Name: "comp2",
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
					Name: "comp2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:              "Missing Component",
			componentToDelete: "comp12",
			components: []v1.Component{
				{
					Name: "comp2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
					},
				},
			},
			wantComponents: []v1.Component{
				{
					Name: "comp2",
					ComponentUnion: v1.ComponentUnion{
						Volume: &v1.VolumeComponent{},
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

			err := d.DeleteComponent(tt.componentToDelete)
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
