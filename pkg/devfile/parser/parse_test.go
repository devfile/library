package parser

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/ghodss/yaml"
	"github.com/kylelemons/godebug/pretty"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const schemaV200 = "2.0.0"

func Test_parseParentAndPluginFromURI(t *testing.T) {

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
			name: "it should override the requested parent's data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "handle a parent'data without any local override and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should error out when the override is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same parent command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same parent component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "should not have error if the same event is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the parent project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should merge the plugin's uri data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should override the plugin's data with local overrides and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should error out when the plugin devfile is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same plugin command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same plugin component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the plugin project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same project is defined in the both plugin devfile and parent",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same command is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the same component is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should override the requested parent's data and plugin's data, and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the plugin component is defined with a different component type in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should override with no errors if the plugin component is defined with a different component type in the plugin override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the parent component is defined with a different component type in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "it should override with no errors if the parent component is defined with a different component type in the parent override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			name: "error out if the URI is recursively referenced",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
			err := parseParentAndPlugin(tt.args.devFileObj, &resolutionContextTree{}, resolverTools{})

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
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
		Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
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
		err := parseParentAndPlugin(devFileObj, &resolutionContextTree{}, resolverTools{})
		// devfile has an cycle in references: main devfile -> uri: http://127.0.0.1:8080 -> uri: http://127.0.0.1:9090 -> uri: http://127.0.0.1:8090 -> uri: http://127.0.0.1:8080
		expectedErr := fmt.Sprintf("devfile has an cycle in references: main devfile -> uri: %s%s -> uri: %s%s -> uri: %s%s -> uri: %s%s", httpPrefix, uri1,
			httpPrefix, uri2, httpPrefix, uri3, httpPrefix, uri1)
		// Unexpected error
		if err == nil || !reflect.DeepEqual(expectedErr, err.Error()) {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference_withMultipleURI() unexpected error = %v", err)
			return
		}

	})
}

func Test_parseParentFromRegistry(t *testing.T) {
	const validRegistry = "127.0.0.1:8080"
	const invalidRegistry = "invalid-registry.io"
	tool := resolverTools{
		registryURLs: []string{"http://" + validRegistry},
	}
	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "parent-runtime",
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
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "/devfiles/nodejs") {
			data, err = yaml.Marshal(parentDevfile.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", validRegistry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	mainDevfileContent := v1.Devfile{
		DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
			Parent: &v1.Parent{
				ImportReference: v1.ImportReference{
					RegistryUrl: "http://" + validRegistry,
					ImportReferenceUnion: v1.ImportReferenceUnion{
						Id: "nodejs",
					},
				},
				ParentOverrides: v1.ParentOverrides{
					Components: []v1.ComponentParentOverride{
						{
							Name: "parent-runtime",
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
						Name: "runtime2",
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
	}
	wantDevfileContent := v1.Devfile{
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
						Name: "parent-runtime",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
					{
						Name: "runtime2",
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
						PostStart: []string{},
						PostStop:  []string{"post-stop"},
						PreStop:   []string{},
						PreStart:  []string{},
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
	}

	tests := []struct {
		name                   string
		mainDevfile            DevfileObj
		registryURI            string
		wantDevFile            DevfileObj
		wantErr                bool
		testRecursiveReference bool
	}{
		{
			name: "it should override the requested parent's data from provided registryURL and add the local devfile's data",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: mainDevfileContent,
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: wantDevfileContent,
				},
			},
		},
		{
			name: "it should override the requested parent's data from registryURLs set in context and add the local devfile's data",
			mainDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: mainDevfileContent,
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: wantDevfileContent,
				},
			},
		},
		{
			name: "it should error out with invalid registry provided",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "nodejs",
									},
									RegistryUrl: invalidRegistry,
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "it should error out with non-exist registry id provided",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "not-exist",
									},
								},
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := parseParentAndPlugin(tt.mainDevfile, &resolutionContextTree{}, tool)

			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParentAndPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Expected error and got an err
			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.mainDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.mainDevfile.Data, pretty.Compare(tt.mainDevfile.Data, tt.wantDevFile.Data))
			}

		})
	}
}

func Test_parseFromURI(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const httpPrefix = "http://"
	const localRelativeURI = "testTmp/dir/devfile.yaml"
	const notExistURI = "notexist/devfile.yaml"
	const invalidURL = "http//invalid.com"
	uri2 := path.Join(uri1, localRelativeURI)

	localDevfile := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(localRelativeURI),
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
											Image: "nodejs",
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

	// prepare for local file
	err := os.MkdirAll(path.Dir(localRelativeURI), 0755)
	if err != nil {
		t.Errorf("failed to create folder: %v, error: %v", path.Dir(localRelativeURI), err)
		return
	}
	yamlData, err := yaml.Marshal(localDevfile.Data)
	if err != nil {
		t.Errorf("failed to marshall devfile data: %v", err)
		return
	}
	err = ioutil.WriteFile(localRelativeURI, yamlData, 0644)
	if err != nil {
		t.Errorf("fail to write to file: %v", err)
		return
	}
	defer os.RemoveAll("testTmp/")

	parentDevfile := DevfileObj{
		Ctx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Uri: localRelativeURI,
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
	relativeParentDevfile := DevfileObj{
		Ctx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri2),
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
	}

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "notexist") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "devfile.yaml") {
			data, err = yaml.Marshal(relativeParentDevfile.Data)
		} else {
			data, err = yaml.Marshal(parentDevfile.Data)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name            string
		curDevfileCtx   devfileCtx.DevfileCtx
		importReference v1.ImportReference
		wantDevFile     DevfileObj
		wantErr         bool
	}{
		{
			name:          "should be able to parse from relative uri on local disk",
			curDevfileCtx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
			wantDevFile:   localDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: localRelativeURI,
				},
			},
		},
		{
			name:          "should be able to parse relative uri from URL",
			curDevfileCtx: parentDevfile.Ctx,
			wantDevFile:   relativeParentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: localRelativeURI,
				},
			},
		},
		{
			name:          "should fail if no path or url has been set for devfile ctx",
			curDevfileCtx: devfileCtx.DevfileCtx{},
			wantErr:       true,
		},
		{
			name:          "should fail if file not exist",
			curDevfileCtx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: true,
		},
		{
			name:          "should fail if url not exist",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: true,
		},
		{
			name:          "should fail if with invalid URI format",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: invalidURL,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if the main devfile is from local, need to set absolute path
			if tt.curDevfileCtx.GetURL() == "" {
				err := tt.curDevfileCtx.SetAbsPath()
				if err != nil {
					t.Errorf("Test_parseFromURI() unexpected error = %v", err)
					return
				}
			}
			got, err := parseFromURI(tt.importReference, tt.curDevfileCtx, &resolutionContextTree{}, resolverTools{})
			if tt.wantErr == (err == nil) {
				t.Errorf("Test_parseFromURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			}
		})
	}
}

func Test_parseFromRegistry(t *testing.T) {
	const (
		registry        = "127.0.0.1:8080"
		httpPrefix      = "http://"
		notExistId      = "notexist"
		invalidRegistry = "http//invalid.com"
		registryId      = "nodejs"
	)

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaV200,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
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

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "/devfiles/"+registryId) {
			data, err = yaml.Marshal(parentDevfile.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", registry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	tests := []struct {
		name            string
		curDevfileCtx   devfileCtx.DevfileCtx
		importReference v1.ImportReference
		tool            resolverTools
		wantDevFile     DevfileObj
		wantErr         bool
	}{
		{
			name:        "should fail if provided registryUrl does not have protocol prefix",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: registry,
			},
			wantErr: true,
		},
		{
			name:        "should be able to parse from provided registryUrl with prefix",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: httpPrefix + registry,
			},
		},
		{
			name:        "should be able to parse from registry URL defined in tool",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
			},
			tool: resolverTools{
				registryURLs: []string{"http://" + registry},
			},
		},
		{
			name: "should fail if registryId does not exist",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + registry,
			},
			wantErr: true,
		},
		{
			name: "should fail if registryUrl is not provided, and no registry URLs has been set in tool",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
			},
			wantErr: true,
		},
		{
			name: "should fail if registryUrl is invalid",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + invalidRegistry,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFromRegistry(tt.importReference, &resolutionContextTree{}, tt.tool)
			if tt.wantErr == (err == nil) {
				t.Errorf("Test_parseFromRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			}
		})
	}
}

func Test_parseFromKubeCRD(t *testing.T) {
	const (
		namespace  = "default"
		name       = "test-parent-k8s"
		apiVersion = "testgroup/v1alpha2"
	)
	parentSpec := v1.DevWorkspaceTemplateSpec{
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
	}
	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevWorkspaceTemplateSpec: parentSpec,
			},
		},
	}

	tests := []struct {
		name                  string
		curDevfileCtx         devfileCtx.DevfileCtx
		importReference       v1.ImportReference
		devWorkspaceResources map[string]v1.DevWorkspaceTemplate
		errors                map[string]string
		wantDevFile           DevfileObj
		wantErr               bool
	}{
		{
			name:        "should successfully parse the parent with namespace specified in devfile",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Kubernetes: &v1.KubernetesCustomResourceImportReference{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec,
				},
			},
			wantErr: false,
		},
		{
			name:        "should fail if kclient get returns error",
			wantDevFile: parentDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Kubernetes: &v1.KubernetesCustomResourceImportReference{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{},
			errors: map[string]string{
				name: "not found",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testK8sClient := &testingutil.FakeK8sClient{
				DevWorkspaceResources: tt.devWorkspaceResources,
				Errors:                tt.errors,
			}
			tool := resolverTools{
				k8sClient: testK8sClient,
				context:   context.Background(),
			}
			got, err := parseFromKubeCRD(tt.importReference, &resolutionContextTree{}, tool)
			if tt.wantErr == (err == nil) {
				t.Errorf("Test_parseFromKubeCRD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			}
		})
	}
}
