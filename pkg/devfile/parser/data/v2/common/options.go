package common

import (
	"reflect"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
)

// DevfileOptions provides options for Devfile operations
type DevfileOptions struct {
	// Filter is a map that lets you filter devfile object against their attributes. Interface can be string, float, boolean or a map
	Filter map[string]interface{}

	// CommandOptions specifies the various options available to filter commands
	CommandOptions CommandOptions

	// ComponentOptions specifies the various options available to filter components
	ComponentOptions ComponentOptions
}

// CommandOptions specifies the various options available to filter commands
type CommandOptions struct {
	// CommandKind is an option that allows you to filter command based on their kind
	CommandKind v1.CommandGroupKind

	// CommandType is an option that allows you to filter command based on their type
	CommandType v1.CommandType
}

// ComponentOptions specifies the various options available to filter components
type ComponentOptions struct {

	// ComponentType is an option that allows you to filter component based on their type
	ComponentType v1.ComponentType
}

// FilterDevfileObject filters devfile attributes with the given options
func FilterDevfileObject(attributes apiAttributes.Attributes, options DevfileOptions) (bool, error) {
	filterIn := true
	for key, value := range options.Filter {
		var err error
		currentFilterIn := false
		attrValue := attributes.Get(key, &err)
		var keyNotFoundErr = &apiAttributes.KeyNotFoundError{Key: key}
		if err != nil && err.Error() != keyNotFoundErr.Error() {
			return false, err
		} else if reflect.DeepEqual(attrValue, value) {
			currentFilterIn = true
		}

		filterIn = filterIn && currentFilterIn
	}

	return filterIn, nil
}
