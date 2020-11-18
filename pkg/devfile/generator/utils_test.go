package generator

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/testingutil"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestConvertEnvs(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name    string
		envVars []v1.EnvVar
		want    []corev1.EnvVar
	}{
		{
			name: "Case 1: One env var",
			envVars: []v1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []v1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
				},
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
				},
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
				},
			},
		},
		{
			name:    "Case 3: No env vars",
			envVars: []v1.EnvVar{},
			want:    []corev1.EnvVar{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := convertEnvs(tt.envVars)
			if !reflect.DeepEqual(tt.want, envVars) {
				t.Errorf("expected %v, wanted %v", envVars, tt.want)
			}
		})
	}
}

func TestConvertPorts(t *testing.T) {
	endpointsNames := []string{"endpoint1", "endpoint2"}
	endpointsPorts := []int{8080, 9090}
	tests := []struct {
		name      string
		endpoints []v1.Endpoint
		want      []corev1.ContainerPort
	}{
		{
			name: "Case 1: One Endpoint",
			endpoints: []v1.Endpoint{
				{
					Name:       endpointsNames[0],
					TargetPort: endpointsPorts[0],
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          endpointsNames[0],
					ContainerPort: int32(endpointsPorts[0]),
				},
			},
		},
		{
			name: "Case 2: Multiple env vars",
			endpoints: []v1.Endpoint{
				{
					Name:       endpointsNames[0],
					TargetPort: endpointsPorts[0],
				},
				{
					Name:       endpointsNames[1],
					TargetPort: endpointsPorts[1],
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          endpointsNames[0],
					ContainerPort: int32(endpointsPorts[0]),
				},
				{
					Name:          endpointsNames[1],
					ContainerPort: int32(endpointsPorts[1]),
				},
			},
		},
		{
			name:      "Case 3: No endpoints",
			endpoints: []v1.Endpoint{},
			want:      []corev1.ContainerPort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := convertPorts(tt.endpoints)
			if !reflect.DeepEqual(tt.want, ports) {
				t.Errorf("expected %v, wanted %v", ports, tt.want)
			}
		})
	}
}

func TestGetResourceReqs(t *testing.T) {
	limit := "1024Mi"
	quantity, err := resource.ParseQuantity(limit)
	if err != nil {
		t.Errorf("expected %v", err)
	}
	tests := []struct {
		name      string
		component v1.Component
		want      corev1.ResourceRequirements
	}{
		{
			name: "Case 1: One Endpoint",
			component: v1.Component{
				Name: "testcomponent",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{
						Container: v1.Container{
							MemoryLimit: "1024Mi",
						},
					},
				},
			},
			want: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: quantity,
				},
			},
		},
		{
			name:      "Case 2: Empty Component",
			component: v1.Component{},
			want:      corev1.ResourceRequirements{},
		},
		{
			name: "Case 3: Valid container, but empty memoryLimit",
			component: v1.Component{
				Name: "testcomponent",
				ComponentUnion: v1.ComponentUnion{
					Container: &v1.ContainerComponent{
						Container: v1.Container{
							Image: "testimage",
						},
					},
				},
			},
			want: corev1.ResourceRequirements{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := getResourceReqs(tt.component)
			if !reflect.DeepEqual(tt.want, req) {
				t.Errorf("expected %v, wanted %v", req, tt.want)
			}
		})
	}
}

func TestAddSyncRootFolder(t *testing.T) {

	tests := []struct {
		name               string
		sourceMapping      string
		wantSyncRootFolder string
	}{
		{
			name:               "Case 1: Valid Source Mapping",
			sourceMapping:      "/mypath",
			wantSyncRootFolder: "/mypath",
		},
		{
			name:               "Case 2: No Source Mapping",
			sourceMapping:      "",
			wantSyncRootFolder: DevfileSourceVolumeMount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := testingutil.CreateFakeContainer("container1")

			syncRootFolder := addSyncRootFolder(&container, tt.sourceMapping)

			if syncRootFolder != tt.wantSyncRootFolder {
				t.Errorf("TestAddSyncRootFolder sync root folder error - expected %v got %v", tt.wantSyncRootFolder, syncRootFolder)
			}

			for _, env := range container.Env {
				if env.Name == EnvProjectsRoot && env.Value != tt.wantSyncRootFolder {
					t.Errorf("PROJECT_ROOT error expected %s, actual %s", tt.wantSyncRootFolder, env.Value)
				}
			}
		})
	}
}

func TestAddSyncFolder(t *testing.T) {
	projectNames := []string{"some-name", "another-name"}
	projectRepos := []string{"https://github.com/some/repo.git", "https://github.com/another/repo.git"}
	projectClonePath := "src/github.com/golang/example/"
	invalidClonePaths := []string{"/var", "../var", "pkg/../../var"}
	sourceVolumePath := "/projects/app"

	tests := []struct {
		name     string
		projects []v1.Project
		want     string
		wantErr  bool
	}{
		{
			name:     "Case 1: No projects",
			projects: []v1.Project{},
			want:     sourceVolumePath,
			wantErr:  false,
		},
		{
			name: "Case 2: One project",
			projects: []v1.Project{
				{
					Name: projectNames[0],
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{
							GitLikeProjectSource: v1.GitLikeProjectSource{
								Remotes: map[string]string{"origin": projectRepos[0]},
							},
						},
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
			wantErr: false,
		},
		{
			name: "Case 3: Multiple projects",
			projects: []v1.Project{
				{
					Name: projectNames[0],
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{
							GitLikeProjectSource: v1.GitLikeProjectSource{
								Remotes: map[string]string{"origin": projectRepos[0]},
							},
						},
					},
				},
				{
					Name: projectNames[1],
					ProjectSource: v1.ProjectSource{
						Github: &v1.GithubProjectSource{
							GitLikeProjectSource: v1.GitLikeProjectSource{
								Remotes: map[string]string{"origin": projectRepos[0]},
							},
						},
					},
				},
				{
					Name: projectNames[1],
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{
							Location: projectRepos[1],
						},
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
			wantErr: false,
		},
		{
			name: "Case 4: Clone path set",
			projects: []v1.Project{
				{
					ClonePath: projectClonePath,
					Name:      projectNames[0],
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{
							Location: projectRepos[0],
						},
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectClonePath)),
			wantErr: false,
		},
		{
			name: "Case 5: Invalid clone path, set with absolute path",
			projects: []v1.Project{
				{
					ClonePath: invalidClonePaths[0],
					Name:      projectNames[0],
					ProjectSource: v1.ProjectSource{
						Github: &v1.GithubProjectSource{
							GitLikeProjectSource: v1.GitLikeProjectSource{
								Remotes: map[string]string{"origin": projectRepos[0]},
							},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 6: Invalid clone path, starts with ..",
			projects: []v1.Project{
				{
					ClonePath: invalidClonePaths[1],
					Name:      projectNames[0],
					ProjectSource: v1.ProjectSource{
						Git: &v1.GitProjectSource{
							GitLikeProjectSource: v1.GitLikeProjectSource{
								Remotes: map[string]string{"origin": projectRepos[0]},
							},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 7: Invalid clone path, contains ..",
			projects: []v1.Project{
				{
					ClonePath: invalidClonePaths[2],
					Name:      projectNames[0],
					ProjectSource: v1.ProjectSource{
						Zip: &v1.ZipProjectSource{
							Location: projectRepos[0],
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := testingutil.CreateFakeContainer("container1")

			err := addSyncFolder(&container, sourceVolumePath, tt.projects)

			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, actual %v", tt.wantErr, err)
			}

			for _, env := range container.Env {
				if env.Name == EnvProjectsSrc && env.Value != tt.want {
					t.Errorf("expected %s, actual %s", tt.want, env.Value)
				}
			}
		})
	}
}

func TestGetContainer(t *testing.T) {

	tests := []struct {
		name          string
		containerName string
		image         string
		isPrivileged  bool
		command       []string
		args          []string
		envVars       []corev1.EnvVar
		resourceReqs  corev1.ResourceRequirements
		ports         []corev1.ContainerPort
	}{
		{
			name:          "Case 1: Empty container params",
			containerName: "",
			image:         "",
			isPrivileged:  false,
			command:       []string{},
			args:          []string{},
			envVars:       []corev1.EnvVar{},
			resourceReqs:  corev1.ResourceRequirements{},
			ports:         []corev1.ContainerPort{},
		},
		{
			name:          "Case 2: Valid container params",
			containerName: "container1",
			image:         "quay.io/eclipse/che-java8-maven:nightly",
			isPrivileged:  true,
			command:       []string{"tail"},
			args:          []string{"-f", "/dev/null"},
			envVars: []corev1.EnvVar{
				{
					Name:  "test",
					Value: "123",
				},
			},
			resourceReqs: fakeResources,
			ports: []corev1.ContainerPort{
				{
					Name:          "port-9090",
					ContainerPort: 9090,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerParams := containerParams{
				Name:         tt.containerName,
				Image:        tt.image,
				IsPrivileged: tt.isPrivileged,
				Command:      tt.command,
				Args:         tt.args,
				EnvVars:      tt.envVars,
				ResourceReqs: tt.resourceReqs,
				Ports:        tt.ports,
			}
			container := getContainer(containerParams)

			if container.Name != tt.containerName {
				t.Errorf("expected %s, actual %s", tt.containerName, container.Name)
			}

			if container.Image != tt.image {
				t.Errorf("expected %s, actual %s", tt.image, container.Image)
			}

			if tt.isPrivileged {
				if *container.SecurityContext.Privileged != tt.isPrivileged {
					t.Errorf("expected %t, actual %t", tt.isPrivileged, *container.SecurityContext.Privileged)
				}
			} else if tt.isPrivileged == false && container.SecurityContext != nil {
				t.Errorf("expected security context to be nil but it was defined")
			}

			if len(container.Command) != len(tt.command) {
				t.Errorf("expected %d, actual %d", len(tt.command), len(container.Command))
			} else {
				for i := range container.Command {
					if container.Command[i] != tt.command[i] {
						t.Errorf("expected %s, actual %s", tt.command[i], container.Command[i])
					}
				}
			}

			if len(container.Args) != len(tt.args) {
				t.Errorf("expected %d, actual %d", len(tt.args), len(container.Args))
			} else {
				for i := range container.Args {
					if container.Args[i] != tt.args[i] {
						t.Errorf("expected %s, actual %s", tt.args[i], container.Args[i])
					}
				}
			}

			if len(container.Env) != len(tt.envVars) {
				t.Errorf("expected %d, actual %d", len(tt.envVars), len(container.Env))
			} else {
				for i := range container.Env {
					if container.Env[i].Name != tt.envVars[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.envVars[i].Name, container.Env[i].Name)
					}
					if container.Env[i].Value != tt.envVars[i].Value {
						t.Errorf("expected value %s, actual value %s", tt.envVars[i].Value, container.Env[i].Value)
					}
				}
			}

			if len(container.Ports) != len(tt.ports) {
				t.Errorf("expected %d, actual %d", len(tt.ports), len(container.Ports))
			} else {
				for i := range container.Ports {
					if container.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, container.Ports[i].Name)
					}
					if container.Ports[i].ContainerPort != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, container.Ports[i].ContainerPort)
					}
				}
			}

		})
	}
}

func TestGenerateServiceSpec(t *testing.T) {
	port1 := corev1.ContainerPort{
		Name:          "port-9090",
		ContainerPort: 9090,
	}
	port2 := corev1.ContainerPort{
		Name:          "port-8080",
		ContainerPort: 8080,
	}

	tests := []struct {
		name  string
		ports []corev1.ContainerPort
	}{
		{
			name:  "singlePort",
			ports: []corev1.ContainerPort{port1},
		},
		{
			name:  "multiplePorts",
			ports: []corev1.ContainerPort{port1, port2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceSpecParams := serviceSpecParams{
				ContainerPorts: tt.ports,
				SelectorLabels: map[string]string{
					"component": tt.name,
				},
			}
			serviceSpec := getServiceSpec(serviceSpecParams)

			if len(serviceSpec.Ports) != len(tt.ports) {
				t.Errorf("expected service ports length is %v, actual %v", len(tt.ports), len(serviceSpec.Ports))
			} else {
				for i := range serviceSpec.Ports {
					if serviceSpec.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, serviceSpec.Ports[i].Name)
					}
					if serviceSpec.Ports[i].Port != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, serviceSpec.Ports[i].Port)
					}
				}
			}

		})
	}
}
