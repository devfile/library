package version210

import (
	"testing"

	apiComp "github.com/devfile/kubernetes-api/pkg/apis/workspaces/v1alpha1"
	common "github.com/devfile/parser/pkg/devfile/parser/data/common"
)

func TestGetCommands(t *testing.T) {

	testDevfile, execCommands := getTestDevfileData()

	got := testDevfile.GetCommands()
	want := execCommands

	for i, command := range got {
		if command.Exec != want[i].Exec {
			t.Error("Commands returned don't match expected commands")
		}
	}

}

func getTestDevfileData() (testDevfile Devfile210, commands []common.DevfileCommand) {

	command := "ls -la"
	component := "alias1"
	debugCommand := "nodemon --inspect={DEBUG_PORT}"
	debugComponent := "alias2"
	workDir := "/root"

	execCommands := []common.DevfileCommand{
		{
			Exec: &apiComp.ExecCommand{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
		},
		{
			Exec: &apiComp.ExecCommand{
				CommandLine: debugCommand,
				Component:   debugComponent,
				WorkingDir:  workDir,
			},
		},
	}

	testDevfileobj := Devfile210{
		Commands: execCommands,
	}

	return testDevfileobj, execCommands
}
