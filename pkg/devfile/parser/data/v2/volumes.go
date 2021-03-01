package v2

import (
	"fmt"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// AddVolume adds the volume to the devFile and mounts it to all the container components
func (d *DevfileV2) AddVolume(volumeComponent v1.Component, path string) error {
	volumeExists := false
	var pathErrorContainers []string
	for _, component := range d.Components {
		if component.Container != nil {
			for _, volumeMount := range component.Container.VolumeMounts {
				if volumeMount.Path == path {
					var err = fmt.Errorf("another volume, %s, is mounted to the same path: %s, on the container: %s", volumeMount.Name, path, component.Name)
					pathErrorContainers = append(pathErrorContainers, err.Error())
				}
			}
			component.Container.VolumeMounts = append(component.Container.VolumeMounts, v1.VolumeMount{
				Name: volumeComponent.Name,
				Path: path,
			})
		} else if component.Volume != nil && component.Name == volumeComponent.Name {
			volumeExists = true
			break
		}
	}

	if volumeExists {
		return &common.FieldAlreadyExistError{
			Field: "volume",
			Name:  volumeComponent.Name,
		}
	}

	if len(pathErrorContainers) > 0 {
		return fmt.Errorf("errors while creating volume:\n%s", strings.Join(pathErrorContainers, "\n"))
	}

	d.Components = append(d.Components, volumeComponent)

	return nil
}

// DeleteVolume removes the volume from the devFile and removes all the related volume mounts
func (d *DevfileV2) DeleteVolume(name string) error {

	return d.DeleteComponent(name)
}

// GetVolumeMountPath gets the mount path of the required volume
func (d *DevfileV2) GetVolumeMountPath(name string) (string, error) {
	volumeFound := false
	mountFound := false
	path := ""

	for _, component := range d.Components {
		if component.Container != nil {
			for _, volumeMount := range component.Container.VolumeMounts {
				if volumeMount.Name == name {
					mountFound = true
					path = volumeMount.Path
				}
			}
		} else if component.Volume != nil {
			volumeFound = true
		}
	}
	if volumeFound && mountFound {
		return path, nil
	} else if !mountFound && volumeFound {
		return "", fmt.Errorf("volume not mounted to any component")
	}
	return "", &common.FieldNotFoundError{
		Field: "volume",
		Name:  "name",
	}
}
