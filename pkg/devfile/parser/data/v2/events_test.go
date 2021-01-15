package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestDevfile200_AddEvents(t *testing.T) {

	tests := []struct {
		name          string
		currentEvents v1.Events
		newEvents     v1.Events
		wantErr       bool
	}{
		{
			name: "case 1: successfully add the events",
			currentEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PostStart: []string{"postStart1"},
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: event already present",
			currentEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PreStart: []string{"preStart2"},
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
							Events: &tt.currentEvents,
						},
					},
				},
			}

			got := d.AddEvents(tt.newEvents)

			if !tt.wantErr && got != nil {
				t.Errorf("TestDevfile200_AddEvents() unexpected error - %+v", got)
			} else if tt.wantErr && got == nil {
				t.Errorf("TestDevfile200_AddEvents() expected error but got nil")
			}

		})
	}
}

func TestDevfile200_UpdateEvents(t *testing.T) {

	tests := []struct {
		name          string
		currentEvents v1.Events
		newEvents     v1.Events
	}{
		{
			name: "case 1: successfully add the events",
			currentEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PreStart: []string{"preStart1"},
				},
			},
			newEvents: v1.Events{
				WorkspaceEvents: v1.WorkspaceEvents{
					PreStart:  []string{"preStart2"},
					PostStart: []string{"postStart2"},
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
							Events: &tt.currentEvents,
						},
					},
				},
			}

			d.UpdateEvents(tt.newEvents.PostStart, tt.newEvents.PostStop, tt.newEvents.PreStart, tt.newEvents.PreStop)

			events := d.GetEvents()
			if !reflect.DeepEqual(events, tt.newEvents) {
				t.Errorf("TestDevfile200_UpdateEvents events did not get updated. got - %+v, wanted - %+v", events, tt.newEvents)
			}

		})
	}
}
