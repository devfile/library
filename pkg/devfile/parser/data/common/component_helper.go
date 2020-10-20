package common

import (
	"fmt"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"
)

// IsContainer checks if the component is a container
func IsContainer(component v1.Component) bool {
	if component.Container != nil {
		klog.V(2).Infof("Found component \"%v\" with name \"%v\"\n", v1.ContainerComponentType, component.Name)
		return true
	}
	return false
}

// IsVolume checks if the component is a volume
func IsVolume(component v1.Component) bool {
	if component.Volume != nil {
		klog.V(2).Infof("Found component \"%v\" with name \"%v\"\n", v1.VolumeComponentType, component.Name)
		return true
	}
	return false
}

// GetComponentType returns the component type of a given component
func GetComponentType(component v1.Component) (v1.ComponentType, error) {
	if component.Container != nil {
		return v1.ContainerComponentType, nil
	}
	if component.Volume != nil {
		return v1.VolumeComponentType, nil
	}
	if component.Plugin != nil {
		return v1.PluginComponentType, nil
	}
	if component.Kubernetes != nil {
		return v1.KubernetesComponentType, nil
	}
	if component.Openshift != nil {
		return v1.OpenshiftComponentType, nil
	}
	if component.Custom != nil {
		return v1.CustomComponentType, nil
	}
	return "", fmt.Errorf("Unknown component type")
}
