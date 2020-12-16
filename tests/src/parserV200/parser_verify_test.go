package parserV200

import "testing"

func TestGetRandomNumber(t *testing.T) {
	randomNum := GetRandomNumber(10)
	if randomNum > 10 {
		t.Errorf("Random mumber was bigger than 10 : %d", randomNum)
	}
}
