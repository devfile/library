package parser

import (
	"fmt"
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
)

const ImportSourceAttribute = "library.devfile.io/imported-from"

// AddSourceAttributesForTemplateSpecContent adds an attribute 'library.devfile.io/imported-from=<source reference>'
//  to all elements of template spec content that support attributes.
func AddSourceAttributesForTemplateSpecContent(sourceImportReference v1.ImportReference, template *v1.DevWorkspaceTemplateSpecContent) {
	for idx, component := range template.Components {
		if component.Attributes == nil {
			template.Components[idx].Attributes = attributes.Attributes{}
		}
		template.Components[idx].Attributes.PutString(ImportSourceAttribute, resolveImportReference(sourceImportReference))
	}
	for idx, command := range template.Commands {
		if command.Attributes == nil {
			template.Commands[idx].Attributes = attributes.Attributes{}
		}
		template.Commands[idx].Attributes.PutString(ImportSourceAttribute, resolveImportReference(sourceImportReference))
	}
	for idx, project := range template.Projects {
		if project.Attributes == nil {
			template.Projects[idx].Attributes = attributes.Attributes{}
		}
		template.Projects[idx].Attributes.PutString(ImportSourceAttribute, resolveImportReference(sourceImportReference))
	}
	for idx, project := range template.StarterProjects {
		if project.Attributes == nil {
			template.StarterProjects[idx].Attributes = attributes.Attributes{}
		}
		template.StarterProjects[idx].Attributes.PutString(ImportSourceAttribute, resolveImportReference(sourceImportReference))
	}
}


// AddSourceAttributesForParentOverride adds an attribute 'library.devfile.io/imported-from=<source reference>'
//  to all elements of parent override that support attributes.
func AddSourceAttributesForParentOverride(sourceImportReference v1.ImportReference, parentoverride *v1.ParentOverrides) {
	for idx, component := range parentoverride.Components {
		if component.Attributes == nil {
			parentoverride.Components[idx].Attributes = attributes.Attributes{}
		}
		parentoverride.Components[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("parentOverrides from: %s", resolveImportReference(sourceImportReference)))
	}
	for idx, command := range parentoverride.Commands {
		if command.Attributes == nil {
			parentoverride.Commands[idx].Attributes = attributes.Attributes{}
		}
		parentoverride.Commands[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("parentOverrides from: %s", resolveImportReference(sourceImportReference)))
	}
	for idx, project := range parentoverride.Projects {
		if project.Attributes == nil {
			parentoverride.Projects[idx].Attributes = attributes.Attributes{}
		}
		parentoverride.Projects[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("parentOverrides from: %s", resolveImportReference(sourceImportReference)))
	}
	for idx, project := range parentoverride.StarterProjects {
		if project.Attributes == nil {
			parentoverride.StarterProjects[idx].Attributes = attributes.Attributes{}
		}
		parentoverride.StarterProjects[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("parentOverrides from: %s", resolveImportReference(sourceImportReference)))
	}

}


// AddSourceAttributesForPluginOverride adds an attribute 'library.devfile.io/imported-from=<source reference>'
//  to all elements of plugin override that support attributes.
func AddSourceAttributesForPluginOverride(sourceImportReference v1.ImportReference, pluginId string,  pluginoverride *v1.PluginOverrides) {
	for idx, component := range pluginoverride.Components {
		if component.Attributes == nil {
			pluginoverride.Components[idx].Attributes = attributes.Attributes{}
		}
		pluginoverride.Components[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("pluginOverrides from: %s, plugin : %s", resolveImportReference(sourceImportReference), pluginId))
	}
	for idx, command := range pluginoverride.Commands {
		if command.Attributes == nil {
			pluginoverride.Commands[idx].Attributes = attributes.Attributes{}
		}
		pluginoverride.Commands[idx].Attributes.PutString(ImportSourceAttribute, fmt.Sprintf("pluginOverrides from: %s, plugin : %s", resolveImportReference(sourceImportReference), pluginId))
	}

}