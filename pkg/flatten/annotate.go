package flatten

import (
	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
)

const (
	ImportSourceAttribute = "library.devfile.io/imported-by"
)

// AddSourceAttributesForPlugin adds an attribute 'library.devfile.io/imported-by=<plugin-name>' to all elements of
// a plugin that support attributes.
func AddSourceAttributesForPlugin(sourceID string, plugin *dw.DevWorkspaceTemplateSpec) {
	for idx, component := range plugin.Components {
		if component.Attributes == nil {
			plugin.Components[idx].Attributes = attributes.Attributes{}
		}
		plugin.Components[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, command := range plugin.Commands {
		if command.Attributes == nil {
			plugin.Commands[idx].Attributes = attributes.Attributes{}
		}
		plugin.Commands[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, project := range plugin.Projects {
		if project.Attributes == nil {
			plugin.Projects[idx].Attributes = attributes.Attributes{}
		}
		plugin.Projects[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, project := range plugin.StarterProjects {
		if project.Attributes == nil {
			plugin.Projects[idx].Attributes = attributes.Attributes{}
		}
		plugin.Projects[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
}
