package tests

import (
	"fmt"
)

func ExecCommand() {
	FileName := "devfile.yaml"
	LogMessage(fmt.Sprintf("Start test for %s", FileName))
	testDevfile := GetDevfile(FileName)

	err := testDevfile.Verify()
	if err != nil {
		LogMessage(fmt.Sprintf("End test for %s", FileName))
	}
}
