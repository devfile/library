package generator

import (
	"reflect"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
)

var fakeResources corev1.ResourceRequirements

func init() {
	fakeResources, _ = testingutil.FakeResourceRequirements("0.5m", "300Mi")
}

func TestGetContainers(t *testing.T) {

	containerNames := []string{"testcontainer1", "testcontainer2"}
	containerImages := []string{"image1", "image2"}
	trueMountSources := true
	falseMountSources := false
	tests := []struct {
		name                  string
		containerComponents   []v1.Component
		wantContainerName     string
		wantContainerImage    string
		wantContainerEnv      []corev1.EnvVar
		wantContainerVolMount []corev1.VolumeMount
		wantErr               bool
	}{
		{
			name: "Case 1: Container with default project root",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &trueMountSources,
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{

				{
					Name:  "PROJECTS_ROOT",
					Value: "/projects",
				},
				{
					Name:  "PROJECT_SOURCE",
					Value: "/projects/test-project",
				},
			},
			wantContainerVolMount: []corev1.VolumeMount{
				{
					Name:      "devfile-projects",
					MountPath: "/projects",
				},
			},
		},
		{
			name: "Case 2: Container with source mapping",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:         containerImages[0],
								MountSources:  &trueMountSources,
								SourceMapping: "/myroot",
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
			wantContainerEnv: []corev1.EnvVar{

				{
					Name:  "PROJECTS_ROOT",
					Value: "/myroot",
				},
				{
					Name:  "PROJECT_SOURCE",
					Value: "/myroot/test-project",
				},
			},
			wantContainerVolMount: []corev1.VolumeMount{
				{
					Name:      "devfile-projects",
					MountPath: "/myroot",
				},
			},
		},
		{
			name: "Case 3: Container with no mount source",
			containerComponents: []v1.Component{
				{
					Name: containerNames[0],
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: v1.Container{
								Image:        containerImages[0],
								MountSources: &falseMountSources,
							},
						},
					},
				},
			},
			wantContainerName:  containerNames[0],
			wantContainerImage: containerImages[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.containerComponents,
				},
			}

			containers, err := GetContainers(devObj)
			// Unexpected error
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetContainers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Expected error and got an err
			if tt.wantErr && err != nil {
				return
			}

			for _, container := range containers {
				if container.Name != tt.wantContainerName {
					t.Errorf("TestGetContainers error: Name mismatch - got: %s, wanted: %s", container.Name, tt.wantContainerName)
				}
				if container.Image != tt.wantContainerImage {
					t.Errorf("TestGetContainers error: Image mismatch - got: %s, wanted: %s", container.Image, tt.wantContainerImage)
				}
				if len(container.Env) > 0 && !reflect.DeepEqual(container.Env, tt.wantContainerEnv) {
					t.Errorf("TestGetContainers error: Env mismatch - got: %+v, wanted: %+v", container.Env, tt.wantContainerEnv)
				}
				if len(container.VolumeMounts) > 0 && !reflect.DeepEqual(container.VolumeMounts, tt.wantContainerVolMount) {
					t.Errorf("TestGetContainers error: Vol Mount mismatch - got: %+v, wanted: %+v", container.VolumeMounts, tt.wantContainerVolMount)
				}
			}
		})
	}

}
