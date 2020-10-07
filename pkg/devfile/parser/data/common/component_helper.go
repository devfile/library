package common

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"
)

// IsContainer checks if the component is a container
func IsContainer(component v1.Component) bool {
	// Currently odo only uses devfile components of type container, since most of the Che registry devfiles use it
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
