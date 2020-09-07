package validate

import (
	"reflect"
	"strings"
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []v1.Component{}

		got := validateComponents(components)
		want := &NoComponentsError{}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestValidateComponents error - got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Container type of component present", func(t *testing.T) {

		components := []v1.Component{
			{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Name: "container",
					},
				},
			},
		}

		got := validateComponents(components)

		if got != nil {
			t.Errorf("TestValidateComponents error - Not expecting an error: '%v'", got)
		}
	})

	t.Run("Duplicate volume components present", func(t *testing.T) {

		components := []v1.Component{
			{
				Volume: &v1.VolumeComponent{
					Volume: v1.Volume{
						Name: "myvol",
						Size: "1Gi",
					},
				},
			},
			{
				Volume: &v1.VolumeComponent{
					Volume: v1.Volume{
						Name: "myvol",
						Size: "1Gi",
					},
				},
			},
		}

		got := validateComponents(components)
		want := &DuplicateVolumeComponentsError{}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestValidateComponents error - got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Valid container and volume component", func(t *testing.T) {

		components := []v1.Component{
			{
				Volume: &v1.VolumeComponent{
					Volume: v1.Volume{
						Name: "myvol",
						Size: "1Gi",
					},
				},
			},
			{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Name: "container",
						VolumeMounts: []v1.VolumeMount{
							{
								Name: "myvol",
								Path: "/some/path/",
							},
						},
					},
				},
			},
			{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Name: "container2",
						VolumeMounts: []v1.VolumeMount{
							{
								Name: "myvol",
							},
						},
					},
				},
			},
		}

		got := validateComponents(components)

		if got != nil {
			t.Errorf("TestValidateComponents error - got: '%v'", got)
		}
	})

	t.Run("Invalid volume component size", func(t *testing.T) {

		components := []v1.Component{
			{
				Volume: &v1.VolumeComponent{
					Volume: v1.Volume{
						Name: "myvol",
						Size: "randomgarbage",
					},
				},
			},
			{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Name: "container",
						VolumeMounts: []v1.VolumeMount{
							{
								Name: "myvol",
								Path: "/some/path/",
							},
						},
					},
				},
			},
		}

		got := validateComponents(components)
		want := "size randomgarbage for volume component myvol is invalid"

		if !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substring: '%v'", got.Error(), want)
		}
	})

	t.Run("Invalid volume mount", func(t *testing.T) {

		components := []v1.Component{
			{
				Volume: &v1.VolumeComponent{
					Volume: v1.Volume{
						Name: "myvol",
						Size: "2Gi",
					},
				},
			},
			{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Name: "container",
						VolumeMounts: []v1.VolumeMount{
							{
								Name: "myinvalidvol",
							},
							{
								Name: "myinvalidvol2",
							},
						},
					},
				},
			},
		}

		got := validateComponents(components)
		want := "unable to find volume mount"

		if !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substr: '%v'", got.Error(), want)
		}
	})
}
