package v2

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/attributes"
)

// GetAttributes gets the devfile top level attributes
func (d *DevfileV2) GetAttributes() (attributes.Attributes, error) {
	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return attributes.Attributes{}, fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		return d.Attributes, nil
	}
}

// UpdateAttributes updates the devfile top level attributes
func (d *DevfileV2) UpdateAttributes(attr attributes.Attributes) error {
	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		d.Attributes = attr
	}

	return nil
}

// AddAttributes adds to the devfile top level attributes
func (d *DevfileV2) AddAttributes(key string, value interface{}) error {
	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		var err error
		d.Attributes.Put(key, value, &err)
		if err != nil {
			return err
		}
	}

	return nil
}
