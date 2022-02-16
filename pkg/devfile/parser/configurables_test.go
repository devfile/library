package parser

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
)

func TestAddAndRemoveEnvVars(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		listToAdd      map[string][]v1.EnvVar
		listToRemove   map[string][]string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
		wantRemoveErr  bool
	}{
		{
			name: "add and remove env vars",
			listToAdd: map[string][]v1.EnvVar{
				"runtime": {
					{
						Name:  "DATABASE_PASSWORD",
						Value: "苦痛",
					},
				},
				"loadbalancer": {
					{
						Name:  "DATABASE_PASSWORD",
						Value: "苦痛",
					},
				},
			},
			listToRemove: map[string][]string{
				"loadbalancer": {
					"DATABASE_PASSWORD",
				},
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
														TargetPort: 3030,
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
													Env:   []v1.EnvVar{},
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
								StarterProjects: []v1.StarterProject{
									{
										SubDir: "/projects",
										Name:   "starter-project-2",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "remove non-existent env vars",
			listToAdd: map[string][]v1.EnvVar{
				"runtime": {
					{
						Name:  "DATABASE_PASSWORD",
						Value: "苦痛",
					},
				},
			},
			listToRemove: map[string][]string{
				"runtime": {
					"NON_EXISTENT_KEY",
				},
			},
			currentDevfile: testDevfileObj(fs),
			wantRemoveErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.currentDevfile.AddEnvVars(tt.listToAdd)

			if err != nil {
				t.Errorf("TestAddAndRemoveEnvVars() unexpected error while adding env vars %+v", err.Error())
			}

			err = tt.currentDevfile.RemoveEnvVars(tt.listToRemove)

			if (err != nil) != tt.wantRemoveErr {
				t.Errorf("TestAddAndRemoveEnvVars() unexpected error while removing env vars %+v", err.Error())
			}

			if !tt.wantRemoveErr && !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("TestAddAndRemoveEnvVars() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
			}

		})
	}

}

func TestSetAndRemovePorts(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		portToSet      map[string][]string
		portToRemove   map[string][]string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
		wantRemoveErr  bool
	}{
		{
			name:           "add and remove ports",
			portToSet:      map[string][]string{"runtime": {"9000"}, "loadbalancer": {"8000"}},
			portToRemove:   map[string][]string{"runtime": {"3030"}},
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
														Name:       "port-9000-tcp",
														TargetPort: 9000,
														Protocol:   "tcp",
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
														Name:       "port-8000-tcp",
														TargetPort: 8000,
														Protocol:   "tcp",
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
								StarterProjects: []v1.StarterProject{
									{
										SubDir: "/projects",
										Name:   "starter-project-2",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:           "remove non-existent ports",
			portToSet:      map[string][]string{"runtime": {"9000"}},
			portToRemove:   map[string][]string{"runtime": {"3050"}},
			currentDevfile: testDevfileObj(fs),
			wantRemoveErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.currentDevfile.SetPorts(tt.portToSet)

			if err != nil {
				t.Errorf("TestSetAndRemovePorts() unexpected error while adding ports %+v", err.Error())
			}

			err = tt.currentDevfile.RemovePorts(tt.portToRemove)

			if (err != nil) != tt.wantRemoveErr {
				t.Errorf("TestSetAndRemovePorts() unexpected error while removing ports %+v", err.Error())
			}

			if !tt.wantRemoveErr && !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("TestSetAndRemovePorts() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
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
												TargetPort: 3030,
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
						StarterProjects: []v1.StarterProject{
							{
								SubDir: "/projects",
								Name:   "starter-project-2",
							},
						},
					},
				},
			},
		},
	}
}
