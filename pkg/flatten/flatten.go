//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package flatten

import (
	"context"
	"fmt"
	"net/url"
	"path"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/utils/overriding"
	"github.com/devfile/library/pkg/flatten/network"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolverTools contains required structs and data for resolving remote components of a devfile (plugins and parents)
type ResolverTools struct {
	// DefaultNamespace is the default namespace to use for resolving Kubernetes ImportReferences that do not include one
	DefaultNamespace string
	// Context is the context used for making Kubernetes or HTTP requests
	Context context.Context
	// K8sClient is the Kubernetes client instance used for interacting with a cluster
	K8sClient client.Client
	// HttpClient is the HTTP client used for making network requests when resolving plugins or parents.
	HttpClient network.HTTPGetter
}

// ResolveDevWorkspace takes a DevWorkspaceTemplateSpec and returns a "resolved" version of it -- i.e. one where all plugins and parents
// are inlined as components.
// TODO:
// - Implement flattening for DevWorkspace parents
func ResolveDevWorkspace(workspace *devfile.DevWorkspaceTemplateSpec, tooling ResolverTools) (*devfile.DevWorkspaceTemplateSpec, error) {
	resolutionCtx := &resolutionContextTree{}
	resolvedDW, err := recursiveResolve(workspace, tooling, resolutionCtx)
	if err != nil {
		return nil, err
	}
	return resolvedDW, nil
}

// recursiveResolve recursively resolves plugins and parents until the result contains no parents or plugin components.
// This is a recursive function, where resolveCtx is used to build a tree of resolved components. This is used to avoid
// plugin or parent import cycles.
func recursiveResolve(workspace *devfile.DevWorkspaceTemplateSpec, tooling ResolverTools, resolveCtx *resolutionContextTree) (*devfile.DevWorkspaceTemplateSpec, error) {
	if DevWorkspaceIsFlattened(workspace) {
		return workspace.DeepCopy(), nil
	}
	if workspace.Parent != nil {
		// TODO: Add support for flattening DevWorkspace parents
		return nil, fmt.Errorf("DevWorkspace parent is unsupported")
	}

	resolvedContent := &devfile.DevWorkspaceTemplateSpecContent{}
	resolvedContent.Projects = workspace.Projects
	resolvedContent.StarterProjects = workspace.StarterProjects
	resolvedContent.Commands = workspace.Commands
	resolvedContent.Events = workspace.Events

	var pluginSpecContents []*devfile.DevWorkspaceTemplateSpecContent
	for _, component := range workspace.Components {
		if component.Plugin == nil {
			// No action necessary
			resolvedContent.Components = append(resolvedContent.Components, component)
		} else {
			pluginComponent, err := resolvePluginComponent(component.Name, component.Plugin, tooling)
			if err != nil {
				return nil, err
			}
			newCtx := resolveCtx.addPlugin(component.Name, component.Plugin)
			if err := newCtx.hasCycle(); err != nil {
				return nil, err
			}

			resolvedPlugin, err := recursiveResolve(pluginComponent, tooling, newCtx)
			if err != nil {
				return nil, err
			}

			AddSourceAttributesForPlugin(component.Name, resolvedPlugin)
			pluginSpecContents = append(pluginSpecContents, &resolvedPlugin.DevWorkspaceTemplateSpecContent)
		}
	}

	resolvedContent, err := overriding.MergeDevWorkspaceTemplateSpec(resolvedContent, nil, pluginSpecContents...)
	if err != nil {
		return nil, fmt.Errorf("failed to merge DevWorkspace parents/plugins: %w", err)
	}

	return &devfile.DevWorkspaceTemplateSpec{
		DevWorkspaceTemplateSpecContent: *resolvedContent,
	}, nil
}

// resolvePluginComponent resolves the DevWorkspaceTemplateSpec that a plugin component refers to. The name parameter is
// used to construct meaningful error messages (e.g. issue resolving plugin 'name')
func resolvePluginComponent(
	name string,
	plugin *devfile.PluginComponent,
	tooling ResolverTools) (resolvedPlugin *devfile.DevWorkspaceTemplateSpec, err error) {
	switch {
	case plugin.Kubernetes != nil:
		// Search in default namespace if namespace ref is unset
		if plugin.Kubernetes.Namespace == "" {
			plugin.Kubernetes.Namespace = tooling.DefaultNamespace
		}
		resolvedPlugin, err = resolvePluginComponentByKubernetesReference(name, plugin, tooling)
	case plugin.Uri != "":
		resolvedPlugin, err = resolvePluginComponentByURI(name, plugin, tooling)
	case plugin.Id != "":
		resolvedPlugin, err = resolvePluginComponentById(name, plugin, tooling)
	default:
		err = fmt.Errorf("plugin %s does not define any resources", name)
	}
	if err != nil {
		return nil, err
	}

	if plugin.Components != nil || plugin.Commands != nil {
		overrideSpec, err := overriding.OverrideDevWorkspaceTemplateSpec(&resolvedPlugin.DevWorkspaceTemplateSpecContent, devfile.PluginOverrides{
			Components: plugin.Components,
			Commands:   plugin.Commands,
		})

		if err != nil {
			return nil, err
		}
		resolvedPlugin.DevWorkspaceTemplateSpecContent = *overrideSpec
	}
	return resolvedPlugin, nil
}

// resolvePluginComponentByKubernetesReference resolves a plugin specified by a Kubernetes reference.
// The name parameter is used to construct meaningful error messages (e.g. issue resolving plugin 'name')
func resolvePluginComponentByKubernetesReference(
	name string,
	plugin *devfile.PluginComponent,
	tooling ResolverTools) (resolvedPlugin *devfile.DevWorkspaceTemplateSpec, err error) {

	var dwTemplate devfile.DevWorkspaceTemplate
	namespacedName := types.NamespacedName{
		Name:      plugin.Kubernetes.Name,
		Namespace: plugin.Kubernetes.Namespace,
	}
	err = tooling.K8sClient.Get(tooling.Context, namespacedName, &dwTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("plugin for component %s not found", name)
		}
		return nil, fmt.Errorf("failed to retrieve plugin referenced by kubernetes name and namespace '%s': %w", name, err)
	}
	return &dwTemplate.Spec, nil
}

// resolvePluginComponentById resolves a plugin specified by plugin ID and registry URL. The name parameter is used to
// construct meaningful error messages (e.g. issue resolving plugin 'name')
func resolvePluginComponentById(
	name string,
	plugin *devfile.PluginComponent,
	tools ResolverTools) (resolvedPlugin *devfile.DevWorkspaceTemplateSpec, err error) {

	// TODO: Default registry when empty
	pluginURL, err := url.Parse(plugin.RegistryUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry URL for plugin %s: %w", name, err)
	}
	pluginURL.Path = path.Join(pluginURL.Path, "plugins", plugin.Id)

	dwt, err := network.FetchDevWorkspaceTemplate(pluginURL.String(), tools.HttpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve plugin %s from registry %s: %w", name, plugin.RegistryUrl, err)
	}
	return dwt, nil
}

// resolvePluginComponentByURI resolves a plugin defined by URI. The name parameter is used to construct meaningful
// error messages (e.g. issue resolving plugin 'name')
func resolvePluginComponentByURI(
	name string,
	plugin *devfile.PluginComponent,
	tools ResolverTools) (resolvedPlugin *devfile.DevWorkspaceTemplateSpec, err error) {

	dwt, err := network.FetchDevWorkspaceTemplate(plugin.Uri, tools.HttpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve plugin %s by URI: %w", name, err)
	}
	return dwt, nil
}
