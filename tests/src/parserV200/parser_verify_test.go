package parserV200

import "testing"

func TestGetRandomNumber(t *testing.T) {
	if GetRandomNumber(10) > 10 {
		t.Error("Random mumber was bigger than 10")
	}
}
