package version210

import (
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
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

func getTestDevfileData() (testDevfile Devfile210, commands []v1.Command) {

	command := "ls -la"
	component := "alias1"
	debugCommand := "nodemon --inspect={DEBUG_PORT}"
	debugComponent := "alias2"
	workDir := "/root"

	execCommands := []v1.Command{
		{
			Exec: &v1.ExecCommand{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
		},
		{
			Exec: &v1.ExecCommand{
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
