package validate

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/validation"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"strings"
)

// ValidateDevfileData validates whether sections of devfile are compatible
func ValidateDevfileData(data devfileData.DevfileData) error {

	commands, err := data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return err
	}
	components, err := data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	projects, err := data.GetProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}
	starterProjects, err := data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}

	var errstrings []string
	// validate components
	err = validation.ValidateComponents(components)
	if err != nil {
		errstrings = append(errstrings, err.Error())
	}

	// validate commands
	err = validation.ValidateCommands(commands, components)
	if err != nil {
		errstrings = append(errstrings, err.Error())
	}

	err = validation.ValidateEvents(data.GetEvents(), commands)
	if err != nil {
		errstrings = append(errstrings, err.Error())
	}

	err = validation.ValidateProjects(projects)
	if err != nil {
		errstrings = append(errstrings, err.Error())
	}

	err = validation.ValidateStarterProjects(starterProjects)
	if err != nil {
		errstrings = append(errstrings, err.Error())
	}

	if len(errstrings) > 0 {
		return fmt.Errorf(strings.Join(errstrings, "\n"))
	} else {
		return nil
	}

}
