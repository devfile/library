package data

import (
	"github.com/devfile/api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	SetSchemaVersion(version string)

	GetMetadata() common.DevfileMetadata
	SetMetadata(name, version string)

	// parent related methods
	GetParent() *v1alpha1.Parent
	SetParent(parent *v1alpha1.Parent)

	// event related methods
	GetEvents() v1alpha1.Events
	AddEvents(events v1alpha1.Events) error
	UpdateEvents(postStart, postStop, preStart, preStop []string)

	// component related methods
	GetComponents() []v1alpha1.Component
	AddComponents(components []v1alpha1.Component) error
	UpdateComponent(component v1alpha1.Component)
	GetAliasedComponents() []v1alpha1.Component

	// project related methods
	GetProjects() []v1alpha1.Project
	AddProjects(projects []v1alpha1.Project) error
	UpdateProject(project v1alpha1.Project)

	// command related methods
	GetCommands() []v1alpha1.Command
	AddCommands(commands []v1alpha1.Command) error
	UpdateCommand(command v1alpha1.Command)
}
