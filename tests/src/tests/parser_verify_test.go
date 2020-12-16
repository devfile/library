package tests

import (
	"testing"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

func TestExecCommand(t *testing.T) {
	testContent := TestContent{}
	testContent.CommandTypes = []schema.CommandType{schema.ExecCommandType}
	testContent.CreateWithParser = false
	testContent.EditContent = false
	testContent.FileName = GetDevFileName()
	err := RunTest(testContent)
	if err != nil {
		t.Fatalf("ERROR : %v", err)
	}
}
