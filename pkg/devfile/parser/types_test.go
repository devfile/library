package parser

import (
	"reflect"
	"testing"

	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
	"github.com/devfile/parser/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

const devfileTempPath = "devfile.yaml"

func TestDevfileObj_OverrideCommands(t *testing.T) {
	componentName0 := "component-0"
	overrideComponent0 := "override-component-0"

	commandLineBuild := "npm build"
	overrideBuild := "npm custom build"
	commandLineRun := "npm run"

	workingDir := "/project"
	overrideWorkingDir := "/data"

	type args struct {
		overridePatch []v1.CommandParentOverride
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override a command's non list/map fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
												Component:   componentName0,
												Env:         nil,
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Group: &v1.CommandGroup{
															IsDefault: false,
															Kind:      v1.BuildCommandGroupKind,
														},
													},
												},
												WorkingDir: workingDir,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "devbuild",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Exec: &v1.ExecCommandParentOverride{
								CommandLine: overrideBuild,
								Component:   overrideComponent0,
								LabeledCommandParentOverride: v1.LabeledCommandParentOverride{
									BaseCommandParentOverride: v1.BaseCommandParentOverride{
										Group: &v1.CommandGroupParentOverride{
											IsDefault: true,
											Kind:      v1.CommandGroupKindParentOverride(v1.BuildCommandGroupKind),
										},
									},
								},
								WorkingDir: overrideWorkingDir,
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: overrideBuild,
												Component:   overrideComponent0,
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Group: &v1.CommandGroup{
															IsDefault: true,
															Kind:      v1.BuildCommandGroupKind,
														},
													},
												},
												WorkingDir: overrideWorkingDir,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 2: append/override a command's list fields based on the key",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Attributes: map[string]string{
															"key-0": "value-0",
														},
													},
												},
												Env: []v1.EnvVar{
													testingutil.GetFakeEnv("env-0", "value-0"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "devbuild",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Exec: &v1.ExecCommandParentOverride{
								LabeledCommandParentOverride: v1.LabeledCommandParentOverride{
									BaseCommandParentOverride: v1.BaseCommandParentOverride{
										Attributes: map[string]string{
											"key-1": "value-1",
										},
									},
								},
								Env: []v1.EnvVarParentOverride{
									testingutil.GetFakeEnvParentOverride("env-0", "value-0-0"),
									testingutil.GetFakeEnvParentOverride("env-1", "value-1"),
								},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Attributes: map[string]string{
															"key-0": "value-0",
															"key-1": "value-1",
														},
													},
												},
												Env: []v1.EnvVar{
													testingutil.GetFakeEnv("env-0", "value-0-0"),
													testingutil.GetFakeEnv("env-1", "value-1"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 3: if multiple, override the correct command",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineRun,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "devbuild",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Exec: &v1.ExecCommandParentOverride{
								CommandLine: overrideBuild,
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: overrideBuild,
											},
										},
									},
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineRun,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 4: throw error if command to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												Env: []v1.EnvVar{
													testingutil.GetFakeEnv("env-0", "value-0"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "devbuild-custom",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Exec: &v1.ExecCommandParentOverride{
								Env: []v1.EnvVarParentOverride{
									testingutil.GetFakeEnvParentOverride("env-0", "value-0-0"),
									testingutil.GetFakeEnvParentOverride("env-1", "value-1"),
								},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "case 5: override a composite command",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "exec1",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "exec2",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "exec3",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "mycomposite",
										CommandUnion: v1.CommandUnion{
											Composite: &v1.CompositeCommand{
												Commands: []string{"exec1", "exec2"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "mycomposite",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Composite: &v1.CompositeCommandParentOverride{
								Commands: []string{"exec1", "exec3"},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "exec1",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "exec2",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "exec3",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: commandLineBuild,
											},
										},
									},
									{
										Id: "mycomposite",
										CommandUnion: v1.CommandUnion{
											Composite: &v1.CompositeCommand{
												Commands: []string{"exec1", "exec3"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 6: throw error if trying to overide command with different type",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												Env: []v1.EnvVar{
													testingutil.GetFakeEnv("env-0", "value-0-0"),
												},
												CommandLine: commandLineBuild,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1.CommandParentOverride{
					{
						Id: "devbuild",
						CommandUnionParentOverride: v1.CommandUnionParentOverride{
							Composite: &v1.CompositeCommandParentOverride{},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideCommands(tt.args.overridePatch)

			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideCommands() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

// func TestDevfileObj_OverrideComponents(t *testing.T) {

// 	containerImage0 := "image-0"
// 	containerImage1 := "image-1"

// 	overrideContainerImage := "image-0-override"

// 	type args struct {
// 		overridePatch []v1.Component
// 	}
// 	tests := []struct {
// 		name           string
// 		devFileObj     DevfileObj
// 		args           args
// 		wantDevFileObj DevfileObj
// 		wantErr        bool
// 	}{
// 		{
// 			name: "case 1: override a container's non list/map fields",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Args:          []string{"arg-0", "arg-1"},
// 									Command:       []string{"cmd-0", "cmd-1"},
// 									Image:         containerImage0,
// 									MemoryLimit:   "512Mi",
// 									MountSources:  false,
// 									Name:          "nodejs",
// 									SourceMapping: "/source",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Component{
// 					{
// 						Container: &v1.ContainerComponent{
// 							Container: v1.Container{
// 								Args:          []string{"arg-0-0", "arg-1-1"},
// 								Command:       []string{"cmd-0-0", "cmd-1-1"},
// 								Image:         overrideContainerImage,
// 								MemoryLimit:   "1Gi",
// 								MountSources:  true,
// 								Name:          "nodejs",
// 								SourceMapping: "/data",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Args:          []string{"arg-0-0", "arg-1-1"},
// 									Command:       []string{"cmd-0-0", "cmd-1-1"},
// 									Image:         overrideContainerImage,
// 									MemoryLimit:   "1Gi",
// 									MountSources:  true,
// 									Name:          "nodejs",
// 									SourceMapping: "/data",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "case 2: append/override a command's list fields based on the key",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Endpoints: []v1.Endpoint{
// 									{
// 										Attributes: map[string]string{
// 											"key-0": "value-0",
// 											"key-1": "value-1",
// 										},
// 										Name:       "endpoint-0",
// 										TargetPort: 8080,
// 									},
// 								},
// 								Container: v1.Container{
// 									Env: []v1.EnvVar{
// 										testingutil.GetFakeEnv("env-0", "value-0"),
// 									},
// 									Name: "nodejs",
// 									VolumeMounts: []v1.VolumeMount{
// 										testingutil.GetFakeVolumeMount("volume-0", "path-0"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Component{
// 					{
// 						Container: &v1.ContainerComponent{
// 							Endpoints: []v1.Endpoint{
// 								{
// 									Attributes: map[string]string{
// 										"key-1":      "value-1-1",
// 										"key-append": "value-append",
// 									},
// 									Name:       "endpoint-0",
// 									TargetPort: 9090,
// 								},
// 								{
// 									Attributes: map[string]string{
// 										"key-0": "value-0",
// 									},
// 									Name:       "endpoint-1",
// 									TargetPort: 3000,
// 								},
// 							},
// 							Container: v1.Container{
// 								Env: []v1.EnvVar{
// 									testingutil.GetFakeEnv("env-0", "value-0-0"),
// 									testingutil.GetFakeEnv("env-1", "value-1"),
// 								},
// 								Name: "nodejs",
// 								VolumeMounts: []v1.VolumeMount{
// 									testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
// 									testingutil.GetFakeVolumeMount("volume-1", "path-1"),
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Env: []v1.EnvVar{
// 										testingutil.GetFakeEnv("env-0", "value-0-0"),
// 										testingutil.GetFakeEnv("env-1", "value-1"),
// 									},
// 									Name: "nodejs",
// 									VolumeMounts: []v1.VolumeMount{
// 										testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
// 										testingutil.GetFakeVolumeMount("volume-1", "path-1"),
// 									},
// 								},
// 								Endpoints: []v1.Endpoint{
// 									{
// 										Attributes: map[string]string{
// 											"key-0":      "value-0",
// 											"key-1":      "value-1-1",
// 											"key-append": "value-append",
// 										},
// 										Name:       "endpoint-0",
// 										TargetPort: 9090,
// 									},
// 									{
// 										Attributes: map[string]string{
// 											"key-0": "value-0",
// 										},
// 										Name:       "endpoint-1",
// 										TargetPort: 3000,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "case 3: if multiple, override the correct command",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Image: containerImage0,
// 									Name:  "nodejs",
// 								},
// 							},
// 						},
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Image: containerImage1,
// 									Name:  "runtime",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Component{
// 					{
// 						Container: &v1.ContainerComponent{
// 							Container: v1.Container{
// 								Image: overrideContainerImage,
// 								Name:  "nodejs",
// 							},
// 						},
// 					},
// 					{
// 						Container: &v1.ContainerComponent{
// 							Container: v1.Container{
// 								Image: containerImage1,
// 								Name:  "runtime",
// 							},
// 						}},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Image: overrideContainerImage,
// 									Name:  "nodejs",
// 								},
// 							},
// 						},
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Image: containerImage1,
// 									Name:  "runtime",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "case 4: throw error if component to override is not found",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Components: []v1.Component{
// 						{
// 							Container: &v1.ContainerComponent{
// 								Container: v1.Container{
// 									Image: containerImage0,
// 									Name:  "nodejs",
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Component{
// 					{
// 						Container: &v1.ContainerComponent{
// 							Container: v1.Container{
// 								Image: containerImage0,
// 								Name:  "nodejs-custom",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{},
// 			wantErr:        true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		if tt.name != "case 1: override a container's non list/map fields" {
// 			continue
// 		}
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := tt.devFileObj.OverrideComponents(tt.args.overridePatch)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("OverrideComponents() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			if tt.wantErr && err != nil {
// 				return
// 			}

// 			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
// 				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
// 			}
// 		})
// 	}
// }

// func TestDevfileObj_OverrideEvents(t *testing.T) {
// 	type args struct {
// 		overridePatch v1.Events
// 	}
// 	tests := []struct {
// 		name           string
// 		devFileObj     DevfileObj
// 		args           args
// 		wantDevFileObj DevfileObj
// 		wantErr        bool
// 	}{
// 		{
// 			name: "case 1: override the events",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Events: v1.Events{
// 						WorkspaceEvents: v1.WorkspaceEvents{
// 							PostStart: []string{"post-start-0", "post-start-1"},
// 							PostStop:  []string{"post-stop-0", "post-stop-1"},
// 							PreStart:  []string{"pre-start-0", "pre-start-1"},
// 							PreStop:   []string{"pre-stop-0", "pre-stop-1"},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: v1.Events{
// 					WorkspaceEvents: v1.WorkspaceEvents{
// 						PostStart: []string{"override-post-start-0", "override-post-start-1"},
// 						PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
// 						PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
// 						PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Events: v1.Events{
// 						WorkspaceEvents: v1.WorkspaceEvents{
// 							PostStart: []string{"override-post-start-0", "override-post-start-1"},
// 							PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
// 							PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
// 							PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "case 2: override some of the events",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Events: v1.Events{
// 						WorkspaceEvents: v1.WorkspaceEvents{
// 							PostStart: []string{"post-start-0", "post-start-1"},
// 							PostStop:  []string{"post-stop-0", "post-stop-1"},
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: v1.Events{
// 					WorkspaceEvents: v1.WorkspaceEvents{
// 						PostStart: []string{"override-post-start-0", "override-post-start-1"},
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Events: v1.Events{
// 						WorkspaceEvents: v1.WorkspaceEvents{
// 							PostStart: []string{"override-post-start-0", "override-post-start-1"},
// 							PostStop:  []string{"post-stop-0", "post-stop-1"},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.devFileObj.OverrideEvents(tt.args.overridePatch); (err != nil) != tt.wantErr {
// 				t.Errorf("OverrideEvents() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
// 				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
// 			}
// 		})
// 	}
// }

// func TestDevfileObj_OverrideProjects(t *testing.T) {
// 	projectName0 := "project-0"
// 	projectName1 := "project-1"

// 	type args struct {
// 		overridePatch []v1.Project
// 	}
// 	tests := []struct {
// 		name           string
// 		devFileObj     DevfileObj
// 		wantDevFileObj DevfileObj
// 		args           args
// 		wantErr        bool
// 	}{
// 		{
// 			name: "case 1: override a project's fields",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Projects: []v1.Project{
// 						{
// 							ClonePath: "/data",
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "master",
// 										},
// 									},
// 								},
// 								Zip: nil,
// 							},
// 							Name: projectName0,
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Project{
// 					{
// 						ClonePath: "/source",
// 						ProjectSource: v1.ProjectSource{
// 							Github: &v1.GithubProjectSource{
// 								GitLikeProjectSource: v1.GitLikeProjectSource{
// 									CheckoutFrom: &v1.CheckoutFrom{
// 										Revision: "release-1.0.0",
// 									},
// 								},
// 							},
// 							Zip: nil,
// 						},
// 						Name: projectName0,
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Projects: []v1.Project{
// 						{
// 							ClonePath: "/source",
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "release-1.0.0",
// 										},
// 									},
// 								},
// 								Zip: nil,
// 							},
// 							Name: projectName0,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "case 2: if multiple, override the correct project",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Projects: []v1.Project{
// 						{
// 							ClonePath: "/data",
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "master",
// 										},
// 									},
// 								},
// 								Zip: nil,
// 							},
// 							Name: projectName0,
// 						},
// 						{
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "master",
// 										},
// 									},
// 								},
// 							},
// 							Name: projectName1,
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Project{
// 					{
// 						ClonePath: "/source",
// 						ProjectSource: v1.ProjectSource{
// 							Github: &v1.GithubProjectSource{
// 								GitLikeProjectSource: v1.GitLikeProjectSource{
// 									CheckoutFrom: &v1.CheckoutFrom{
// 										Revision: "release-1.0.0",
// 									},
// 								},
// 							},
// 							Zip: nil,
// 						},
// 						Name: projectName0,
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Projects: []v1.Project{
// 						{
// 							ClonePath: "/source",
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "release-1.0.0",
// 										},
// 									},
// 								},
// 								Zip: nil,
// 							},
// 							Name: projectName0,
// 						},
// 						{
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "master",
// 										},
// 									},
// 								},
// 							},
// 							Name: projectName1,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "case 3: throw error if project to override is not found",
// 			devFileObj: DevfileObj{
// 				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
// 				Data: &v200.Devfile200{
// 					Projects: []v1.Project{
// 						{
// 							ClonePath: "/data",
// 							ProjectSource: v1.ProjectSource{
// 								Github: &v1.GithubProjectSource{
// 									GitLikeProjectSource: v1.GitLikeProjectSource{
// 										CheckoutFrom: &v1.CheckoutFrom{
// 											Revision: "master",
// 										},
// 									},
// 								},
// 								Zip: nil,
// 							},
// 							Name: projectName0,
// 						},
// 					},
// 				},
// 			},
// 			args: args{
// 				overridePatch: []v1.Project{
// 					{
// 						ClonePath: "/source",
// 						ProjectSource: v1.ProjectSource{
// 							Github: &v1.GithubProjectSource{
// 								GitLikeProjectSource: v1.GitLikeProjectSource{
// 									CheckoutFrom: &v1.CheckoutFrom{
// 										Revision: "release-1.0.0",
// 									},
// 								},
// 							},
// 							Zip: nil,
// 						},
// 						Name: "custom-project",
// 					},
// 				},
// 			},
// 			wantDevFileObj: DevfileObj{},
// 			wantErr:        true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := tt.devFileObj.OverrideProjects(tt.args.overridePatch)

// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("OverrideProjects() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if tt.wantErr && err != nil {
// 				return
// 			}

// 			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
// 				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
// 			}
// 		})
// 	}
// }
