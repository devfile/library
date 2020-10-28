package parser

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/pkg/devfile"
	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v2 "github.com/devfile/parser/pkg/devfile/parser/data/v2"
	"github.com/ghodss/yaml"
	"github.com/kylelemons/godebug/pretty"
)

const schemaV200 = "2.0.0"

func Test_parseParent(t *testing.T) {
	type args struct {
		devFileObj DevfileObj
	}
	tests := []struct {
		name          string
		args          args
		parentDevFile DevfileObj
		wantDevFile   DevfileObj
		wantErr       bool
	}{
		{
			name: "case 1: it should override the requested parent's data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir: "/projects/nodejs-starter",
													},
												},
											},
										},
										Components: []v1.ComponentParentOverride{
											{
												Name: "nodejs",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
															Image: "quay.io/nodejs-12",
														},
													},
												},
											},
										},
										Projects: []v1.ProjectParentOverride{
											{
												ClonePath: "/projects",
												Name:      "nodejs-starter",
											},
										},
									},
								},
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
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm run",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Github: &v1.GithubProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
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
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
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
											},
										},
									},
									{
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
									{
										ClonePath: "/projects",
										ProjectSource: v1.ProjectSource{
											Github: &v1.GithubProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 2: handle a parent'data without any local override and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
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
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm run",
											},
										},
									},
								},
								Components: []v1.Component{
									{
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Github: &v1.GithubProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
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
									{
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects",
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
											},
										},
									},
									{
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Github: &v1.GithubProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"master": "https://githube.com/somerepo/someproject.git",
													},
												},
											},
										},
										Name: "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 3: it should error out when the override is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir: "/projects/nodejs-starter",
													},
												},
											},
										},
										Components: []v1.ComponentParentOverride{
											{
												Name: "nodejs",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
															Image: "quay.io/nodejs-12",
														},
													},
												},
											},
										},
										Projects: []v1.ProjectParentOverride{
											{
												ClonePath: "/projects",
												Name:      "nodejs-starter",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands:   []v1.Command{},
								Components: []v1.Component{},
								Projects:   []v1.Project{},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: true,
		},
		{
			name: "case 4: error out if the same parent command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
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
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
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
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: true,
		},
		{
			name: "case 5: error out if the same parent component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Components: []v1.Component{
										{
											Name: "runtime",
											ComponentUnion: v1.ComponentUnion{
												Container: &v1.ContainerComponent{
													Container: v1.Container{
														Image: "quay.io/nodejs-12",
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
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
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
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: true,
		},
		{
			name: "case 6: error out if the same event is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Events: &v1.Events{
										WorkspaceEvents: v1.WorkspaceEvents{
											PostStop: []string{"post-stop"},
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStop: []string{"post-stop"},
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: true,
		},
		{
			name: "case 7: error out if the same project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
										},
									},
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
									},
								},
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := yaml.Marshal(tt.parentDevFile.Data)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				_, err = w.Write(data)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}))
			defer testServer.Close()

			parent := tt.args.devFileObj.Data.GetParent()
			if parent == nil {
				parent = &v1.Parent{}
			}
			parent.Uri = testServer.URL

			tt.args.devFileObj.Data.SetParent(parent)
			tt.wantDevFile.Data.SetParent(parent)
			err := parseParent(tt.args.devFileObj)

			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Expected error and got an err
			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.args.devFileObj.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.args.devFileObj, pretty.Compare(tt.args.devFileObj.Data, tt.wantDevFile.Data))
			}
		})
	}
}
