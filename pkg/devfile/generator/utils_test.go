package generator

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	buildv1 "github.com/openshift/api/build/v1"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func TestGetPodTemplateSpec(t *testing.T) {

	container := []corev1.Container{
		{
			Name:            "container1",
			Image:           "image1",
			ImagePullPolicy: corev1.PullAlways,

			Command: []string{"tail"},
			Args:    []string{"-f", "/dev/null"},
			Env:     []corev1.EnvVar{},
		},
	}

	volume := []corev1.Volume{
		{
			Name: "vol1",
		},
	}

	tests := []struct {
		podName        string
		namespace      string
		serviceAccount string
		labels         map[string]string
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {

			objectMeta := GetObjectMeta(tt.podName, tt.namespace, tt.labels, nil)
			podTemplateSpecParams := podTemplateSpecParams{
				ObjectMeta:     objectMeta,
				Containers:     container,
				Volumes:        volume,
				InitContainers: container,
			}
			podTemplateSpec := getPodTemplateSpec(podTemplateSpecParams)

			if podTemplateSpec.Name != tt.podName {
				t.Errorf("expected %s, actual %s", tt.podName, podTemplateSpec.Name)
			}
			if podTemplateSpec.Namespace != tt.namespace {
				t.Errorf("expected %s, actual %s", tt.namespace, podTemplateSpec.Namespace)
			}
			if !hasVolumeWithName("vol1", podTemplateSpec.Spec.Volumes) {
				t.Errorf("volume with name: %s not found", "vol1")
			}
			if !reflect.DeepEqual(podTemplateSpec.Labels, tt.labels) {
				t.Errorf("expected %+v, actual %+v", tt.labels, podTemplateSpec.Labels)
			}
			if !reflect.DeepEqual(podTemplateSpec.Spec.Containers, container) {
				t.Errorf("expected %+v, actual %+v", container, podTemplateSpec.Spec.Containers)
			}
			if !reflect.DeepEqual(podTemplateSpec.Spec.InitContainers, container) {
				t.Errorf("expected %+v, actual %+v", container, podTemplateSpec.Spec.InitContainers)
			}
		})
	}
}

func TestGetServiceSpec(t *testing.T) {

	endpointNames := []string{"port-8080-1", "port-8080-2", "port-9090"}

	tests := []struct {
		name                string
		containerComponents []v1.Component
		labels              map[string]string
		wantPorts           []corev1.ServicePort
		wantErr             bool
	}{
		{
			name: "Case 1: multiple endpoints share the same port",
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Endpoints: []v1.Endpoint{
								{
									Name:       endpointNames[0],
									TargetPort: 8080,
								},
								{
									Name:       endpointNames[1],
									TargetPort: 8080,
								},
							},
						},
					},
				},
			},
			labels: map[string]string{},
			wantPorts: []corev1.ServicePort{
				{
					Name:       "port-8080",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: multiple endpoints have different ports",
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Endpoints: []v1.Endpoint{
								{
									Name:       endpointNames[0],
									TargetPort: 8080,
								},
								{
									Name:       endpointNames[2],
									TargetPort: 9090,
								},
							},
						},
					},
				},
			},
			labels: map[string]string{
				"component": "testcomponent",
			},
			wantPorts: []corev1.ServicePort{
				{
					Name:       "port-8080",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "port-9090",
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.containerComponents,
				},
			}

			serviceSpec, err := getServiceSpec(devObj, tt.labels)

			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetServiceSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Expected error and got an err
			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(serviceSpec.Selector, tt.labels) {
				t.Errorf("expected service selector is %v, actual %v", tt.labels, serviceSpec.Selector)
			}
			if len(serviceSpec.Ports) != len(tt.wantPorts) {
				t.Errorf("expected service ports length is %v, actual %v", len(tt.wantPorts), len(serviceSpec.Ports))
			} else {
				for i := range serviceSpec.Ports {
					if serviceSpec.Ports[i].Name != tt.wantPorts[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.wantPorts[i].Name, serviceSpec.Ports[i].Name)
					}
					if serviceSpec.Ports[i].Port != tt.wantPorts[i].Port {
						t.Errorf("expected port number is %v, actual %v", tt.wantPorts[i].Port, serviceSpec.Ports[i].Port)
					}
				}
			}
		})
	}
}

func TestGetPortExposure(t *testing.T) {
	urlName := "testurl"
	urlName2 := "testurl2"
	tests := []struct {
		name                string
		containerComponents []v1.Component
		wantMap             map[int]v1.EndpointExposure
		wantErr             bool
	}{
		{
			name: "Case 1: devfile has single container with single endpoint",
			wantMap: map[int]v1.EndpointExposure{
				8080: v1.PublicEndpointExposure,
			},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.PublicEndpointExposure,
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Case 2: devfile no endpoints",
			wantMap: map[int]v1.EndpointExposure{},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
						},
					},
				},
			},
		},
		{
			name: "Case 3: devfile has multiple endpoints with same port, 1 public and 1 internal, should assign public",
			wantMap: map[int]v1.EndpointExposure{
				8080: v1.PublicEndpointExposure,
			},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.PublicEndpointExposure,
								},
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.InternalEndpointExposure,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Case 4: devfile has multiple endpoints with same port, 1 public and 1 none, should assign public",
			wantMap: map[int]v1.EndpointExposure{
				8080: v1.PublicEndpointExposure,
			},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.PublicEndpointExposure,
								},
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.NoneEndpointExposure,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Case 5: devfile has multiple endpoints with same port, 1 internal and 1 none, should assign internal",
			wantMap: map[int]v1.EndpointExposure{
				8080: v1.InternalEndpointExposure,
			},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.InternalEndpointExposure,
								},
								{
									Name:       urlName,
									TargetPort: 8080,
									Exposure:   v1.NoneEndpointExposure,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Case 6: devfile has multiple endpoints with different port",
			wantMap: map[int]v1.EndpointExposure{
				8080: v1.PublicEndpointExposure,
				9090: v1.InternalEndpointExposure,
				3000: v1.NoneEndpointExposure,
			},
			containerComponents: []v1.Component{
				{
					Name: "testcontainer1",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName,
									TargetPort: 8080,
								},
								{
									Name:       urlName,
									TargetPort: 3000,
									Exposure:   v1.NoneEndpointExposure,
								},
							},
						},
					},
				},
				{
					Name: "testcontainer2",
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image: "image",
							},
							Endpoints: []v1.Endpoint{
								{
									Name:       urlName2,
									TargetPort: 9090,
									Secure:     true,
									Path:       "/testpath",
									Exposure:   v1.InternalEndpointExposure,
									Protocol:   v1.HTTPSEndpointProtocol,
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
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.containerComponents,
				},
			}
			mapCreated := getPortExposure(devObj)
			if !reflect.DeepEqual(mapCreated, tt.wantMap) {
				t.Errorf("Expected: %v, got %v", tt.wantMap, mapCreated)
			}

		})
	}

}

func TestGenerateIngressSpec(t *testing.T) {

	tests := []struct {
		name      string
		parameter IngressSpecParams
	}{
		{
			name: "1",
			parameter: IngressSpecParams{
				ServiceName:   "service1",
				IngressDomain: "test.1.2.3.4.nip.io",
				PortNumber: intstr.IntOrString{
					IntVal: 8080,
				},
				TLSSecretName: "testTLSSecret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ingressSpec := getIngressSpec(tt.parameter)

			if ingressSpec.Rules[0].Host != tt.parameter.IngressDomain {
				t.Errorf("expected %s, actual %s", tt.parameter.IngressDomain, ingressSpec.Rules[0].Host)
			}

			if ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServicePort != tt.parameter.PortNumber {
				t.Errorf("expected %v, actual %v", tt.parameter.PortNumber, ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServicePort)
			}

			if ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServiceName != tt.parameter.ServiceName {
				t.Errorf("expected %s, actual %s", tt.parameter.ServiceName, ingressSpec.Rules[0].HTTP.Paths[0].Backend.ServiceName)
			}

			if ingressSpec.TLS[0].SecretName != tt.parameter.TLSSecretName {
				t.Errorf("expected %s, actual %s", tt.parameter.TLSSecretName, ingressSpec.TLS[0].SecretName)
			}

		})
	}
}

func TestGetRouteSpec(t *testing.T) {

	tests := []struct {
		name      string
		parameter RouteSpecParams
	}{
		{
			name: "Case 1: insecure route",
			parameter: RouteSpecParams{
				ServiceName: "service1",
				PortNumber: intstr.IntOrString{
					IntVal: 8080,
				},
				Secure: false,
				Path:   "/test",
			},
		},
		{
			name: "Case 2: secure route",
			parameter: RouteSpecParams{
				ServiceName: "service1",
				PortNumber: intstr.IntOrString{
					IntVal: 8080,
				},
				Secure: true,
				Path:   "/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			routeSpec := getRouteSpec(tt.parameter)

			if routeSpec.Port.TargetPort != tt.parameter.PortNumber {
				t.Errorf("expected %v, actual %v", tt.parameter.PortNumber, routeSpec.Port.TargetPort)
			}

			if routeSpec.To.Name != tt.parameter.ServiceName {
				t.Errorf("expected %s, actual %s", tt.parameter.ServiceName, routeSpec.To.Name)
			}

			if routeSpec.Path != tt.parameter.Path {
				t.Errorf("expected %s, actual %s", tt.parameter.Path, routeSpec.Path)
			}

			if (routeSpec.TLS != nil) != tt.parameter.Secure {
				t.Errorf("the route TLS does not match secure level %v", tt.parameter.Secure)
			}

		})
	}
}

func TestGetPVCSpec(t *testing.T) {

	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{
			name:    "Case 1: Valid resource size",
			size:    "1Gi",
			wantErr: false,
		},
		{
			name:    "Case 2: Resource size missing",
			size:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			quantity, err := resource.ParseQuantity(tt.size)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("resource.ParseQuantity unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			pvcSpec := getPVCSpec(quantity)
			if pvcSpec.AccessModes[0] != corev1.ReadWriteOnce {
				t.Errorf("AccessMode Error: expected %s, actual %s", corev1.ReadWriteMany, pvcSpec.AccessModes[0])
			}

			pvcSpecQuantity := pvcSpec.Resources.Requests["storage"]
			if pvcSpecQuantity.String() != quantity.String() {
				t.Errorf("pvcSpec.Resources.Requests Error: expected %v, actual %v", pvcSpecQuantity.String(), quantity.String())
			}
		})
	}
}

func hasVolumeWithName(name string, volMounts []corev1.Volume) bool {
	for _, vm := range volMounts {
		if vm.Name == name {
			return true
		}
	}
	return false
}

func TestGetBuildConfigSpec(t *testing.T) {

	image := "image"
	namespace := "namespace"

	tests := []struct {
		name          string
		GitURL        string
		GitRef        string
		buildStrategy buildv1.BuildStrategy
	}{
		{
			name:          "Case 1: Get a Source Strategy Build Config",
			GitURL:        "url",
			GitRef:        "ref",
			buildStrategy: GetSourceBuildStrategy(image, namespace),
		},
		{
			name:          "Case 2: Get a Docker Strategy Build Config",
			GitURL:        "url",
			GitRef:        "ref",
			buildStrategy: GetDockerBuildStrategy("dockerfilePath", []corev1.EnvVar{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonObjectMeta := GetObjectMeta(image, namespace, nil, nil)
			params := BuildConfigSpecParams{
				ImageStreamTagName: commonObjectMeta.Name,
				BuildStrategy:      tt.buildStrategy,
				GitURL:             tt.GitURL,
				GitRef:             tt.GitRef,
			}
			buildConfigSpec := getBuildConfigSpec(params)

			if !strings.Contains(buildConfigSpec.CommonSpec.Output.To.Name, image) {
				t.Error("TestGetBuildConfigSpec error - build config output name does not match")
			}

			if buildConfigSpec.Source.Git.Ref != tt.GitRef || buildConfigSpec.Source.Git.URI != tt.GitURL {
				t.Error("TestGetBuildConfigSpec error - build config git source does not match")
			}
		})
	}

}
