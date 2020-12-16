package tests

import (
	"testing"
)

func TestExecCommand(t *testing.T) {
	if GetRandomNumber(10) > 10 {
		t.Error("Randon mumber was bigger than 10")
	}
}
