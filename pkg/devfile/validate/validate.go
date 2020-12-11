package validate

import (
	"k8s.io/klog"
)

// ValidateDevfileData validates whether sections of devfile are compatible
func ValidateDevfileData(data interface{}) error {

	// Skipped
	klog.V(4).Info("No validation present. Skipped for the moment.")
	return nil

}
