package data

import (
	v1 "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetMetadata() common.DevfileMetadata
	GetParent() v1.Parent
	GetEvents() v1.WorkspaceEvents
	GetComponents() []common.DevfileComponent
	GetProjects() []v1.Project
	GetCommands() []v1.Command
}
