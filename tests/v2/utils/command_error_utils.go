package utils

import (
	"fmt"
	"strings"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

type ErrorDefinition interface {
	checkError(command *schema.Command) (bool,string)
}



type CommandErrorDefinition interface {
	addCommandError(command *schema.Command)
	checkError(errorMessage string) (bool,string)
}

type CommandNoId struct {
	errorString string
}

func (c *CommandNoId) addCommandError(command *schema.Command) {
	command.Id = ""
}

func (c *CommandNoId) checkError(errorMessage string) (bool,string) {
	if !strings.Contains(errorMessage,c.errorString) {
		return false,fmt.Sprintf("Error message Expected: %s, Received: %s ",c.errorString, errorMessage)
	}
	return true,""
}
