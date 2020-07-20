package main

import (
	"fmt"

	parser "github.com/redhat-developer/devfile-parser/pkg/devfile/parser"
)

type DevfileObject struct {
	devfileObj parser.DevfileObj
}

func main() {
	d, _ := ParseDevfile("devfile.yaml")

	for _, comp := range d.Data.GetAliasedComponents() {
		if comp.Dockerfile != nil {
			fmt.Println(comp.Dockerfile.DockerfileLocation)
		}
	}
}

func newDevfileObject() *DevfileObject {
	return &DevfileObject{}
}

func ParseDevfile(devfileLocation string) (devfileoj parser.DevfileObj, err error) {

	devfileObject := newDevfileObject()
	devfileObject.devfileObj, err = parser.ParseAndValidate(devfileLocation)
	return devfileObject.devfileObj, err
}
