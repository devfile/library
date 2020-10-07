package common

import (
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func TestIsContainer(t *testing.T) {

	tests := []struct {
		name            string
		component       v1.Component
		wantIsSupported bool
	}{
		{
			name: "Case 1: Container component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{},
				},
			},
			wantIsSupported: true,
		},
		{
			name: "Case 2: Not a container component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Openshift: &v1.OpenshiftComponent{},
				},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := IsContainer(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsContainer error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}

func TestIsVolume(t *testing.T) {

	tests := []struct {
		name            string
		component       v1.Component
		wantIsSupported bool
	}{
		{
			name: "Case 1: Volume component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Volume: &v1.VolumeComponent{
						Volume: v1.Volume{
							Size: "size",
						},
					},
				},
			},
			wantIsSupported: true,
		},
		{
			name: "Case 2: Not a volume component",
			component: v1.Component{
				Name: "name",
				ComponentUnion: v1.ComponentUnion{
					Openshift: &v1.OpenshiftComponent{},
				},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := IsVolume(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsVolume error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}
