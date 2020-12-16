package tests

import (
	"testing"
)

func Test_execCommand(t *testing.T) {
	if GetRandomNumber(10) > 10 {
		t.Error("Randon mumber was bigger than 10")
	}
}
