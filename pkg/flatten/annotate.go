package flatten

import (
	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
)

const (
	ImportSourceAttribute = "library.devfile.io/imported-by"
)

// AddSourceAttributesForTemplate adds an attribute 'library.devfile.io/imported-by=<plugin-name>' to all elements of
// a plugin that support attributes.
func AddSourceAttributesForTemplate(sourceID string, template *dw.DevWorkspaceTemplateSpec) {
	for idx, component := range template.Components {
		if component.Attributes == nil {
			template.Components[idx].Attributes = attributes.Attributes{}
		}
		template.Components[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, command := range template.Commands {
		if command.Attributes == nil {
			template.Commands[idx].Attributes = attributes.Attributes{}
		}
		template.Commands[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, project := range template.Projects {
		if project.Attributes == nil {
			template.Projects[idx].Attributes = attributes.Attributes{}
		}
		template.Projects[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
	for idx, project := range template.StarterProjects {
		if project.Attributes == nil {
			template.StarterProjects[idx].Attributes = attributes.Attributes{}
		}
		template.StarterProjects[idx].Attributes.PutString(ImportSourceAttribute, sourceID)
	}
}
