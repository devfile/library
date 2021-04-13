package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestDevfile200_SetDevfileWorkspaceSpecContent(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name              string
		workspace         v1.DevWorkspaceTemplateSpecContent
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
	}{
		{
			name: "set workspace",
			devfilev2: &DevfileV2{
				v1.Devfile{},
			},
			workspace: v1.DevWorkspaceTemplateSpecContent{
				Components: []v1.Component{
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
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Components: []v1.Component{
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
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.devfilev2.SetDevfileWorkspaceSpecContent(tt.workspace)
			if !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetDevfileWorkspaceSpecContent() expected %v, got %v", tt.expectedDevfilev2, tt.devfilev2)
			}
		})
	}
}
