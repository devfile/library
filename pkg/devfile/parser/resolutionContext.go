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

// addPlugin adds a plugin component to the resolution context.
func (t *resolutionContextTree) addPlugin(name string, importReference v1.ImportReference) *resolutionContextTree {
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
				return fmt.Errorf("devfile has an cycle in references: %v", currNode.importReference)
			}
		}
		seenRefs = append(seenRefs, currNode.importReference)
		currNode = currNode.parentNode
	}
	return nil
}

//// formatImportCycle is a utility method for formatting a cycle that has been detected. Output is formatted as
//// plugin1 -> plugin2 -> plugin3 -> plugin1, where pluginX are component names.
//func formatImportCycle(end *resolutionContextTree) string {
//	cycle := fmt.Sprintf("%v", end.importReference)
//	for end.parentNode != nil {
//		end = end.parentNode
//		if end.parentNode == nil {
//			end.componentName = "devfile"
//		}
//		cycle = fmt.Sprintf("%s -> %s", end.componentName, cycle)
//	}
//	return cycle
//}