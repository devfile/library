package parser

import (
	"context"
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
)

const schemaVersion = string(data.APISchemaVersion220)

var isTrue bool = true
var isFalse bool = false
var apiSchemaVersions = []string{data.APISchemaVersion200.String(), data.APISchemaVersion210.String(), data.APISchemaVersion220.String()}

var defaultDiv testingutil.DockerImageValues = testingutil.DockerImageValues{
	ImageName:    "image:latest",
	Uri:          "/local/image",
	BuildContext: "/src",
}

func Test_parseParentAndPluginFromURI(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const uri2 = "127.0.0.1:9090"
	importFromUri1 := attributes.Attributes{}.PutString(importSourceAttribute, fmt.Sprintf("uri: http://%s", uri1))
	importFromUri2 := attributes.Attributes{}.PutString(importSourceAttribute, fmt.Sprintf("uri: http://%s", uri2))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		fmt.Sprintf("uri: http://%s", uri1)).PutString(parentOverrideAttribute, "main devfile")
	pluginOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		fmt.Sprintf("uri: http://%s", uri2)).PutString(pluginOverrideAttribute, "main devfile")

	divRRTrue := defaultDiv
	divRRTrue.RootRequired = &isTrue

	divRRFalse := divRRTrue
	divRRFalse.RootRequired = &isFalse

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Commands: []v1.Command{
							{
								Id: "devrun",
								CommandUnion: v1.CommandUnion{
									Exec: &v1.ExecCommand{
										WorkingDir:       "/projects",
										CommandLine:      "npm run",
										HotReloadCapable: &isTrue,
									},
								},
							},
							{
								Id: "testrun",
								CommandUnion: v1.CommandUnion{
									Apply: &v1.ApplyCommand{
										LabeledCommand: v1.LabeledCommand{
											BaseCommand: v1.BaseCommand{
												Group: &v1.CommandGroup{
													Kind:      v1.TestCommandGroupKind,
													IsDefault: &isTrue,
												},
											},
										},
									},
								},
							},
							{
								Id: "allcmds",
								CommandUnion: v1.CommandUnion{
									Composite: &v1.CompositeCommand{
										Commands: []string{"testrun", "devrun"},
										Parallel: &isTrue,
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
											Annotation: &v1.Annotation{
												Deployment: map[string]string{
													"deploy-key1": "deploy-value1",
													"deploy-key2": "deploy-value2",
												},
												Service: map[string]string{
													"svc-key1": "svc-value1",
													"svc-key2": "svc-value2",
												},
											},
											Image:        "quay.io/nodejs-10",
											DedicatedPod: &isTrue,
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "log",
												TargetPort: 443,
												Secure:     &isFalse,
												Annotations: map[string]string{
													"ingress-key1": "ingress-value1",
													"ingress-key2": "ingress-value2",
												},
											},
										},
									},
								},
							},
							{
								Name: "volume",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size:      "2Gi",
											Ephemeral: &isFalse,
										},
									},
								},
							},
							{
								Name: "openshift",
								ComponentUnion: v1.ComponentUnion{
									Openshift: &v1.OpenshiftComponent{
										K8sLikeComponent: v1.K8sLikeComponent{
											K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
												Uri: "https://xyz.com/dir/file.yaml",
											},
											Endpoints: []v1.Endpoint{
												{
													Name:       "metrics",
													TargetPort: 8080,
												},
											},
										},
									},
								},
							},
							testingutil.GetDockerImageTestComponent(divRRTrue, nil, nil),
						},
						Events: &v1.Events{
							DevWorkspaceEvents: v1.DevWorkspaceEvents{
								PostStart: []string{"post-start-0"},
							},
						},
						Projects: []v1.Project{
							{
								ClonePath: "/data",
								ProjectSource: v1.ProjectSource{
									Git: &v1.GitProjectSource{
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
	}

	parentCmdAlreadyDefinedErr := "Some Commands are already defined in parent.* If you want to override them, you should do it in the parent scope."
	parentCmpAlreadyDefinedErr := "Some Components are already defined in parent.* If you want to override them, you should do it in the parent scope."
	parentProjectAlreadyDefinedErr := "Some Projects are already defined in parent.* If you want to override them, you should do it in the parent scope."
	pluginCmdAlreadyDefinedErr := "Some Commands are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	pluginCmpAlreadyDefinedErr := "Some Components are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	pluginProjectAlreadyDefinedErr := "Some Projects are already defined in plugin.* If you want to override them, you should do it in the plugin scope."
	newCmdErr := "Some Commands do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	newCmpErr := "Some Components do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	newProjectErr := "Some Projects do not override any existing element.* They should be defined in the main body, as new elements, not in the overriding section"
	importCycleErr := "devfile has an cycle in references: main devfile -> .*"
	parentDevfileVersionErr := "the parent devfile version from .* is greater than the child devfile version from main devfile"
	pluginDevfileVersionErr := "the plugin devfile version from .* is greater than the child devfile version from main devfile"

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
		wantErr                []string
		testRecursiveReference bool
	}{
		{
			name: "it should override the requested parent's data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								Parent: &v1.Parent{
									ParentOverrides: v1.ParentOverrides{
										Commands: []v1.CommandParentOverride{
											{
												Id: "devrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Exec: &v1.ExecCommandParentOverride{
														WorkingDir:       "/projects/nodejs-starter",
														HotReloadCapable: &isFalse,
													},
												},
											},
											{
												Id: "testrun",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Apply: &v1.ApplyCommandParentOverride{
														LabeledCommandParentOverride: v1.LabeledCommandParentOverride{
															BaseCommandParentOverride: v1.BaseCommandParentOverride{
																Group: &v1.CommandGroupParentOverride{
																	Kind:      v1.CommandGroupKindParentOverride(v1.BuildCommandGroupKind),
																	IsDefault: &isFalse,
																},
															},
														},
													},
												},
											},
											{
												Id: "allcmds",
												CommandUnionParentOverride: v1.CommandUnionParentOverride{
													Composite: &v1.CompositeCommandParentOverride{
														Parallel: &isFalse,
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
															Annotation: &v1.AnnotationParentOverride{
																Deployment: map[string]string{
																	"deploy-key2": "deploy-value3",
																	"deploy-key3": "deploy-value3",
																},
																Service: map[string]string{
																	"svc-key2": "svc-value3",
																	"svc-key3": "svc-value3",
																},
															},
															Image:        "quay.io/nodejs-12",
															DedicatedPod: &isFalse,
															MountSources: &isTrue, //overrides an unset value to true
														},
														Endpoints: []v1.EndpointParentOverride{
															{
																Name:       "log",
																TargetPort: 443,
																Secure:     &isTrue,
																Annotations: map[string]string{
																	"ingress-key2": "ingress-value3",
																	"ingress-key3": "ingress-value3",
																},
															},
														},
													},
												},
											},
											{
												Name: "volume",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Volume: &v1.VolumeComponentParentOverride{
														VolumeParentOverride: v1.VolumeParentOverride{
															Size:      "2Gi",
															Ephemeral: &isTrue,
														},
													},
												},
											},
											{
												Name: "openshift",
												ComponentUnionParentOverride: v1.ComponentUnionParentOverride{
													Openshift: &v1.OpenshiftComponentParentOverride{
														K8sLikeComponentParentOverride: v1.K8sLikeComponentParentOverride{
															Endpoints: []v1.EndpointParentOverride{
																{
																	Name:       "metrics",
																	TargetPort: 8080,
																	Secure:     &isFalse,
																},
															},
														},
													},
												},
											},
											testingutil.GetDockerImageTestComponentParentOverride(divRRFalse),
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
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
			parentDevfile: parentDevfile,
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine:      "npm run",
												WorkingDir:       "/projects/nodejs-starter",
												HotReloadCapable: &isFalse,
											},
										},
									},
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "testrun",
										CommandUnion: v1.CommandUnion{
											Apply: &v1.ApplyCommand{
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Group: &v1.CommandGroup{
															Kind:      v1.BuildCommandGroupKind,
															IsDefault: &isFalse,
														},
													},
												},
											},
										},
									},
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "allcmds",
										CommandUnion: v1.CommandUnion{
											Composite: &v1.CompositeCommand{
												Commands: []string{"testrun", "devrun"},
												Parallel: &isFalse,
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
										Attributes: parentOverridesFromMainDevfile,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value3",
															"deploy-key3": "deploy-value3",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value3",
															"svc-key3": "svc-value3",
														},
													},
													Image:        "quay.io/nodejs-12",
													DedicatedPod: &isFalse,
													MountSources: &isTrue,
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "log",
														TargetPort: 443,
														Secure:     &isTrue,
														Annotations: map[string]string{
															"ingress-key1": "ingress-value1",
															"ingress-key2": "ingress-value3",
															"ingress-key3": "ingress-value3",
														},
													},
												},
											},
										},
									},
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "volume",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size:      "2Gi",
													Ephemeral: &isTrue,
												},
											},
										},
									},
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "openshift",
										ComponentUnion: v1.ComponentUnion{
											Openshift: &v1.OpenshiftComponent{
												K8sLikeComponent: v1.K8sLikeComponent{
													K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
														Uri: "https://xyz.com/dir/file.yaml",
													},
													Endpoints: []v1.Endpoint{
														{
															Name:       "metrics",
															TargetPort: 8080,
															Secure:     &isFalse,
														},
													},
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(divRRFalse, nil, parentOverridesFromMainDevfile),
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: parentOverridesFromMainDevfile,
										ClonePath:  "/projects",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
			name: "handle a parent's data without any local override and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
											Name: "Kubernetes",
											ComponentUnion: v1.ComponentUnion{
												Kubernetes: &v1.KubernetesComponent{
													K8sLikeComponent: v1.K8sLikeComponent{
														K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
															Uri: "/devfiles",
														},
														Endpoints: []v1.Endpoint{
															{
																Name:       "messages",
																TargetPort: 8080,
																Secure:     &isTrue,
															},
														},
													},
												},
											},
										},
									},
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
			parentDevfile: parentDevfile,
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: importFromUri1,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine:      "npm run",
												WorkingDir:       "/projects",
												HotReloadCapable: &isTrue,
											},
										},
									},
									{
										Attributes: importFromUri1,
										Id:         "testrun",
										CommandUnion: v1.CommandUnion{
											Apply: &v1.ApplyCommand{
												LabeledCommand: v1.LabeledCommand{
													BaseCommand: v1.BaseCommand{
														Group: &v1.CommandGroup{
															Kind:      v1.TestCommandGroupKind,
															IsDefault: &isTrue,
														},
													},
												},
											},
										},
									},
									{
										Attributes: importFromUri1,
										Id:         "allcmds",
										CommandUnion: v1.CommandUnion{
											Composite: &v1.CompositeCommand{
												Commands: []string{"testrun", "devrun"},
												Parallel: &isTrue,
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
										Attributes: importFromUri1,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value2",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value2",
														},
													},
													Image:        "quay.io/nodejs-10",
													DedicatedPod: &isTrue,
												},
												Endpoints: []v1.Endpoint{
													{
														Name:       "log",
														TargetPort: 443,
														Secure:     &isFalse,
														Annotations: map[string]string{
															"ingress-key1": "ingress-value1",
															"ingress-key2": "ingress-value2",
														},
													},
												},
											},
										},
									},
									{
										Attributes: importFromUri1,
										Name:       "volume",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size:      "2Gi",
													Ephemeral: &isFalse,
												},
											},
										},
									},
									{
										Attributes: importFromUri1,
										Name:       "openshift",
										ComponentUnion: v1.ComponentUnion{
											Openshift: &v1.OpenshiftComponent{
												K8sLikeComponent: v1.K8sLikeComponent{
													K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
														Uri: "https://xyz.com/dir/file.yaml",
													},
													Endpoints: []v1.Endpoint{
														{
															Name:       "metrics",
															TargetPort: 8080,
														},
													},
												},
											},
										},
									},
									//no overrides so expected values are the same as the parent
									testingutil.GetDockerImageTestComponent(divRRTrue, nil, importFromUri1),
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
										Name: "Kubernetes",
										ComponentUnion: v1.ComponentUnion{
											Kubernetes: &v1.KubernetesComponent{
												K8sLikeComponent: v1.K8sLikeComponent{
													K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
														Uri: "/devfiles",
													},
													Endpoints: []v1.Endpoint{
														{
															Name:       "messages",
															TargetPort: 8080,
															Secure:     &isTrue,
														},
													},
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri1,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
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
							SchemaVersion: schemaVersion,
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
			wantErr: []string{newCmpErr, newCmdErr, newProjectErr},
		},
		{
			name: "error out if the same parent command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{parentCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same parent component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{parentCmpAlreadyDefinedErr},
		},
		{
			name: "should not have error if the same event is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Events: &v1.Events{
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
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
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
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
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentProjectAlreadyDefinedErr},
		},
		{
			name: "error out if the parent devfile version is greater than main devfile version",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: "2.0.0",
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
							},
						},
					},
				},
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{parentDevfileVersionErr},
		},
		{
			name: "it should merge the plugin's uri data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
							SchemaVersion: schemaVersion,
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
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value2",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value2",
														},
													},
													Image: "quay.io/nodejs-10",
												},
											},
										},
									},
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: importFromUri2,
										Id:         "devrun",
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
										Attributes: importFromUri2,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value2",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value2",
														},
													},
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri2,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
							SchemaVersion: schemaVersion,
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
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value2",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value2",
														},
													},
													Image: "quay.io/nodejs-10",
												},
												Endpoints: []v1.Endpoint{
													{
														Annotations: map[string]string{
															"ingress-key1": "ingress-value1",
															"ingress-key2": "ingress-value2",
														},
														Name:       "url",
														TargetPort: 8080,
													},
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(divRRFalse, nil, nil),
								},
								Events: &v1.Events{
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop-2"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
									Annotation: &v1.AnnotationPluginOverride{
										Deployment: map[string]string{
											"deploy-key2": "deploy-value3",
											"deploy-key3": "deploy-value3",
										},
										Service: map[string]string{
											"svc-key2": "svc-value3",
											"svc-key3": "svc-value3",
										},
									},
									Image: "quay.io/nodejs-12",
								},
								Endpoints: []v1.EndpointPluginOverride{
									{
										Annotations: map[string]string{
											"ingress-key2": "ingress-value3",
											"ingress-key3": "ingress-value3",
										},
										Name:       "url",
										TargetPort: 9090,
									},
								},
							},
						},
					},
					testingutil.GetDockerImageTestComponentPluginOverride(divRRTrue),
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Id:         "devrun",
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
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "nodejs",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Annotation: &v1.Annotation{
														Deployment: map[string]string{
															"deploy-key1": "deploy-value1",
															"deploy-key2": "deploy-value3",
															"deploy-key3": "deploy-value3",
														},
														Service: map[string]string{
															"svc-key1": "svc-value1",
															"svc-key2": "svc-value3",
															"svc-key3": "svc-value3",
														},
													},
													Image: "quay.io/nodejs-12",
												},
												Endpoints: []v1.Endpoint{
													{
														Annotations: map[string]string{
															"ingress-key1": "ingress-value1",
															"ingress-key2": "ingress-value3",
															"ingress-key3": "ingress-value3",
														},
														Name:       "url",
														TargetPort: 9090,
													},
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(divRRTrue, nil, pluginOverridesFromMainDevfile),
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop-1", "post-stop-2"},
										PreStop:   []string{},
										PreStart:  []string{},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: importFromUri2,
										ClonePath:  "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{newCmdErr},
		},
		{
			name: "error out if the same plugin command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{pluginCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same plugin component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "error out if the plugin project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
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
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter-build",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
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
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginProjectAlreadyDefinedErr},
		},
		{
			name: "error out if the same project is defined in the both plugin devfile and parent",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
							DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
								DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
									Projects: []v1.Project{
										{
											ClonePath: "/projects",
											Name:      "nodejs-starter-build",
											ProjectSource: v1.ProjectSource{
												Git: &v1.GitProjectSource{
													GitLikeProjectSource: v1.GitLikeProjectSource{
														Remotes: map[string]string{
															"origin": "url",
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
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
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
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Projects: []v1.Project{
									{
										ClonePath: "/projects",
										Name:      "nodejs-starter",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
												GitLikeProjectSource: v1.GitLikeProjectSource{
													Remotes: map[string]string{
														"origin": "url",
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
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr: []string{pluginProjectAlreadyDefinedErr},
		},
		{
			name: "error out if the same command is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
							SchemaVersion: schemaVersion,
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
			wantErr: []string{pluginCmdAlreadyDefinedErr},
		},
		{
			name: "error out if the same component is defined in both plugin devfile and parent devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
							SchemaVersion: schemaVersion,
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
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "it should override the requested parent's data and plugin's data, and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
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
										DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
							SchemaVersion: schemaVersion,
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										ClonePath: "/data",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
							SchemaVersion: schemaVersion,
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Commands: []v1.Command{
									{
										Attributes: parentOverridesFromMainDevfile,
										Id:         "devrun",
										CommandUnion: v1.CommandUnion{
											Exec: &v1.ExecCommand{
												CommandLine: "npm run",
												WorkingDir:  "/projects/nodejs-starter",
											},
										},
									},
									{
										Attributes: importFromUri2,
										Id:         "devdebug",
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
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "nodejs",
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
									DevWorkspaceEvents: v1.DevWorkspaceEvents{
										PostStart: []string{"post-start-0"},
										PostStop:  []string{"post-stop"},
										PreStop:   []string{},
										PreStart:  []string{"pre-start-0"},
									},
								},
								Projects: []v1.Project{
									{
										Attributes: parentOverridesFromMainDevfile,
										ClonePath:  "/projects",
										ProjectSource: v1.ProjectSource{
											Git: &v1.GitProjectSource{
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
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{pluginCmpAlreadyDefinedErr},
		},
		{
			name: "it should override with no errors if the plugin component is defined with a different component type in the plugin override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
						},
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
							SchemaVersion: schemaVersion,
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Attributes: pluginOverridesFromMainDevfile,
										Name:       "runtime",
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
		{
			name: "error out if the parent component is defined with a different component type in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
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
			},
			parentDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
			wantErr: []string{parentCmpAlreadyDefinedErr},
		},
		{
			name: "it should override with no errors if the parent component is defined with a different component type in the parent override",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: schemaVersion,
							},
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
							SchemaVersion: schemaVersion,
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
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Attributes: parentOverridesFromMainDevfile,
										Name:       "runtime",
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
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: "http://" + uri2,
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
			wantErr:                []string{importCycleErr},
			testRecursiveReference: true,
		},
		{
			name: "error out if the plugin devfile is greater than main devfile version",
			args: args{
				devFileObj: DevfileObj{
					Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
					Data: &v2.DevfileV2{
						Devfile: v1.Devfile{
							DevfileHeader: devfilepkg.DevfileHeader{
								SchemaVersion: "2.0.0",
							},
						},
					},
				},
			},
			pluginDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{},
			},
			wantErr:                []string{pluginDevfileVersionErr},
			testRecursiveReference: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentTestServer *httptest.Server
			var pluginTestServer *httptest.Server
			if !reflect.DeepEqual(tt.parentDevfile, DevfileObj{}) {
				parentTestServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.parentDevfile.Data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while doing yaml marshal: %v", err)
					}
					_, err = w.Write(data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while writing data: %v", err)
					}
				}))
				// create a listener with the desired port.
				l1, err := net.Listen("tcp", uri1)
				if err != nil {
					t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while creating listener: %v", err)
				}

				// NewUnstartedServer creates a listener. Close that listener and replace
				// with the one we created.
				parentTestServer.Listener.Close()
				parentTestServer.Listener = l1

				parentTestServer.Start()
				defer parentTestServer.Close()

				parent := tt.args.devFileObj.Data.GetParent()
				if parent == nil {
					parent = &v1.Parent{}
				}
				parent.Uri = parentTestServer.URL

				tt.args.devFileObj.Data.SetParent(parent)
			}
			if !reflect.DeepEqual(tt.pluginDevfile, DevfileObj{}) {

				pluginTestServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					data, err := yaml.Marshal(tt.pluginDevfile.Data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while doing yaml marshal: %v", err)
					}
					_, err = w.Write(data)
					if err != nil {
						t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while writing data: %v", err)
					}
				}))
				l, err := net.Listen("tcp", uri2)
				if err != nil {
					t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while creating listener: %v", err)
				}

				// NewUnstartedServer creates a listener. Close that listener and replace
				// with the one we created.
				pluginTestServer.Listener.Close()
				pluginTestServer.Listener = l

				pluginTestServer.Start()
				defer pluginTestServer.Close()

				plugincomp := []v1.Component{
					{
						Name: "plugincomp",
						ComponentUnion: v1.ComponentUnion{
							Plugin: &v1.PluginComponent{
								ImportReference: v1.ImportReference{
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Uri: pluginTestServer.URL,
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
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentAndPluginFromURI() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.args.devFileObj.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentAndPluginFromURI() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.args.devFileObj.Data, pretty.Compare(tt.args.devFileObj.Data, tt.wantDevFile.Data))
			} else if err != nil {
				for _, wantErr := range tt.wantErr {
					assert.Regexp(t, wantErr, err.Error(), "Test_parseParentAndPluginFromURI(): Error message should match")
				}
			}
		})
	}
}

func Test_parseParentAndPlugin_RecursivelyReference(t *testing.T) {
	const uri1 = "127.0.0.1:8080"
	const uri2 = "127.0.0.1:8090"
	const httpPrefix = "http://"
	const name = "testcrd"
	const namespace = "defaultnamespace"

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
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					Parent: &v1.Parent{
						ImportReference: v1.ImportReference{
							ImportReferenceUnion: v1.ImportReferenceUnion{
								Kubernetes: &v1.KubernetesCustomResourceImportReference{
									Name:      name,
									Namespace: namespace,
								},
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
					SchemaVersion: schemaVersion,
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
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while doing yaml marshal: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l1, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while creating listener: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer1.Listener.Close()
	testServer1.Listener = l1

	testServer1.Start()
	defer testServer1.Close()

	testServer2 := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		if strings.Contains(r.URL.Path, "/devfiles/nodejs") {
			data, err = yaml.Marshal(parentDevfile2.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l3, err := net.Listen("tcp", uri2)
	if err != nil {
		t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error while creating listener: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer2.Listener.Close()
	testServer2.Listener = l3

	testServer2.Start()
	defer testServer2.Close()

	parentSpec := v1.DevWorkspaceTemplateSpec{
		Parent: &v1.Parent{
			ImportReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: "nodejs",
				},
				RegistryUrl: httpPrefix + uri2,
			},
		},
		DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
			Components: []v1.Component{
				{
					Name: "crdcomponent",
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
	devWorkspaceResources := map[string]v1.DevWorkspaceTemplate{
		name: {
			TypeMeta: kubev1.TypeMeta{
				Kind:       "DevWorkspaceTemplate",
				APIVersion: "testgroup/v1alpha2",
			},
			Spec: parentSpec,
		},
	}

	t.Run("it should error out if import reference has a cycle", func(t *testing.T) {
		testK8sClient := &testingutil.FakeK8sClient{
			DevWorkspaceResources: devWorkspaceResources,
		}
		tool := resolverTools{
			k8sClient: testK8sClient,
			context:   context.Background(),
		}

		err := parseParentAndPlugin(devFileObj, &resolutionContextTree{}, tool)
		// devfile has a cycle in references: main devfile -> uri: http://127.0.0.1:8080 -> name: testcrd, namespace: defaultnamespace -> id: nodejs, registryURL: http://127.0.0.1:8090 -> uri: http://127.0.0.1:8080
		expectedErr := fmt.Sprintf("devfile has an cycle in references: main devfile -> uri: %s%s -> name: %s, namespace: %s -> id: nodejs, registryURL: %s%s -> uri: %s%s", httpPrefix, uri1, name, namespace,
			httpPrefix, uri2, httpPrefix, uri1)
		// Unexpected error
		if err == nil || !reflect.DeepEqual(expectedErr, err.Error()) {
			t.Errorf("Test_parseParentAndPlugin_RecursivelyReference() unexpected error: %v", err)

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

	invalidURLErr := "the provided registryURL: .* is not a valid URL"
	idNotFoundErr := "failed to get id: .* from registry URLs provided"

	parentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
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
							testingutil.GetDockerImageTestComponent(defaultDiv, nil, nil),
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
			t.Errorf("Test_parseParentFromRegistry() unexpected error while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseParentFromRegistry() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", validRegistry)
	if err != nil {
		t.Errorf("Test_parseParentFromRegistry() unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	div := defaultDiv
	div.RootRequired = &isTrue

	mainDevfileContent := v1.Devfile{
		DevfileHeader: devfilepkg.DevfileHeader{
			SchemaVersion: schemaVersion,
		},
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
						testingutil.GetDockerImageTestComponentParentOverride(div),
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
					DevWorkspaceEvents: v1.DevWorkspaceEvents{
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

	importFromRegistry := attributes.Attributes{}.PutString(importSourceAttribute, resolveImportReference(mainDevfileContent.Parent.ImportReference))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		resolveImportReference(mainDevfileContent.Parent.ImportReference)).PutString(parentOverrideAttribute, "main devfile")

	wantDevfileContent := v1.Devfile{
		DevfileHeader: devfilepkg.DevfileHeader{
			SchemaVersion: schemaVersion,
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
				Components: []v1.Component{
					{
						Attributes: parentOverridesFromMainDevfile,
						Name:       "parent-runtime",
						ComponentUnion: v1.ComponentUnion{
							Container: &v1.ContainerComponent{
								Container: v1.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
					testingutil.GetDockerImageTestComponent(div, nil, parentOverridesFromMainDevfile),
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
					DevWorkspaceEvents: v1.DevWorkspaceEvents{
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
		wantErr                *string
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
			name: "it should merge the requested parent's data from provided registryURL if no override is set",
			mainDevfile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: v1.ImportReference{
									RegistryUrl: "http://" + validRegistry,
									ImportReferenceUnion: v1.ImportReferenceUnion{
										Id: "nodejs",
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
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
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
								Components: []v1.Component{
									{
										Attributes: importFromRegistry,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(defaultDiv, nil, importFromRegistry),
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
							},
						},
					},
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
			wantErr: &invalidURLErr,
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
			wantErr: &idNotFoundErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := parseParentAndPlugin(tt.mainDevfile, &resolutionContextTree{}, tool)

			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentFromRegistry() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.mainDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentFromRegistry() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.mainDevfile.Data, pretty.Compare(tt.mainDevfile.Data, tt.wantDevFile.Data))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseParentFromRegistry(): Error message should match")
			}

		})
	}
}

func Test_parseParentFromKubeCRD(t *testing.T) {

	const (
		namespace  = "default"
		name       = "test-parent-k8s"
		apiVersion = "testgroup/v1alpha2"
	)

	kubeCRDReference := v1.ImportReference{
		ImportReferenceUnion: v1.ImportReferenceUnion{
			Kubernetes: &v1.KubernetesCustomResourceImportReference{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	importFromKubeCRD := attributes.Attributes{}.PutString(importSourceAttribute, resolveImportReference(kubeCRDReference))
	parentOverridesFromMainDevfile := attributes.Attributes{}.PutString(importSourceAttribute,
		resolveImportReference(kubeCRDReference)).PutString(parentOverrideAttribute, "main devfile")

	parentSpec := v1.DevWorkspaceTemplateSpec{
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
				testingutil.GetDockerImageTestComponent(defaultDiv, nil, nil),
			},
		},
	}

	//this is a copy of parentSpec which can't be reused because defaults are being set on the SrcType and ImageType properties in the override code.
	parentSpec2 := v1.DevWorkspaceTemplateSpec{
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
				testingutil.GetDockerImageTestComponent(defaultDiv, nil, nil),
			},
		},
	}

	crdNotFoundErr := "not found"

	//override all properties
	div := testingutil.DockerImageValues{
		ImageName:    "image:next",
		Uri:          "/local/image2",
		BuildContext: "/src2",
		RootRequired: &isTrue,
	}

	tests := []struct {
		name                  string
		devWorkspaceResources map[string]v1.DevWorkspaceTemplate
		errors                map[string]string
		mainDevfile           DevfileObj
		wantDevFile           DevfileObj
		wantErr               *string
	}{
		{
			name: "should successfully override the parent data",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
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
										testingutil.GetDockerImageTestComponentParentOverride(div),
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
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
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
										Attributes: parentOverridesFromMainDevfile,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Container: &v1.ContainerComponent{
												Container: v1.Container{
													Image: "quay.io/nodejs-12",
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(div, nil, parentOverridesFromMainDevfile),
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
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec,
				},
			},
		},
		{
			name: "should successfully merge the parent data without override defined",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
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
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
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
										Attributes: importFromKubeCRD,
										Name:       "parent-runtime",
										ComponentUnion: v1.ComponentUnion{
											Volume: &v1.VolumeComponent{
												Volume: v1.Volume{
													Size: "500Mi",
												},
											},
										},
									},
									testingutil.GetDockerImageTestComponent(defaultDiv, nil, importFromKubeCRD),
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
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{
				name: {
					TypeMeta: kubev1.TypeMeta{
						Kind:       "DevWorkspaceTemplate",
						APIVersion: apiVersion,
					},
					Spec: parentSpec2,
				},
			},
		},
		{
			name: "should fail if kclient get returns error",
			mainDevfile: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							Parent: &v1.Parent{
								ImportReference: kubeCRDReference,
							},
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{},
						},
					},
				},
			},
			devWorkspaceResources: map[string]v1.DevWorkspaceTemplate{},
			errors: map[string]string{
				name: crdNotFoundErr,
			},
			wantErr: &crdNotFoundErr,
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
			err := parseParentAndPlugin(tt.mainDevfile, &resolutionContextTree{}, tool)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseParentFromKubeCRD() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(tt.mainDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseParentFromKubeCRD() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile.Data, tt.mainDevfile.Data, pretty.Compare(tt.mainDevfile.Data, tt.wantDevFile.Data))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseParentFromKubeCRD(): Error message should match")
			}

		})
	}
}

func Test_parseFromURI(t *testing.T) {
	const (
		uri1             = "127.0.0.1:8080"
		httpPrefix       = "http://"
		localRelativeURI = "testTmp/dir/devfile.yaml"
		notExistURI      = "notexist/devfile.yaml"
		invalidURL       = "http//invalid.com"
	)
	uri2 := path.Join(uri1, localRelativeURI)

	localDevfile := DevfileObj{
		Ctx: devfileCtx.NewDevfileCtx(localRelativeURI),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Container: &v1.ContainerComponent{
										Container: v1.Container{
											Image:        "nodejs",
											DedicatedPod: &isFalse,
											MountSources: &isTrue,
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

	invalidFilePathErr := "the provided path is not a valid yaml filepath, and devfile.yaml or .devfile.yaml not found in the provided path.*"
	readDevfileErr := "failed to read devfile from path.*"
	URLNotFoundErr := "error getting devfile info from url: failed to retrieve .*, 404: Not Found"
	invalidURLErr := "parse .* invalid URI for request"

	// prepare for local file
	err := os.MkdirAll(path.Dir(localRelativeURI), 0755)
	if err != nil {
		fmt.Errorf("Test_parseFromURI() error: failed to create folder: %v, error: %v", path.Dir(localRelativeURI), err)
	}
	yamlData, err := yaml.Marshal(localDevfile.Data)
	if err != nil {
		fmt.Errorf("Test_parseFromURI() error: failed to marshall devfile data: %v", err)
	}
	err = ioutil.WriteFile(localRelativeURI, yamlData, 0644)
	if err != nil {
		fmt.Errorf("Test_parseFromURI() error: fail to write to file: %v", err)
	}

	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll("testTmp/")

	parentDevfile := DevfileObj{
		Ctx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
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
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime",
								ComponentUnion: v1.ComponentUnion{
									Volume: &v1.VolumeComponent{
										Volume: v1.Volume{
											Size:      "500Mi",
											Ephemeral: &isFalse,
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
			t.Errorf("Test_parseFromURI() unexpected while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseFromURI() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", uri1)
	if err != nil {
		t.Errorf("Test_parseFromURI() unexpected error while creating listener: %v", err)
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
		wantErr         *string
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
			wantErr:       &invalidFilePathErr,
		},
		{
			name:          "should fail if file not exist",
			curDevfileCtx: devfileCtx.NewDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: &readDevfileErr,
		},
		{
			name:          "should fail if url not exist",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(httpPrefix + uri1),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: notExistURI,
				},
			},
			wantErr: &URLNotFoundErr,
		},
		{
			name:          "should fail if with invalid URI format",
			curDevfileCtx: devfileCtx.NewURLDevfileCtx(OutputDevfileYamlPath),
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Uri: invalidURL,
				},
			},
			wantErr: &invalidURLErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if the main devfile is from local, need to set absolute path
			if tt.curDevfileCtx.GetURL() == "" {
				err := tt.curDevfileCtx.SetAbsPath()
				if err != nil {
					t.Errorf("Test_parseFromURI() unexpected error: %v", err)
					return
				}
			}
			got, err := parseFromURI(tt.importReference, tt.curDevfileCtx, &resolutionContextTree{}, resolverTools{})
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromURI() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromURI() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromURI(): Error message should match")
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
					SchemaVersion: schemaVersion,
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

	latestParentDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime-latest",
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

	wantDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
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

	latestWantDevfile := DevfileObj{
		Data: &v2.DevfileV2{
			Devfile: v1.Devfile{
				DevfileHeader: devfilepkg.DevfileHeader{
					SchemaVersion: schemaVersion,
				},
				DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
						Components: []v1.Component{
							{
								Name: "runtime-latest",
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

	invalidURLErr := "the provided registryURL: .* is not a valid URL"
	URLNotFoundErr := "failed to retrieve .*, 404: Not Found"
	missingRegistryURLErr := "failed to fetch from registry, registry URL is not provided"
	invalidRegistryURLErr := "Get .* dial tcp: lookup http: .*"

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var err error
		if strings.Contains(r.URL.Path, "/devfiles/"+registryId) {
			if strings.Contains(r.URL.Path, "latest") {
				data, err = yaml.Marshal(latestParentDevfile.Data)
			} else if strings.Contains(r.URL.Path, "1.1.0") {
				data, err = yaml.Marshal(parentDevfile.Data)
			} else if r.URL.Path == fmt.Sprintf("/devfiles/%s/", registryId) {
				data, err = yaml.Marshal(parentDevfile.Data)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("Test_parseFromRegistry() unexpected error while doing yaml marshal: %v", err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Test_parseFromRegistry() unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", registry)
	if err != nil {
		t.Errorf("Test_parseFromRegistry() unexpected error while creating listener: %v", err)
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
		wantErr         *string
	}{
		{
			name:        "should fail if provided registryUrl does not have protocol prefix",
			wantDevFile: wantDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: registry,
			},
			wantErr: &invalidURLErr,
		},
		{
			name:        "should be able to parse from provided registryUrl with prefix",
			wantDevFile: wantDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				RegistryUrl: httpPrefix + registry,
			},
		},
		{
			name:        "should be able to parse from registry URL defined in tool",
			wantDevFile: wantDevfile,
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
			name:        "should be able to parse from provided registryUrl with latest version specified",
			wantDevFile: latestWantDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				Version:     "latest",
				RegistryUrl: httpPrefix + registry,
			},
		},
		{
			name:        "should be able to parse from provided registryUrl with version specified",
			wantDevFile: wantDevfile,
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				Version:     "1.1.0",
				RegistryUrl: httpPrefix + registry,
			},
		},
		{
			name: "should fail if version does not exist",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
				Version:     "999.9.9",
				RegistryUrl: httpPrefix + registry,
			},
			wantErr: &URLNotFoundErr,
		},
		{
			name: "should fail if registryId does not exist",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + registry,
			},
			wantErr: &URLNotFoundErr,
		},
		{
			name: "should fail if registryUrl is not provided, and no registry URLs has been set in tool",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: registryId,
				},
			},
			wantErr: &missingRegistryURLErr,
		},
		{
			name: "should fail if registryUrl is invalid",
			importReference: v1.ImportReference{
				ImportReferenceUnion: v1.ImportReferenceUnion{
					Id: notExistId,
				},
				RegistryUrl: httpPrefix + invalidRegistry,
			},
			wantErr: &invalidRegistryURLErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFromRegistry(tt.importReference, &resolutionContextTree{}, tt.tool)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromRegistry() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromRegistry() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromRegistry(): Error message should match")
			}
		})
	}
}

func Test_getDevfileFromDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("Failed to create temp dir: %s, error: %v", tempDir, err)
	}
	defer os.RemoveAll(tempDir)

	devfile := filepath.Join(tempDir, "devfile.yaml")
	if err := ioutil.WriteFile(devfile, []byte(""), 0666); err != nil {
		t.Errorf("Failed to create temp devfile, error: %v", err)
	}

	missingDevfileDir := "no/devfile/here"

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{
			name:    "should be able to get devfile from dir",
			dir:     tempDir,
			wantErr: false,
		},
		{
			name:    "should fail if directory doesn't have devfile",
			dir:     missingDevfileDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := false
			_, err = getDevfileFromDir(tt.dir)
			if err != nil {
				gotErr = true
			}

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("Got error: %t, want error: %t", gotErr, tt.wantErr)
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

	crdNotFoundErr := "not found"

	tests := []struct {
		name                  string
		curDevfileCtx         devfileCtx.DevfileCtx
		importReference       v1.ImportReference
		devWorkspaceResources map[string]v1.DevWorkspaceTemplate
		errors                map[string]string
		wantDevFile           DevfileObj
		wantErr               *string
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
				name: crdNotFoundErr,
			},
			wantErr: &crdNotFoundErr,
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
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Test_parseFromKubeCRD() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && !reflect.DeepEqual(got.Data, tt.wantDevFile.Data) {
				t.Errorf("Test_parseFromKubeCRD() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, got, pretty.Compare(tt.wantDevFile, got))
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "Test_parseFromKubeCRD(): Error message should match")
			}
		})
	}
}

func Test_setDefaults(t *testing.T) {
	type testType struct {
		name        string
		dataObj     data.DevfileData
		wantDevFile data.DevfileData
	}

	var tests []testType
	var version string

	// set up tests for unset boolean properties
	for i := range apiSchemaVersions {
		version = apiSchemaVersions[i]
		testName := fmt.Sprintf("Verify defaults on unset boolean properties for devfile %s", version)
		want, err := getBooleanDevfileTestData(version, true)
		if err != nil {
			t.Errorf("GetBooleanDevfileTestData() unexpected error %v ", err)
		}
		obj, err := getUnsetBooleanDevfileTestData(version)
		if err != nil {
			t.Errorf("GetUnsetBooleanDevfileTestData() unexpected error %v ", err)
		}
		tests = append(tests, testType{
			name:        testName,
			dataObj:     obj,
			wantDevFile: want,
		})
	}

	//repeat tests on set boolean properties
	for i := range apiSchemaVersions {
		version = apiSchemaVersions[i]
		testName := fmt.Sprintf("Verify defaults on set boolean properties for devfile %s", version)
		obj, err := getBooleanDevfileTestData(version, false)
		if err != nil {
			t.Errorf("GetBooleanDevfileTestData() unexpected error %v ", err)
		}

		tests = append(tests, testType{
			name:        testName,
			dataObj:     obj,
			wantDevFile: obj, //setDefaults should not alter properties that are explicitly set, so "want" structure should be identical
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DevfileObj{Data: tt.dataObj}
			err := setDefaults(d)
			if err != nil {
				t.Errorf("Test_setDefaults() unexpected error setting defaults %v ", err)
			} else if err == nil && !reflect.DeepEqual(d.Data, tt.wantDevFile) {
				t.Errorf("Test_setDefaults() error: wanted: %v, got: %v, difference at %v/ ", tt.wantDevFile, d.Data, pretty.Compare(tt.wantDevFile, tt.dataObj))
			}

		})
	}
}

// getUnsetBooleanDevfileObj returns a DevfileData object that contains unset boolean properties
func getUnsetBooleanDevfileTestData(apiVersion string) (devfileData data.DevfileData, err error) {
	devfileData = &v2.DevfileV2{
		Devfile: v1.Devfile{
			DevfileHeader: devfilepkg.DevfileHeader{
				SchemaVersion: apiVersion,
			},
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
							Id: "testrun",
							CommandUnion: v1.CommandUnion{
								Apply: &v1.ApplyCommand{
									LabeledCommand: v1.LabeledCommand{
										BaseCommand: v1.BaseCommand{
											Group: &v1.CommandGroup{
												Kind: v1.BuildCommandGroupKind,
											},
										},
									},
								},
							},
						},
						{
							Id: "allcmds",
							CommandUnion: v1.CommandUnion{
								Composite: &v1.CompositeCommand{
									Commands: []string{"testrun", "devrun"},
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
										Annotation: &v1.Annotation{
											Deployment: map[string]string{
												"deploy-key1": "deploy-value1",
											},
											Service: map[string]string{
												"svc-key1": "svc-value1",
												"svc-key2": "svc-value3",
											},
										},
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []v1.Endpoint{
										{
											Name:       "log",
											TargetPort: 443,
											Annotations: map[string]string{
												"ingress-key1": "ingress-value1",
												"ingress-key2": "ingress-value3",
											},
										},
									},
								},
							},
						},
						testingutil.GetFakeVolumeComponent("volume", "2Gi"),
						{
							Name: "openshift",
							ComponentUnion: v1.ComponentUnion{
								Openshift: &v1.OpenshiftComponent{
									K8sLikeComponent: v1.K8sLikeComponent{
										K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
											Uri: "https://xyz.com/dir/file.yaml",
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "metrics",
												TargetPort: 8080,
											},
										},
									},
								},
							},
						},
					},
					Events: &v1.Events{
						DevWorkspaceEvents: v1.DevWorkspaceEvents{
							PostStart: []string{"post-start-0"},
							PostStop:  []string{"post-stop"},
							PreStop:   []string{},
							PreStart:  []string{},
						},
					},
				},
			},
		},
	}

	if apiVersion != string(data.APISchemaVersion200) && apiVersion != string(data.APISchemaVersion210) {
		comp := []v1.Component{testingutil.GetDockerImageTestComponent(testingutil.DockerImageValues{}, nil, nil)}
		err = devfileData.AddComponents(comp)
	}

	return devfileData, err

}

//getBooleanDevfileTestData returns a DevfileData object that contains set values for the boolean properties.  If setDefault is true, an object with the default boolean values will be returned
func getBooleanDevfileTestData(apiVersion string, setDefault bool) (devfileData data.DevfileData, err error) {

	type boolValues struct {
		hotReloadCapable *bool
		secure           *bool
		parallel         *bool
		dedicatedPod     *bool
		mountSources     *bool
		isDefault        *bool
		rootRequired     *bool
		ephemeral        *bool
		autoBuild        *bool
		deployByDefaul   *bool
	}

	//default values according to spec
	defaultBools := boolValues{&isFalse, &isFalse, &isFalse, &isFalse, &isTrue, &isFalse, &isFalse, &isFalse, &isFalse, &isFalse}
	//set values will be a mix of default and inverse values
	setBools := boolValues{&isTrue, &isTrue, &isFalse, &isTrue, &isFalse, &isFalse, &isTrue, &isFalse, &isTrue, &isFalse}

	var values boolValues

	if setDefault {
		values = defaultBools
	} else {
		values = setBools
	}

	devfileData = &v2.DevfileV2{
		Devfile: v1.Devfile{
			DevfileHeader: devfilepkg.DevfileHeader{
				SchemaVersion: apiVersion,
			},
			DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
					Commands: []v1.Command{
						{
							Id: "devrun",
							CommandUnion: v1.CommandUnion{
								Exec: &v1.ExecCommand{
									CommandLine:      "npm run",
									WorkingDir:       "/projects/nodejs-starter",
									HotReloadCapable: values.hotReloadCapable,
								},
							},
						},
						{
							Id: "testrun",
							CommandUnion: v1.CommandUnion{
								Apply: &v1.ApplyCommand{
									LabeledCommand: v1.LabeledCommand{
										BaseCommand: v1.BaseCommand{
											Group: &v1.CommandGroup{
												Kind:      v1.BuildCommandGroupKind,
												IsDefault: values.isDefault,
											},
										},
									},
								},
							},
						},
						{
							Id: "allcmds",
							CommandUnion: v1.CommandUnion{
								Composite: &v1.CompositeCommand{
									Commands: []string{"testrun", "devrun"},
									Parallel: values.parallel,
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
										Annotation: &v1.Annotation{
											Deployment: map[string]string{
												"deploy-key1": "deploy-value1",
											},
											Service: map[string]string{
												"svc-key1": "svc-value1",
												"svc-key2": "svc-value3",
											},
										},
										Image:        "quay.io/nodejs-12",
										DedicatedPod: values.dedicatedPod,
										MountSources: values.mountSources,
									},
									Endpoints: []v1.Endpoint{
										{
											Name:       "log",
											TargetPort: 443,
											Annotations: map[string]string{
												"ingress-key1": "ingress-value1",
												"ingress-key2": "ingress-value3",
											},
											Secure: values.secure,
										},
									},
								},
							},
						},
						testingutil.GetFakeVolumeComponent("volume", "2Gi"),
						{
							Name: "openshift",
							ComponentUnion: v1.ComponentUnion{
								Openshift: &v1.OpenshiftComponent{
									K8sLikeComponent: v1.K8sLikeComponent{
										K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
											Uri: "https://xyz.com/dir/file.yaml",
										},
										Endpoints: []v1.Endpoint{
											{
												Name:       "metrics",
												TargetPort: 8080,
												Secure:     values.secure,
											},
										},
									},
								},
							},
						},
					},
					Events: &v1.Events{
						DevWorkspaceEvents: v1.DevWorkspaceEvents{
							PostStart: []string{"post-start-0"},
							PostStop:  []string{"post-stop"},
							PreStop:   []string{},
							PreStart:  []string{},
						},
					},
				},
			},
		},
	}

	if apiVersion != string(data.APISchemaVersion200) {
		volComponent, _ := devfileData.GetComponents(common.DevfileOptions{ComponentOptions: common.ComponentOptions{
			ComponentType: v1.VolumeComponentType,
		}})

		volComponent[0].Volume.Ephemeral = values.ephemeral
	}

	if apiVersion != string(data.APISchemaVersion200) && apiVersion != string(data.APISchemaVersion210) {
		comp := []v1.Component{testingutil.GetDockerImageTestComponent(testingutil.DockerImageValues{RootRequired: values.rootRequired}, values.autoBuild, nil)}
		err = devfileData.AddComponents(comp)

		openshiftComponent, _ := devfileData.GetComponents(common.DevfileOptions{ComponentOptions: common.ComponentOptions{
			ComponentType: v1.OpenshiftComponentType,
		}})
		openshiftComponent[0].Openshift.DeployByDefault = values.deployByDefaul

	}

	return devfileData, err
}
