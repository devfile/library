package data

import (
	apiComp "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetMetadata() common.DevfileMetadata
	GetParent() apiComp.Parent
	GetEvents() apiComp.WorkspaceEvents
	GetComponents() []common.DevfileComponent
	GetAliasedComponents() []common.DevfileComponent
	GetProjects() []apiComp.Project
	GetCommands() []common.DevfileCommand
}
