package validate

// import (
// 	"strings"
// 	"testing"

// 	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
// 	"github.com/devfile/parser/pkg/devfile/parser"
// 	"github.com/devfile/parser/pkg/testingutil"
// )

// func TestValidateEvents(t *testing.T) {

// 	containers := []string{"container1", "container2"}
// 	dummyComponents := []v1.Component{
// 		{
// 			Container: &v1.ContainerComponent{
// 				Container: v1.Container{
// 					Name: containers[0],
// 				},
// 			},
// 		},
// 		{
// 			Container: &v1.ContainerComponent{
// 				Container: v1.Container{
// 					Name: containers[1],
// 				}},
// 		},
// 	}

// 	tests := []struct {
// 		name         string
// 		events       v1.Events
// 		components   []v1.Component
// 		execCommands []v1.ExecCommand
// 		compCommands []v1.CompositeCommand
// 		wantErr      bool
// 	}{
// 		{
// 			name:       "Case 1: Valid events",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			events: v1.Events{
// 				WorkspaceEvents: v1.WorkspaceEvents{
// 					PostStart: []string{
// 						"command1",
// 					},
// 					PreStop: []string{
// 						"composite1",
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:       "Case 2: Invalid events with wrong mapping to devfile command",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			events: v1.Events{
// 				WorkspaceEvents: v1.WorkspaceEvents{
// 					PostStart: []string{
// 						"command1iswrong",
// 					},
// 					PreStop: []string{
// 						"composite1",
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:       "Case 3: Invalid event command with mapping to wrong devfile container component",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   "wrongcomponent",
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			events: v1.Events{
// 				WorkspaceEvents: v1.WorkspaceEvents{
// 					PostStart: []string{
// 						"command1",
// 					},
// 					PreStop: []string{
// 						"composite1",
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name:       "Case 4: Invalid events with wrong child command in composite command",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1iswrong", "command2"},
// 				},
// 			},
// 			events: v1.Events{
// 				WorkspaceEvents: v1.WorkspaceEvents{
// 					PostStart: []string{
// 						"command1",
// 					},
// 					PreStop: []string{
// 						"composite1",
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			devObj := parser.DevfileObj{
// 				Data: testingutil.TestDevfileData{
// 					Components:        tt.components,
// 					ExecCommands:      tt.execCommands,
// 					CompositeCommands: tt.compCommands,
// 					Events:            tt.events,
// 				},
// 			}
// 			err := validateEvents(devObj.Data)
// 			if err != nil && !tt.wantErr {
// 				t.Errorf("TestValidateEvents error - %v", err)
// 			}
// 		})
// 	}

// }

// func TestIsEventValid(t *testing.T) {

// 	containers := []string{"container1", "container2"}
// 	dummyComponents := []v1.Component{
// 		{
// 			Container: &v1.ContainerComponent{
// 				Container: v1.Container{
// 					Name: containers[0],
// 				},
// 			},
// 		},
// 		{
// 			Container: &v1.ContainerComponent{
// 				Container: v1.Container{
// 					Name: containers[1],
// 				},
// 			},
// 		},
// 	}

// 	tests := []struct {
// 		name         string
// 		eventType    string
// 		components   []v1.Component
// 		execCommands []v1.ExecCommand
// 		compCommands []v1.CompositeCommand
// 		eventNames   []string
// 		wantErr      bool
// 		wantErrMsg   string
// 	}{
// 		{
// 			name:       "Case 1: Valid events",
// 			eventType:  "preStart",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			eventNames: []string{
// 				"command1",
// 				"composite1",
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:       "Case 2: Invalid events with wrong mapping to devfile command",
// 			eventType:  "preStart",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			eventNames: []string{
// 				"command12345iswrong",
// 				"composite1",
// 			},
// 			wantErr:    true,
// 			wantErrMsg: "does not map to a valid devfile command",
// 		},
// 		{
// 			name:       "Case 3: Invalid event command with mapping to wrong devfile container component",
// 			eventType:  "preStart",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   "wrongcomponent",
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1", "command2"},
// 				},
// 			},
// 			eventNames: []string{
// 				"command1",
// 				"composite1",
// 			},
// 			wantErr:    true,
// 			wantErrMsg: "does not map to a supported component",
// 		},
// 		{
// 			name:       "Case 4: Invalid events with wrong child command in composite command",
// 			eventType:  "preStart",
// 			components: dummyComponents,
// 			execCommands: []v1.ExecCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command1",
// 						},
// 					},
// 					CommandLine: "/some/command1",
// 					Component:   containers[0],
// 					WorkingDir:  "workDir",
// 				},
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "command2",
// 						},
// 					},
// 					CommandLine: "/some/command2",
// 					Component:   containers[1],
// 					WorkingDir:  "workDir",
// 				},
// 			},
// 			compCommands: []v1.CompositeCommand{
// 				{
// 					LabeledCommand: v1.LabeledCommand{
// 						BaseCommand: v1.BaseCommand{
// 							Id: "composite1",
// 						},
// 					},
// 					Commands: []string{"command1iswrong", "command2"},
// 				},
// 			},
// 			eventNames: []string{
// 				"command1",
// 				"composite1",
// 			},
// 			wantErr:    true,
// 			wantErrMsg: "does not exist in the devfile",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			devObj := parser.DevfileObj{
// 				Data: testingutil.TestDevfileData{
// 					Components:        tt.components,
// 					ExecCommands:      tt.execCommands,
// 					CompositeCommands: tt.compCommands,
// 				},
// 			}

// 			err := isEventValid(devObj.Data, tt.eventNames, tt.eventType)
// 			if err != nil && !tt.wantErr {
// 				t.Errorf("TestIsEventValid error: %v", err)
// 			} else if err != nil && tt.wantErr {
// 				if !strings.Contains(err.Error(), tt.wantErrMsg) {
// 					t.Errorf("TestIsEventValid error mismatch - %s; does not contain: %s", err.Error(), tt.wantErrMsg)
// 				}
// 			}
// 		})
// 	}

// }
