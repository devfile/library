package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
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
					Name: "component2",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{},
					},
				},
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

			components := d.GetComponents()

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
