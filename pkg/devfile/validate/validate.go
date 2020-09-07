package validate

import (
	"fmt"

	"k8s.io/klog"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
func ValidateDevfileData(data interface{}) error {
	var components []v1.Component

	switch d := data.(type) {
	case *v200.Devfile200:
		components = d.GetComponents()

		// Validate Events
		if err := validateEvents(d); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	// Validate Components
	if err := validateComponents(components); err != nil {
		return err
	}

	// Successful
	klog.V(4).Info("Successfully validated devfile sections")
	return nil

}
