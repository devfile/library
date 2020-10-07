package common

import (
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// GetID returns the ID of the command
func GetID(dc v1.Command) string {
	if dc.Composite != nil || dc.Exec != nil {
		return dc.Id
	}

	return ""
}

// SetIDToLower converts the command's id to lower case for more efficient processing and returns the new id
func SetIDToLower(dc *v1.Command) string {
	var newId string
	if dc.Exec != nil || dc.Composite != nil {
		newId = strings.ToLower(dc.Id)
		dc.Id = newId
	}
	return newId
}

// GetGroup returns the group the command belongs to
func GetGroup(dc v1.Command) *v1.CommandGroup {
	if dc.Composite != nil {
		return dc.Composite.Group
	} else if dc.Exec != nil {
		return dc.Exec.Group
	}

	return nil
}

// GetExecComponent returns the component of the exec command
func GetExecComponent(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.Component
	}

	return ""
}

// GetExecCommandLine returns the command line of the exec command
func GetExecCommandLine(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.CommandLine
	}

	return ""
}

// GetExecWorkingDir returns the working dir of the exec command
func GetExecWorkingDir(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.WorkingDir
	}

	return ""
}

// IsComposite checks if the command is a composite command
func IsComposite(dc v1.Command) bool {
	return dc.Composite != nil
}
