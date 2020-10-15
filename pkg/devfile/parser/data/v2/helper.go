package v2

import (
	"fmt"

	v200 "github.com/devfile/parser/pkg/devfile/parser/data/v2/2.0.0"
	v210 "github.com/devfile/parser/pkg/devfile/parser/data/v2/2.1.0"
)

// NewDevfileDataV2 returns relevant devfile struct for the provided API version
func NewDevfileDataV2(version string) (obj DevfileDataV2, err error) {
	if version == "2.0.0" {
		return &v200.Devfile200{}, nil
	} else if version == "2.1.0" {
		return &v210.Devfile210{}, nil
	}

	return obj, fmt.Errorf("error on NewDevfileDataV2")
}
