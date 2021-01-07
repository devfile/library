package utils

import (
	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

type GenericProject struct {
	Name    string
	Project *schema.Project
}
