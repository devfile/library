package parser

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/ghodss/yaml"
	"github.com/kylelemons/godebug/pretty"
)

const schemaV200 = "2.0.0"
const devfileTempPath = "devfile.yaml"

func Test_parseParentAndPlugin(t *testing.T) {

	type args struct {
		devFileObj DevfileObj
	}
	tests := []struct {
		name                   string
		args                   args
		parentDevfile          DevfileObj
		pluginDevfile          DevfileObj
		pluginOverride         v1.PluginOverrides
		wantDevFile            DevfileObj
		wantErr                bool
		testRecursiveReference bool
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
			parentDevfile: DevfileObj{
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
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
											},
										},
									},
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
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
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
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
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
			parentDevfile: DevfileObj{
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
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects",
											},
										},
									},
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
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
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
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
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
			parentDevfile: DevfileObj{
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
			parentDevfile: DevfileObj{
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
			parentDevfile: DevfileObj{
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
			name: "case 6: should not have error if the same event is defined again in the local devfile",
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
			parentDevfile: DevfileObj{
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
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									WorkspaceEvents: v1.WorkspaceEvents{
										PostStop:  []string{"post-stop"},
										PreStart:  []string{},
										PreStop:   []string{},
										PostStart: []string{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 7: error out if the parent project is defined again in the local devfile",
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
			parentDevfile: DevfileObj{
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
		{
			name: "case 8: it should merge the plugin's uri data and add the local devfile's data",
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
			pluginDevfile: DevfileObj{
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
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects",
											},
										},
									},
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
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
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
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
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
		{
			name: "case 9: it should override the plugin's data with local overrides and add the local devfile's data",
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
											PostStop: []string{"post-stop-1"},
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
			pluginDevfile: DevfileObj{
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
										PostStop:  []string{"post-stop-2"},
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
			pluginOverride: v1.PluginOverrides{
				OverridesBase: v1.OverridesBase{},
				Components: []v1.ComponentPluginOverride{
					{
						Name: "nodejs",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
				Commands: []v1.CommandPluginOverride{
					{
						Id: "devrun",
						CommandUnionPluginOverride: v1.CommandUnionPluginOverride{
							Exec: &v1.ExecCommandPluginOverride{
								WorkingDir:  "/projects-new",
								CommandLine: "npm build",
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
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm build",
												WorkingDir:  "/projects-new",
											},
										},
									},
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
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
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
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop-1", "post-stop-2"},
										PreStop:   []string{},
										PreStart:  []string{},
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
		{
			name: "case 10: it should error out when the plugin devfile is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
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
			pluginOverride: v1.PluginOverrides{
				Commands: []v1.CommandPluginOverride{
					{
						Id: "devrun",
						CommandUnionPluginOverride: v1.CommandUnionPluginOverride{
							Exec: &v1.ExecCommandPluginOverride{
								WorkingDir: "/projects/nodejs-starter",
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
			name: "case 11: error out if the same plugin command is defined again in the local devfile",
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
			pluginDevfile: DevfileObj{
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
			name: "case 12: error out if the same plugin component is defined again in the local devfile",
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
			pluginDevfile: DevfileObj{
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
			name: "case 13: error out if the plugin project is defined again in the local devfile",
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
			pluginDevfile: DevfileObj{
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
		{
			name: "case 14: error out if the same project is defined in the both plugin devfile and parent",
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
			pluginDevfile: DevfileObj{
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
										Name:      "nodejs-starter",
									},
								},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
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
										Name:      "nodejs-starter",
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
			name: "case 15: error out if the same command is defined in both plugin devfile and parent devfile",
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
			pluginDevfile: DevfileObj{
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
			parentDevfile: DevfileObj{
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
			name: "case 16: error out if the same component is defined in both plugin devfile and parent devfile",
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
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "build",
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
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "build",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-10",
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
			name: "case 17: it should override the requested parent's data and plugin's data, and add the local devfile's data",
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
			parentDevfile: DevfileObj{
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
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Id: "devdebug",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm debug",
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
										PreStart: []string{"pre-start-0"},
									},
								},
							},
						},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name: "nodejs",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
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
										Id: "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
											},
										},
									},
									{
										Id: "devdebug",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												WorkingDir:  "/projects",
												CommandLine: "npm debug",
											},
										},
									},
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
										Name: "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
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
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{"pre-start-0"},
									},
								},
								Projects: []v1.Project{
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
		{
			name: "case 18: error out if the plugin component is defined with a different component type in the local devfile",
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
			pluginDevfile: DevfileObj{
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
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
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
			name: "case 19: it should override with no errors if the plugin component is defined with a different component type in the plugin override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{},
					},
				},
			},
			pluginOverride: v1.PluginOverrides{
				Components: []v1.ComponentPluginOverride{
					{
						Name: "runtime",
						ComponentUnionPluginOverride: v1.ComponentUnionPluginOverride{
							Container: &v1.ContainerComponentPluginOverride{
								ContainerPluginOverride: v1.ContainerPluginOverride{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
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
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
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
			wantErr: false,
		},
		{
			name: "case 20: error out if the parent component is defined with a different component type in the local devfile",
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
			parentDevfile: DevfileObj{
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
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
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
			name: "case 21: it should override with no errors if the parent component is defined with a different component type in the parent override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Components: []v1.ComponentParentOverride{
											{
												Name: "runtime",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Container: &v1.ContainerComponentParentOverride{
														ContainerParentOverride: v1.ContainerParentOverride{
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
			},
			parentDevfile: DevfileObj{
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
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
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
			wantErr: false,
		},
		{
			name: "case 22: error out if the URI is recursively referenced",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaV200,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: "http://127.0.0.1:8080",
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name: "runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
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
			wantErr:                true,
			testRecursiveReference: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.parentDevfile, DevfileObj{}) {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.parentDevfile.Data)
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
			}
			if !reflect.DeepEqual(tt.pluginDevfile, DevfileObj{}) {

				testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.pluginDevfile.Data)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					_, err = w.Write(data)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
				}))
				if tt.testRecursiveReference {
					// create a listener with the desired port.
					l, err := net.Listen("tcp", "127.0.0.1:8080")
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}

					// NewUnstartedServer creates a listener. Close that listener and replace
					// with the one we created.
					testServer.Listener.Close()
					testServer.Listener = l
				}
				testServer.Start()
				defer testServer.Close()

				plugincomp := []v1.Component{
					{
						Name: "plugincomp",
						ComponentUnion: v1.ComponentUnion{
							Plugin: &v1.PluginComponent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: testServer.URL,
									},
								},
								PluginOverrides: tt.pluginOverride,
							},
						},
					},
				}
				tt.args.devFileObj.Data.AddComponents(plugincomp)

			}
			err := parseParentAndPlugin(tt.args.devFileObj)

			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParentAndPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Expected error and got an err
			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.args.devFileObj.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.args.devFileObj.Data, pretty.Compare(tt.args.devFileObj.Data, tt.wantDevFile.Data))
			}

		})
	}
}

func Test_parseParentAndPlugin_RecursivelyReference_withMultipleURI(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const uri2 = "127.0.0.1:9090"
	const uri3 = "127.0.0.1:8090"
	const httpPrefix = "http://"

	devFileObj := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: httpPrefix + uri1,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime2",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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
	parentDevfile1 := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: httpPrefix + uri2,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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
	parentDevfile2 := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "plugin",
								ComponentUnion: v1.ComponentUnion{
									Plugin: &v1.PluginComponent{
										ImportReference: v1.ImportReference{
											ImportReferenceUnion: v1.ImportReferenceUnion{
												Uri: httpPrefix + uri3,
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
	}
	parentDevfile3 := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: httpPrefix + uri1,
							},
						},
					},
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "test",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size: "500Mi",
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

	testServer1 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := yaml.Marshal(parentDevfile1.Data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l1, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer1.Listener.Close()
	testServer1.Listener = l1

	testServer1.Start()
	defer testServer1.Close()

	testServer2 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := yaml.Marshal(parentDevfile2.Data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l2, err := net.Listen("tcp", uri2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer2.Listener.Close()
	testServer2.Listener = l2

	testServer2.Start()
	defer testServer2.Close()

	testServer3 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := yaml.Marshal(parentDevfile3.Data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l3, err := net.Listen("tcp", uri3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer3.Listener.Close()
	testServer3.Listener = l3

	testServer3.Start()
	defer testServer3.Close()
	t.Run("it should error out if URI is recursively referenced with multiple references", func(t *testing.T) {
		err := parseParentAndPlugin(devFileObj)
		expectedErr := fmt.Sprintf("URI %v%v is recursively referenced", httpPrefix, uri1)
		// Unexpected error
		if err == nil || !reflect.DeepEqual(expectedErr, err.Error()) {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference_withMultipleURI() unexpected error = %v", err)
			return
		}

	})
}
