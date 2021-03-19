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
	"fmt"
	"reflect"

	devworkspace "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// resolutionContextTree is a recursive structure representing information about the devfile that is
// lost when flattening (e.g. plugins, parents)
type resolutionContextTree struct {
	componentName   string
	importReference devworkspace.ImportReference
	plugins         []*resolutionContextTree
	parentNode      *resolutionContextTree
}

// addPlugin adds a plugin component to the resolution context.
func (t *resolutionContextTree) addPlugin(name string, plugin *devworkspace.PluginComponent) *resolutionContextTree {
	newNode := &resolutionContextTree{
		componentName:   name,
		importReference: plugin.ImportReference,
		parentNode:      t,
	}
	t.plugins = append(t.plugins, newNode)
	return newNode
}

// hasCycle checks if the current resolutionContextTree has a cycle
func (t *resolutionContextTree) hasCycle() error {
	var seenRefs []devworkspace.ImportReference
	currNode := t
	for currNode.parentNode != nil {
		for _, seenRef := range seenRefs {
			if reflect.DeepEqual(seenRef, currNode.importReference) {
				return fmt.Errorf("DevWorkspace has an cycle in references: %s", formatImportCycle(t))
			}
		}
		seenRefs = append(seenRefs, currNode.importReference)
		currNode = currNode.parentNode
	}
	return nil
}

// formatImportCycle is a utility method for formatting a cycle that has been detected. Output is formatted as
// plugin1 -> plugin2 -> plugin3 -> plugin1, where pluginX are component names.
func formatImportCycle(end *resolutionContextTree) string {
	cycle := fmt.Sprintf("%s", end.componentName)
	for end.parentNode != nil {
		end = end.parentNode
		if end.parentNode == nil {
			end.componentName = "devfile"
		}
		cycle = fmt.Sprintf("%s -> %s", end.componentName, cycle)
	}
	return cycle
}
