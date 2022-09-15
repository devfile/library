//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/kylelemons/godebug/pretty"
)

func TestAddEnvVars(t *testing.T) {

	tests := []struct {
		name           string
		listToAdd      map[string][]v1alpha2.EnvVar
		currentDevfile *DevfileV2
		wantDevFile    *DevfileV2
	}{
		{
			name: "add env vars",
			listToAdd: map[string][]v1alpha2.EnvVar{
				"loadbalancer": {
					{
						Name:  "DATABASE_PASSWORD",
						Value: "苦痛",
					},
				},
			},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env: []v1alpha2.EnvVar{
													{
														Name:  "DATABASE_PASSWORD",
														Value: "苦痛",
													},
												},
											},
											Endpoints: []v1alpha2.Endpoint{
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
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
												Env: []v1alpha2.EnvVar{
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
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.currentDevfile.AddEnvVars(tt.listToAdd)

			if err != nil {
				t.Errorf("TestAddAndRemoveEnvVars() unexpected error while adding env vars %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile, tt.wantDevFile) {
				t.Errorf("TestAddAndRemoveEnvVars() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile, tt.wantDevFile))
			}

		})
	}

}

func TestRemoveEnvVars(t *testing.T) {

	tests := []struct {
		name           string
		listToRemove   map[string][]string
		currentDevfile *DevfileV2
		wantDevFile    *DevfileV2
		wantRemoveErr  bool
	}{
		{
			name: "remove env vars",
			listToRemove: map[string][]string{
				"runtime": {
					"DATABASE_PASSWORD",
				},
			},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env:   []v1alpha2.EnvVar{},
											},
											Endpoints: []v1alpha2.Endpoint{
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
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
												Env:   []v1alpha2.EnvVar{},
											},
										},
									},
								},
							},
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
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
		{
			name: "remove non-existent env vars",
			listToRemove: map[string][]string{
				"runtime": {
					"NON_EXISTENT_KEY",
				},
			},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env: []v1alpha2.EnvVar{
													{
														Name:  "DATABASE_PASSWORD",
														Value: "苦痛",
													},
												},
											},
											Endpoints: []v1alpha2.Endpoint{
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
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
											},
										},
									},
								},
							},
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
								{
									SubDir: "/projects",
									Name:   "starter-project-2",
								},
							},
						},
					},
				},
			},
			wantRemoveErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currentDevfile.RemoveEnvVars(tt.listToRemove)

			if (err != nil) != tt.wantRemoveErr {
				t.Errorf("TestAddAndRemoveEnvVars() unexpected error while removing env vars %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile, tt.wantDevFile) {
				t.Errorf("TestAddAndRemoveEnvVars() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile, tt.wantDevFile))
			}

		})
	}

}

func TestSetPorts(t *testing.T) {

	tests := []struct {
		name           string
		portToSet      map[string][]string
		currentDevfile *DevfileV2
		wantDevFile    *DevfileV2
	}{
		{
			name:           "set ports",
			portToSet:      map[string][]string{"runtime": {"9000"}, "loadbalancer": {"8000"}},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env: []v1alpha2.EnvVar{
													{
														Name:  "DATABASE_PASSWORD",
														Value: "苦痛",
													},
												},
											},
											Endpoints: []v1alpha2.Endpoint{
												{
													Name:       "port-3030",
													TargetPort: 3030,
												},
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
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
											},
											Endpoints: []v1alpha2.Endpoint{
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
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.currentDevfile.SetPorts(tt.portToSet)

			if err != nil {
				t.Errorf("TestSetAndRemovePorts() unexpected error while adding ports %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile, tt.wantDevFile) {
				t.Errorf("TestSetAndRemovePorts() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile, tt.wantDevFile))
			}
		})
	}

}

func TestRemovePorts(t *testing.T) {

	tests := []struct {
		name           string
		portToRemove   map[string][]string
		currentDevfile *DevfileV2
		wantDevFile    *DevfileV2
		wantRemoveErr  bool
	}{
		{
			name:           "remove ports",
			portToRemove:   map[string][]string{"runtime": {"3030"}},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env: []v1alpha2.EnvVar{
													{
														Name:  "DATABASE_PASSWORD",
														Value: "苦痛",
													},
												},
											},
											Endpoints: []v1alpha2.Endpoint{},
										},
									},
								},
								{
									Name: "loadbalancer",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
											},
											Endpoints: []v1alpha2.Endpoint{},
										},
									},
								},
							},
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
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
		{
			name:           "remove non-existent ports",
			portToRemove:   map[string][]string{"runtime": {"3050"}},
			currentDevfile: testDevfileData(),
			wantDevFile: &DevfileV2{
				Devfile: v1alpha2.Devfile{
					DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
						DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
							Commands: []v1alpha2.Command{
								{
									Id: "devbuild",
									CommandUnion: v1alpha2.CommandUnion{
										Exec: &v1alpha2.ExecCommand{
											WorkingDir: "/projects/nodejs-starter",
										},
									},
								},
							},
							Components: []v1alpha2.Component{
								{
									Name: "runtime",
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nodejs-12",
												Env: []v1alpha2.EnvVar{
													{
														Name:  "DATABASE_PASSWORD",
														Value: "苦痛",
													},
												},
											},
											Endpoints: []v1alpha2.Endpoint{
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
									ComponentUnion: v1alpha2.ComponentUnion{
										Container: &v1alpha2.ContainerComponent{
											Container: v1alpha2.Container{
												Image: "quay.io/nginx",
											},
										},
									},
								},
							},
							Events: &v1alpha2.Events{
								DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
									PostStop: []string{"post-stop"},
								},
							},
							Projects: []v1alpha2.Project{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter-build",
								},
							},
							StarterProjects: []v1alpha2.StarterProject{
								{
									SubDir: "/projects",
									Name:   "starter-project-2",
								},
							},
						},
					},
				},
			},
			wantRemoveErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currentDevfile.RemovePorts(tt.portToRemove)

			if (err != nil) != tt.wantRemoveErr {
				t.Errorf("TestSetAndRemovePorts() unexpected error while removing ports %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile, tt.wantDevFile) {
				t.Errorf("TestSetAndRemovePorts() error: wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile, tt.wantDevFile))
			}
		})
	}

}

func testDevfileData() *DevfileV2 {
	return &DevfileV2{
		Devfile: v1alpha2.Devfile{
			DevWorkspaceTemplateSpec: v1alpha2.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
					Commands: []v1alpha2.Command{
						{
							Id: "devbuild",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					},
					Components: []v1alpha2.Component{
						{
							Name: "runtime",
							ComponentUnion: v1alpha2.ComponentUnion{
								Container: &v1alpha2.ContainerComponent{
									Container: v1alpha2.Container{
										Image: "quay.io/nodejs-12",
										Env: []v1alpha2.EnvVar{
											{
												Name:  "DATABASE_PASSWORD",
												Value: "苦痛",
											},
										},
									},
									Endpoints: []v1alpha2.Endpoint{
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
							ComponentUnion: v1alpha2.ComponentUnion{
								Container: &v1alpha2.ContainerComponent{
									Container: v1alpha2.Container{
										Image: "quay.io/nginx",
									},
								},
							},
						},
					},
					Events: &v1alpha2.Events{
						DevWorkspaceEvents: v1alpha2.DevWorkspaceEvents{
							PostStop: []string{"post-stop"},
						},
					},
					Projects: []v1alpha2.Project{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
					},
					StarterProjects: []v1alpha2.StarterProject{
						{
							SubDir: "/projects",
							Name:   "starter-project-2",
						},
					},
				},
			},
		},
	}
}
