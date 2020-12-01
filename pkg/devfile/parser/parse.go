package parser

import (
	"encoding/json"

	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"k8s.io/klog"

	"reflect"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	apiOverride "github.com/devfile/api/pkg/utils/overriding"
	"github.com/pkg/errors"
)

// ParseDevfile func validates the devfile integrity.
// Creates devfile context and runtime objects
func parseDevfile(d DevfileObj) (DevfileObj, error) {

	// Validate devfile
	err := d.Ctx.Validate()
	if err != nil {
		return d, err
	}

	// Create a new devfile data object
	d.Data, err = data.NewDevfileData(d.Ctx.GetApiVersion())
	if err != nil {
		return d, err
	}

	// Unmarshal devfile content into devfile struct
	err = json.Unmarshal(d.Ctx.GetDevfileContent(), &d.Data)
	if err != nil {
		return d, errors.Wrapf(err, "failed to decode devfile content")
	}

	if d.Data.GetParent() != nil {
		if !reflect.DeepEqual(d.Data.GetParent(), &v1.Parent{}) && d.Data.GetParent().Uri != "" {
			err = parseParent(d)
			if err != nil {
				return DevfileObj{}, err
			}
		}
	}

	err = parsePlugin(d)
	if err != nil {
		return DevfileObj{}, err
	}

	// Successful
	return d, nil
}

// Parse func populates the devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d)
}

// ParseFromURL func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromURL(url string) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.NewURLDevfileCtx(url)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.PopulateFromURL()
	if err != nil {
		return d, err
	}
	return parseDevfile(d)
}

// ParseFromData func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromData(data []byte) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.DevfileCtx{}
	err = d.Ctx.SetDevfileContentFromBytes(data)
	if err != nil {
		return d, errors.Wrap(err, "failed to set devfile content from bytes")
	}
	err = d.Ctx.PopulateFromRaw()
	if err != nil {
		return d, err
	}

	return parseDevfile(d)
}

func parseParent(d DevfileObj) error {
	parent := d.Data.GetParent()

	parentData, err := ParseFromURL(parent.Uri)
	if err != nil {
		return err
	}

	parentWorkspaceContent := parentData.Data.GetDevfileWorkspace()
	result, err := apiOverride.OverrideDevWorkspaceTemplateSpec(parentWorkspaceContent, parent)
	if err != nil {
		return err
	}

	parentData.Data.SetDevfileWorkspace(*result)

	klog.V(4).Infof("adding data of devfile with URI: %v", parent.Uri)

	// since the parent's data has been overriden
	// add the items back to the current devfile
	// error indicates that the item has been defined again in the current devfile
	commandsMap := parentData.Data.GetCommands()
	commands := make([]v1.Command, 0, len(commandsMap))
	for _, command := range commandsMap {
		commands = append(commands, command)
	}
	err = d.Data.AddCommands(commands...)
	if err != nil {
		return errors.Wrapf(err, "error while adding commands from the parent devfiles")
	}

	err = d.Data.AddComponents(parentData.Data.GetComponents())
	if err != nil {
		return errors.Wrapf(err, "error while adding components from the parent devfiles")
	}

	err = d.Data.AddProjects(parentData.Data.GetProjects())
	if err != nil {
		return errors.Wrapf(err, "error while adding projects from the parent devfiles")
	}

	err = d.Data.AddStarterProjects(parentData.Data.GetStarterProjects())
	if err != nil {
		return errors.Wrapf(err, "error while adding starter projects from the parent devfiles")
	}

	err = d.Data.AddEvents(parentData.Data.GetEvents())
	if err != nil {
		return errors.Wrapf(err, "error while adding events from the parent devfiles")
	}
	return nil
}

func parsePlugin(d DevfileObj) (err error) {

	for _, component := range d.Data.GetComponents() {
		if component.Plugin != nil && !reflect.DeepEqual(component.Plugin, &v1.PluginComponent{}) {
			plugin := component.Plugin
			var pluginData DevfileObj
			if plugin.Uri != "" {
				pluginData, err = ParseFromURL(plugin.Uri)
				if err != nil {
					return err
				}
			}
			pluginWorkspaceContent := pluginData.Data.GetDevfileWorkspace()
			result, err := apiOverride.OverrideDevWorkspaceTemplateSpec(pluginWorkspaceContent, plugin)
			if err != nil {
				return err
			}
			pluginData.Data.SetDevfileWorkspace(*result)
			klog.V(4).Infof("adding data of devfile with URI: %v", plugin.Uri)
			// since the plugin's data has been overriden
			// add the items back to the current devfile
			// error indicates that the item has been defined again in the current devfile
			commandsMap := pluginData.Data.GetCommands()
			commands := make([]v1.Command, 0, len(commandsMap))
			for _, command := range commandsMap {
				commands = append(commands, command)
			}

			// plugin component contributes components, commands and events
			err = d.Data.AddCommands(commands...)
			if err != nil {
				return errors.Wrapf(err, "error while adding commands from the plugin devfiles")
			}

			err = d.Data.AddComponents(pluginData.Data.GetComponents())
			if err != nil {
				return errors.Wrapf(err, "error while adding components from the plugin devfiles")
			}

			err = d.Data.AddEvents(pluginData.Data.GetEvents())
			if err != nil {
				return errors.Wrapf(err, "error while adding events from the plugin devfiles")
			}
			return nil
		}
	}
	return nil

}
