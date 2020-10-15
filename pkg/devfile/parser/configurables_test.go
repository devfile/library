package parser

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v2 "github.com/devfile/parser/pkg/devfile/parser/data/v2"
	"github.com/devfile/parser/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
)

func TestSetConfiguration(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		args           map[string]string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: set memory to 500Mi",
			args: map[string]string{
				"memory": "500Mi",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image:       "quay.io/nodejs-12",
													MemoryLimit: "500Mi",
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "port-3030",
														TargetPort: 3000,
													},
												},
											},
										},
									},
									{
										Name: "loadbalancer",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image:       "quay.io/nginx",
													MemoryLimit: "500Mi",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStop: []string{"post-stop"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
								StarterProjects: []v1.StarterProject{
									{
										Project: v1.Project{
											ClonePath: "/projects",
											Name:      "starter-project-2",
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
			name: "case 2: set ports array",
			args: map[string]string{
				"ports": "8080,8081/UDP,8080/TCP",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "port-3030",
														TargetPort: 3000,
													},
													{
														Name:       "port-8080-tcp",
														TargetPort: 8080,
														Protocol:   "tcp",
													}, {
														Name:       "port-8081-udp",
														TargetPort: 8081,
														Protocol:   "udp",
													},
												},
											},
										},
									},
									{
										Name: "loadbalancer",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nginx",
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "port-8080-tcp",
														TargetPort: 8080,
														Protocol:   "tcp",
													}, {
														Name:       "port-8081-udp",
														TargetPort: 8081,
														Protocol:   "udp",
													},
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStop: []string{"post-stop"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
								StarterProjects: []v1.StarterProject{
									{
										Project: v1.Project{
											ClonePath: "/projects",
											Name:      "starter-project-2",
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
			name: "case 3: set ports array fails due to validation",
			args: map[string]string{
				"ports": "8080,8081/UDP,8083/",
			},
			currentDevfile: testDevfileObj(fs),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for key, value := range tt.args {
				err := tt.currentDevfile.SetConfiguration(key, value)
				if tt.wantErr {
					if err == nil {
						t.Errorf("expected error but got nil")
					}
					// we dont expect an error here
				} else {
					if err != nil {
						t.Errorf("error while setting configuration %+v", err.Error())
					}
				}
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
					t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
				}
			}

		})
	}

}

func TestAddAndRemoveEnvVars(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		listToAdd      []v1.EnvVar
		listToRemove   []string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
	}{
		{
			name: "case 1: add and remove env vars",
			listToAdd: []v1.EnvVar{
				{
					Name:  "DATABASE_PASSWORD",
					Value: "苦痛",
				},
				{
					Name:  "PORT",
					Value: "3003",
				},
				{
					Name:  "PORT",
					Value: "4342",
				},
			},
			listToRemove: []string{
				"PORT",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devbuild",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir: "/projects/nodejs-starter",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
													Env: []v1.EnvVar{
														{
															Name:  "DATABASE_PASSWORD",
															Value: "苦痛",
														},
													},
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "port-3030",
														TargetPort: 3000,
													},
												},
											},
										},
									},
									{
										Name: "loadbalancer",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nginx",
													Env: []v1.EnvVar{
														{
															Name:  "DATABASE_PASSWORD",
															Value: "苦痛",
														},
													},
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStop: []string{"post-stop"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
								StarterProjects: []v1.StarterProject{
									{
										Project: v1.Project{
											ClonePath: "/projects",
											Name:      "starter-project-2",
										},
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

			err := tt.currentDevfile.AddEnvVars(tt.listToAdd)

			if err != nil {
				t.Errorf("error while adding env vars %+v", err.Error())
			}

			err = tt.currentDevfile.RemoveEnvVars(tt.listToRemove)

			if err != nil {
				t.Errorf("error while removing env vars %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
			}

		})
	}

}

func testDevfileObj(fs filesystem.Filesystem) DevfileObj {
	return DevfileObj{
		Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Commands: []v1.Command{
							{
								Id: "devbuild",
								CommandUnion: v1.CommandUnion{
									Exec: &v1.ExecCommand{
										WorkingDir: "/projects/nodejs-starter",
									},
								},
							},
						},
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Container: &v1.ContainerComponent{
										Container: v1.Container{
											Image: "quay.io/nodejs-12",
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "port-3030",
												TargetPort: 3000,
											},
										},
									},
								},
							},
							{
								Name: "loadbalancer",
								ComponentUnion: v1.ComponentUnion{
									Container: &v1.ContainerComponent{
										Container: v1.Container{
											Image: "quay.io/nginx",
										},
									},
								},
							},
						},
						Events: &v1.Events{
							WorkspaceEvents: v1.WorkspaceEvents{
								PostStop: []string{"post-stop"},
							},
						},
						Projects: []v1.Project{
							{
								ClonePath: "/projects",
								Name:      "nodejs-starter-build",
							},
						},
						StarterProjects: []v1.StarterProject{
							{
								Project: v1.Project{
									ClonePath: "/projects",
									Name:      "starter-project-2",
								},
							},
						},
					},
				},
			},
		},
	}
}
