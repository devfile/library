package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	"github.com/devfile/parser/pkg/devfile/parser/data"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// DevfileObj is the runtime devfile object
type DevfileObj struct {

	// Ctx has devfile context info
	Ctx devfileCtx.DevfileCtx

	// Data has the devfile data
	Data data.DevfileData
}

// OverrideComponents overrides the components of the parent devfile
// overridePatch contains the patches to be applied to the parent's components
func (d DevfileObj) OverrideComponents(overridePatch []v1.ComponentParentOverride) error {
	for _, patchComponent := range overridePatch {
		found := false
		for _, originalComponent := range d.Data.GetComponents() {
			if strings.ToLower(patchComponent.Name) == originalComponent.Name {
				found = true

				var updatedComponent v1.Container

				merged, err := handleMerge(originalComponent.Container.Container, patchComponent.Container.ContainerParentOverride, v1.Container{})
				if err != nil {
					return err
				}
				err = json.Unmarshal(merged, &updatedComponent)
				if err != nil {
					return err
				}

				d.Data.UpdateComponent(v1.Component{
					ComponentUnion: v1.ComponentUnion{
						Container: &v1.ContainerComponent{
							Container: updatedComponent,
						},
					},
				})
			}
		}
		if !found {
			return fmt.Errorf("the component to override is not found in the parent")
		}
	}
	return nil
}

// OverrideCommands overrides the commands of the parent devfile
// overridePatch contains the patches to be applied to the parent's commands
func (d DevfileObj) OverrideCommands(overridePatch []v1.CommandParentOverride) error {
	for _, patchCommand := range overridePatch {
		found := false
		for _, originalCommand := range d.Data.GetCommands() {
			if strings.ToLower(patchCommand.Id) == originalCommand.Id {
				found = true
				var updatedCommand v1.ExecCommand

				merged, err := handleMerge(originalCommand.Exec, patchCommand.Exec, v1.ExecCommand{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedCommand)
				if err != nil {
					return err
				}

				d.Data.UpdateCommand(v1.Command{
					CommandUnion: v1.CommandUnion{
						Exec: &updatedCommand,
					},
				})
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

// OverrideEvents overrides the events of the parent devfile
// overridePatch contains the patches to be applied to the parent's events
func (d DevfileObj) OverrideEvents(overridePatch v1.Events) error {
	var updatedEvents v1.Events

	merged, err := handleMerge(d.Data.GetEvents(), overridePatch, v1.Events{})
	if err != nil {
		return err
	}

	err = json.Unmarshal(merged, &updatedEvents)
	if err != nil {
		return err
	}

	d.Data.UpdateEvents(updatedEvents.PostStart,
		updatedEvents.PostStop,
		updatedEvents.PreStart,
		updatedEvents.PreStop)
	return nil
}

// OverrideProjects overrides the projects of the parent devfile
// overridePatch contains the patches to be applied to the parent's projects
func (d DevfileObj) OverrideProjects(overridePatch []v1.ProjectParentOverride) error {
	for _, patchProject := range overridePatch {
		found := false
		for _, originalProject := range d.Data.GetProjects() {
			if strings.ToLower(patchProject.Name) == originalProject.Name {
				found = true
				var updatedProject v1.Project

				merged, err := handleMerge(originalProject, patchProject, v1.Project{})
				if err != nil {
					return err
				}

				err = json.Unmarshal(merged, &updatedProject)
				if err != nil {
					return err
				}

				d.Data.UpdateProject(updatedProject)
			}
		}
		if !found {
			return fmt.Errorf("the command to override is not found in the parent")
		}
	}
	return nil
}

// handleMerge merges the patch to the original data
// dataStruct is the type of the original and the patch data
func handleMerge(original, patch, dataStruct interface{}) ([]byte, error) {
	if reflect.TypeOf(original) != reflect.TypeOf(patch) {
		return nil, fmt.Errorf("type of original and patch doesn't match")
	}

	originalJson, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	patchJson, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	merged, err := strategicpatch.StrategicMergePatch(originalJson, patchJson, dataStruct)
	if err != nil {
		return nil, err
	}
	return merged, nil
}
