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

package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestDevfile200_SetParent(t *testing.T) {

	tests := []struct {
		name              string
		parent            *v1.Parent
		devfilev2         *DevfileV2
		expectedDevfilev2 *DevfileV2
	}{
		{
			name: "set parent",
			devfilev2: &DevfileV2{
				v1.Devfile{},
			},
			parent: &v1.Parent{
				ImportReference: v1.ImportReference{
					RegistryUrl: "testRegistryUrl",
				},
				ParentOverrides: v1.ParentOverrides{},
			},
			expectedDevfilev2: &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						Parent: &v1.Parent{
							ImportReference: v1.ImportReference{
								RegistryUrl: "testRegistryUrl",
							},
							ParentOverrides: v1.ParentOverrides{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.devfilev2.SetParent(tt.parent)
			if !reflect.DeepEqual(tt.devfilev2, tt.expectedDevfilev2) {
				t.Errorf("TestDevfile200_SetParent() error: expected %v, got %v", tt.expectedDevfilev2, tt.devfilev2)
			}
		})
	}
}
