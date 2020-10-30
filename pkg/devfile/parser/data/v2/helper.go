package v2

import (
	v200 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.0.0"
	v210 "github.com/devfile/library/pkg/devfile/parser/data/v2/2.1.0"
)

// New returns relevant devfile ver 2.x.x struct for the provided API version
func New(version string) (obj DevfileDataV2) {

	switch version {
	case "2.0.0":
		return &v200.Devfile200{}
	case "2.1.0":
		return &v210.Devfile210{}
	}

	return obj
}
