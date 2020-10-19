package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/kylelemons/godebug/pretty"
)

func TestAddStarterProjects(t *testing.T) {
	currentProject := []v1.StarterProject{
		{
			Project: v1.Project{
				Name: "java-starter",
			},
			Description: "starter project for java",
		},
		{
			Project: v1.Project{
				Name: "quarkus-starter",
			},
			Description: "starter project for quarkus",
		},
	}

	d := &DevfileV2{
		v1.Devfile{
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					StarterProjects: currentProject,
				},
			},
		},
	}

	tests := []struct {
		name    string
		wantErr bool
		args    []v1.StarterProject
	}{
		{
			name:    "case:1 It should add starter project",
			wantErr: false,
			args: []v1.StarterProject{
				{
					Project: v1.Project{
						Name: "nodejs",
					},
					Description: "starter project for nodejs",
				},
				{
					Project: v1.Project{
						Name: "spring-pet-clinic",
					},
					Description: "starter project for springboot",
				},
			},
		},

		{
			name:    "case:2 It should give error if tried to add already present starter project",
			wantErr: true,
			args: []v1.StarterProject{
				{
					Project: v1.Project{
						Name: "quarkus-starter",
					},
					Description: "starter project for quarkus",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.AddStarterProjects(tt.args)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("expected error, got %v", err)
				return
			}
			wantProjects := append(currentProject, tt.args...)

			if !reflect.DeepEqual(d.StarterProjects, wantProjects) {
				t.Errorf("wanted: %v, got: %v, difference at %v", wantProjects, d.StarterProjects, pretty.Compare(wantProjects, d.StarterProjects))
			}
		})
	}

}
