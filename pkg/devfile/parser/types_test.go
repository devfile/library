package parser

import (
	"reflect"
	"testing"

	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
	"github.com/devfile/parser/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"

	"github.com/devfile/api/pkg/apis/workspaces/v1alpha1"
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
		overridePatch []v1alpha1.Command
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
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: commandLineBuild,
								Component:   componentName0,
								Env:         nil,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Group: &v1alpha1.CommandGroup{
											IsDefault: false,
											Kind:      v1alpha1.BuildCommandGroupKind,
										},
										Id: "devbuild",
									},
								},
								WorkingDir: workingDir,
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Command{
					{
						Exec: &v1alpha1.ExecCommand{
							CommandLine: overrideBuild,
							Component:   overrideComponent0,
							LabeledCommand: v1alpha1.LabeledCommand{
								BaseCommand: v1alpha1.BaseCommand{
									Group: &v1alpha1.CommandGroup{
										IsDefault: true,
										Kind:      v1alpha1.BuildCommandGroupKind,
									},
									Id: "devbuild",
								},
							},
							WorkingDir: overrideWorkingDir,
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: overrideBuild,
								Component:   overrideComponent0,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Group: &v1alpha1.CommandGroup{
											IsDefault: true,
											Kind:      v1alpha1.BuildCommandGroupKind,
										},
										Id: "devbuild",
									},
								},
								WorkingDir: overrideWorkingDir,
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
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devbuild",
										Attributes: map[string]string{
											"key-0": "value-0",
										},
									},
								},
								Env: []v1alpha1.EnvVar{
									testingutil.GetFakeEnv("env-0", "value-0"),
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Command{
					{
						Exec: &v1alpha1.ExecCommand{
							LabeledCommand: v1alpha1.LabeledCommand{
								BaseCommand: v1alpha1.BaseCommand{
									Id: "devbuild",
									Attributes: map[string]string{
										"key-1": "value-1",
									},
								},
							},
							Env: []v1alpha1.EnvVar{
								testingutil.GetFakeEnv("env-0", "value-0-0"),
								testingutil.GetFakeEnv("env-1", "value-1"),
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devbuild",
										Attributes: map[string]string{
											"key-0": "value-0",
											"key-1": "value-1",
										},
									},
								},
								Env: []v1alpha1.EnvVar{
									testingutil.GetFakeEnv("env-0", "value-0-0"),
									testingutil.GetFakeEnv("env-1", "value-1"),
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
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: commandLineBuild,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devbuild",
									},
								},
							},
						},
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: commandLineRun,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devrun",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Command{
					{
						Exec: &v1alpha1.ExecCommand{
							CommandLine: overrideBuild,
							LabeledCommand: v1alpha1.LabeledCommand{
								BaseCommand: v1alpha1.BaseCommand{
									Id: "devbuild",
								},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: overrideBuild,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devbuild",
									},
								},
							},
						},
						{
							Exec: &v1alpha1.ExecCommand{
								CommandLine: commandLineRun,
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devrun",
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
					Commands: []v1alpha1.Command{
						{
							Exec: &v1alpha1.ExecCommand{
								Env: []v1alpha1.EnvVar{
									testingutil.GetFakeEnv("env-0", "value-0"),
								},
								LabeledCommand: v1alpha1.LabeledCommand{
									BaseCommand: v1alpha1.BaseCommand{
										Id: "devbuild",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Command{
					{
						Exec: &v1alpha1.ExecCommand{
							Env: []v1alpha1.EnvVar{
								testingutil.GetFakeEnv("env-0", "value-0-0"),
								testingutil.GetFakeEnv("env-1", "value-1"),
							},
							LabeledCommand: v1alpha1.LabeledCommand{
								BaseCommand: v1alpha1.BaseCommand{
									Id: "devbuild-custom",
								},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []v1alpha1.Command{},
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

func TestDevfileObj_OverrideComponents(t *testing.T) {

	containerImage0 := "image-0"
	containerImage1 := "image-1"

	overrideContainerImage := "image-0-override"

	type args struct {
		overridePatch []v1alpha1.Component
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override a container's non list/map fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Args:          []string{"arg-0", "arg-1"},
									Command:       []string{"cmd-0", "cmd-1"},
									Image:         containerImage0,
									MemoryLimit:   "512Mi",
									MountSources:  false,
									Name:          "nodejs",
									SourceMapping: "/source",
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Component{
					{
						Container: &v1alpha1.ContainerComponent{
							Container: v1alpha1.Container{
								Args:          []string{"arg-0-0", "arg-1-1"},
								Command:       []string{"cmd-0-0", "cmd-1-1"},
								Image:         overrideContainerImage,
								MemoryLimit:   "1Gi",
								MountSources:  true,
								Name:          "nodejs",
								SourceMapping: "/data",
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Args:          []string{"arg-0-0", "arg-1-1"},
									Command:       []string{"cmd-0-0", "cmd-1-1"},
									Image:         overrideContainerImage,
									MemoryLimit:   "1Gi",
									MountSources:  true,
									Name:          "nodejs",
									SourceMapping: "/data",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: append/override a command's list fields based on the key",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Endpoints: []v1alpha1.Endpoint{
									{
										Attributes: map[string]string{
											"key-0": "value-0",
											"key-1": "value-1",
										},
										Name:       "endpoint-0",
										TargetPort: 8080,
									},
								},
								Container: v1alpha1.Container{
									Env: []v1alpha1.EnvVar{
										testingutil.GetFakeEnv("env-0", "value-0"),
									},
									Name: "nodejs",
									VolumeMounts: []v1alpha1.VolumeMount{
										testingutil.GetFakeVolumeMount("volume-0", "path-0"),
									},
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Component{
					{
						Container: &v1alpha1.ContainerComponent{
							Endpoints: []v1alpha1.Endpoint{
								{
									Attributes: map[string]string{
										"key-1":      "value-1-1",
										"key-append": "value-append",
									},
									Name:       "endpoint-0",
									TargetPort: 9090,
								},
								{
									Attributes: map[string]string{
										"key-0": "value-0",
									},
									Name:       "endpoint-1",
									TargetPort: 3000,
								},
							},
							Container: v1alpha1.Container{
								Env: []v1alpha1.EnvVar{
									testingutil.GetFakeEnv("env-0", "value-0-0"),
									testingutil.GetFakeEnv("env-1", "value-1"),
								},
								Name: "nodejs",
								VolumeMounts: []v1alpha1.VolumeMount{
									testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
									testingutil.GetFakeVolumeMount("volume-1", "path-1"),
								},
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Env: []v1alpha1.EnvVar{
										testingutil.GetFakeEnv("env-0", "value-0-0"),
										testingutil.GetFakeEnv("env-1", "value-1"),
									},
									Name: "nodejs",
									VolumeMounts: []v1alpha1.VolumeMount{
										testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
										testingutil.GetFakeVolumeMount("volume-1", "path-1"),
									},
								},
								Endpoints: []v1alpha1.Endpoint{
									{
										Attributes: map[string]string{
											"key-0":      "value-0",
											"key-1":      "value-1-1",
											"key-append": "value-append",
										},
										Name:       "endpoint-0",
										TargetPort: 9090,
									},
									{
										Attributes: map[string]string{
											"key-0": "value-0",
										},
										Name:       "endpoint-1",
										TargetPort: 3000,
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
			name: "case 3: if multiple, override the correct command",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Image: containerImage0,
									Name:  "nodejs",
								},
							},
						},
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Image: containerImage1,
									Name:  "runtime",
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Component{
					{
						Container: &v1alpha1.ContainerComponent{
							Container: v1alpha1.Container{
								Image: overrideContainerImage,
								Name:  "nodejs",
							},
						},
					},
					{
						Container: &v1alpha1.ContainerComponent{
							Container: v1alpha1.Container{
								Image: containerImage1,
								Name:  "runtime",
							},
						}},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Image: overrideContainerImage,
									Name:  "nodejs",
								},
							},
						},
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Image: containerImage1,
									Name:  "runtime",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 4: throw error if component to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []v1alpha1.Component{
						{
							Container: &v1alpha1.ContainerComponent{
								Container: v1alpha1.Container{
									Image: containerImage0,
									Name:  "nodejs",
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Component{
					{
						Container: &v1alpha1.ContainerComponent{
							Container: v1alpha1.Container{
								Image: containerImage0,
								Name:  "nodejs-custom",
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		if tt.name != "case 1: override a container's non list/map fields" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideComponents(tt.args.overridePatch)
			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideComponents() error = %v, wantErr %v", err, tt.wantErr)
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

func TestDevfileObj_OverrideEvents(t *testing.T) {
	type args struct {
		overridePatch v1alpha1.Events
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override the events",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: v1alpha1.Events{
						WorkspaceEvents: v1alpha1.WorkspaceEvents{
							PostStart: []string{"post-start-0", "post-start-1"},
							PostStop:  []string{"post-stop-0", "post-stop-1"},
							PreStart:  []string{"pre-start-0", "pre-start-1"},
							PreStop:   []string{"pre-stop-0", "pre-stop-1"},
						},
					},
				},
			},
			args: args{
				overridePatch: v1alpha1.Events{
					WorkspaceEvents: v1alpha1.WorkspaceEvents{
						PostStart: []string{"override-post-start-0", "override-post-start-1"},
						PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
						PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
						PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: v1alpha1.Events{
						WorkspaceEvents: v1alpha1.WorkspaceEvents{
							PostStart: []string{"override-post-start-0", "override-post-start-1"},
							PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
							PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
							PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
						},
					},
				},
			},
		},
		{
			name: "case 2: override some of the events",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: v1alpha1.Events{
						WorkspaceEvents: v1alpha1.WorkspaceEvents{
							PostStart: []string{"post-start-0", "post-start-1"},
							PostStop:  []string{"post-stop-0", "post-stop-1"},
						},
					},
				},
			},
			args: args{
				overridePatch: v1alpha1.Events{
					WorkspaceEvents: v1alpha1.WorkspaceEvents{
						PostStart: []string{"override-post-start-0", "override-post-start-1"},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: v1alpha1.Events{
						WorkspaceEvents: v1alpha1.WorkspaceEvents{
							PostStart: []string{"override-post-start-0", "override-post-start-1"},
							PostStop:  []string{"post-stop-0", "post-stop-1"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.devFileObj.OverrideEvents(tt.args.overridePatch); (err != nil) != tt.wantErr {
				t.Errorf("OverrideEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

func TestDevfileObj_OverrideProjects(t *testing.T) {
	projectName0 := "project-0"
	projectName1 := "project-1"

	type args struct {
		overridePatch []v1alpha1.Project
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		wantDevFileObj DevfileObj
		args           args
		wantErr        bool
	}{
		{
			name: "case 1: override a project's fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []v1alpha1.Project{
						{
							ClonePath: "/data",
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "master",
									},
								},
								Zip: nil,
							},
							Name: projectName0,
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Project{
					{
						ClonePath: "/source",
						ProjectSource: v1alpha1.ProjectSource{
							Github: &v1alpha1.GithubProjectSource{
								GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
									Branch: "release-1.0.0",
								},
							},
							Zip: nil,
						},
						Name: projectName0,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []v1alpha1.Project{
						{
							ClonePath: "/source",
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "release-1.0.0",
									},
								},
								Zip: nil,
							},
							Name: projectName0,
						},
					},
				},
			},
		},
		{
			name: "case 2: if multiple, override the correct project",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []v1alpha1.Project{
						{
							ClonePath: "/data",
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "master",
									},
								},
								Zip: nil,
							},
							Name: projectName0,
						},
						{
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "master",
									},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Project{
					{
						ClonePath: "/source",
						ProjectSource: v1alpha1.ProjectSource{
							Github: &v1alpha1.GithubProjectSource{
								GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
									Branch: "release-1.0.0",
								},
							},
							Zip: nil,
						},
						Name: projectName0,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []v1alpha1.Project{
						{
							ClonePath: "/source",
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "release-1.0.0",
									},
								},
								Zip: nil,
							},
							Name: projectName0,
						},
						{
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "master",
									},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
		},
		{
			name: "case 3: throw error if project to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []v1alpha1.Project{
						{
							ClonePath: "/data",
							ProjectSource: v1alpha1.ProjectSource{
								Github: &v1alpha1.GithubProjectSource{
									GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
										Branch: "master",
									},
								},
								Zip: nil,
							},
							Name: projectName0,
						},
					},
				},
			},
			args: args{
				overridePatch: []v1alpha1.Project{
					{
						ClonePath: "/source",
						ProjectSource: v1alpha1.ProjectSource{
							Github: &v1alpha1.GithubProjectSource{
								GitLikeProjectSource: v1alpha1.GitLikeProjectSource{
									Branch: "release-1.0.0",
								},
							},
							Zip: nil,
						},
						Name: "custom-project",
					},
				},
			},
			wantDevFileObj: DevfileObj{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideProjects(tt.args.overridePatch)

			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideProjects() error = %v, wantErr %v", err, tt.wantErr)
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
