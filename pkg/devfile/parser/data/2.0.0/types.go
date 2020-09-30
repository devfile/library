package version200

import (
	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// Devfile200 Devfile schema.
type Devfile200 struct {
	// Devfile schema version
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`

	// Optional metadata
	Metadata v1.DevfileMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Parent workspace template
	Parent *v1.Parent `json:"parent,omitempty" yaml:"parent,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects []v1.Project `json:"projects,omitempty" yaml:"projects,omitempty"`

	// StarterProjects is a project that can be used as a starting point when bootstrapping new projects
	StarterProjects []v1.StarterProject `json:"starterProjects,omitempty" yaml:"starterProjects,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []v1.Component `json:"components,omitempty" yaml:"components,omitempty"`

	// Predefined, ready-to-use, workspace-related commands
	Commands []v1.Command `json:"commands,omitempty" yaml:"commands,omitempty"`

	// Bindings of commands to events. Each command is referred-to by its name.
	Events v1.Events `json:"events,omitempty" yaml:"events,omitempty"`
}
