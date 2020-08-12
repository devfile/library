package version200

import (
	v1 "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// Devfile200 Devfile schema.
type Devfile200 struct {

	// Predefined, ready-to-use, workspace-related commands
	Commands []v1.Command `json:"commands,omitempty"`

	// List of the workspace components, such as editor and plugins, user-provided containers, or other types of components
	Components []common.DevfileComponent `json:"components,omitempty"`

	// Bindings of commands to events. Each command is referred-to by its name.
	Events v1.WorkspaceEvents `json:"events,omitempty"`

	// Optional metadata
	Metadata common.DevfileMetadata `json:"metadata,omitempty"`

	// Parent workspace template
	Parent v1.Parent `json:"parent,omitempty"`

	// Projects worked on in the workspace, containing names and sources locations
	Projects []v1.Project `json:"projects,omitempty"`

	// Devfile schema version
	SchemaVersion string `json:"schemaVersion"`
}
