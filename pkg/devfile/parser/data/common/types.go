package common

// DevfileMetadata metadata for devfile
type DevfileMetadata struct {

	// Name Optional devfile name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Version Optional semver-compatible version
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}
