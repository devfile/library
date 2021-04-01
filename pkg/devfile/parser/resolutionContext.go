package parser

import (
	"fmt"
	"reflect"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// resolutionContextTree is a recursive structure representing information about the devfile that is
// lost when flattening (e.g. plugins, parents)
type resolutionContextTree struct {
	importReference v1.ImportReference
	parentNode      *resolutionContextTree
}

// appendNode adds a new node to the resolution context.
func (t *resolutionContextTree) appendNode(importReference v1.ImportReference) *resolutionContextTree {
	newNode := &resolutionContextTree{
		importReference: importReference,
		parentNode:      t,
	}
	return newNode
}

// hasCycle checks if the current resolutionContextTree has a cycle
func (t *resolutionContextTree) hasCycle() error {
	var seenRefs []v1.ImportReference
	currNode := t
	for currNode.parentNode != nil {
		for _, seenRef := range seenRefs {
			if reflect.DeepEqual(seenRef, currNode.importReference) {
				return fmt.Errorf("devfile has an cycle in references: %v", formatImportCycle(t))
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
	cycle := resolveImportReference(end.importReference)
	for end.parentNode != nil {
		end = end.parentNode
		cycle = fmt.Sprintf("%s -> %s", resolveImportReference(end.importReference), cycle)
	}
	return cycle
}

func resolveImportReference(importReference v1.ImportReference) string {
	if !reflect.DeepEqual(importReference, v1.ImportReference{}) {
		switch {
		case importReference.Uri != "":
			return fmt.Sprintf("uri: %s", importReference.Uri)
		case importReference.Id != "":
			return fmt.Sprintf("id: %s, registryURL: %s", importReference.Id, importReference.RegistryUrl)
		case importReference.Kubernetes != nil:
			return fmt.Sprintf("name: %s, namespace: %s", importReference.Kubernetes.Name, importReference.Kubernetes.Namespace)
		}

	}
	// the first node
	return "main devfile"
}
