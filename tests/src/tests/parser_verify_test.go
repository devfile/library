package tests

import (
	"fmt"
	"testing"
)

func Test_ExecCommand(t *testing.T) {
	FileName := "devfile.yaml"
	LogMessage(fmt.Sprintf("Start test for %s", FileName))
	testDevfile := GetDevfile(FileName)

	err := testDevfile.Verify()
	if err != nil {
		LogMessage(fmt.Sprintf("End test for %s", FileName))
	}
}
