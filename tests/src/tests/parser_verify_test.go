package tests

import (
	"fmt"
	"testing"
)

func Test_ExecCommand(t *testing.T) {
	runTest("devFile", t)
}

func runTest(FileName string, t *testing.T) {

	LogMessage(fmt.Sprintf("Start test for %s", FileName))
	testDevfile := GetDevfile(FileName)

	err := testDevfile.Verify()
	if err != nil {
		LogMessage(fmt.Sprintf("End test for %s", FileName))
	}

}
