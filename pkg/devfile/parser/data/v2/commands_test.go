package v2

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func TestDevfile200_GetCommands(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name            string
		currentCommands []v1.Command
		wantCommands    []string
	}{
		{
			name: "case 1: get the necessary commands",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
					},
				},
			},
			wantCommands: []string{"command1", "command2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DevfileV2{
				v1.Devfile{
					DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
							Commands: tt.currentCommands,
						},
					},
				},
			}

			commandsMap := d.GetCommands()

			for _, wantCommand := range tt.wantCommands {
				if _, ok := commandsMap[wantCommand]; !ok {
					t.Errorf("TestDevfile200_GetCommands() error - command %s not found in the devfile", wantCommand)
				}
			}
		})
	}
}

func TestDevfile200_AddCommands(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name            string
		currentCommands []v1.Command
		newCommands     []v1.Command
		wantErr         bool
	}{
		{
			name: "case 1: Command does not exist",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			newCommands: []v1.Command{
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
				{
					Id: "command3",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: Command does exist",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
					},
				},
			},
			newCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{},
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
							Commands: tt.currentCommands,
						},
					},
				},
			}

			got := d.AddCommands(tt.newCommands...)
			if !tt.wantErr && got != nil {
				t.Errorf("TestDevfile200_AddCommands() unexpected error - %v", got)
			} else if tt.wantErr && got == nil {
				t.Errorf("TestDevfile200_AddCommands() wanted an error but got nil")
			}
		})
	}
}

func TestDevfile200_UpdateCommands(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name            string
		currentCommands []v1.Command
		newCommand      v1.Command
	}{
		{
			name: "case 1: update the command",
			currentCommands: []v1.Command{
				{
					Id: "command1",
					CommandUnion: v1.CommandUnion{
						Exec: &v1.ExecCommand{
							Component: "component1",
						},
					},
				},
				{
					Id: "command2",
					CommandUnion: v1.CommandUnion{
						Composite: &v1.CompositeCommand{},
					},
				},
			},
			newCommand: v1.Command{
				Id: "command1",
				CommandUnion: v1.CommandUnion{
					Exec: &v1.ExecCommand{
						Component: "component1new",
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
							Commands: tt.currentCommands,
						},
					},
				},
			}

			d.UpdateCommand(tt.newCommand)

			commandsMap := d.GetCommands()

			if updatedCommand, ok := commandsMap[tt.newCommand.Id]; ok {
				if !reflect.DeepEqual(updatedCommand, tt.newCommand) {
					t.Errorf("TestDevfile200_UpdateCommands() command mismatch - wanted %+v, got %+v", tt.newCommand, updatedCommand)
				}
			} else {
				t.Errorf("TestDevfile200_UpdateCommands() command mismatch - did not find command with id %s", tt.newCommand.Id)
			}
		})
	}
}
