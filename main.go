package main

import (
	"fmt"

	"github.com/devfile/parser/pkg/devfile/parser"
	devfileParser "github.com/devfile/parser/pkg/devfile/parser"
)

func main() {
	devfile, err := ParseDevfile("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, component := range devfile.Data.GetAliasedComponents() {
			if component.Dockerfile != nil {
				fmt.Println(component.Dockerfile.DockerfileLocation)
				break
			}
		}
	}

}

//ParseDevfile to parse devfile from library
func ParseDevfile(devfileLocation string) (devfileoj parser.DevfileObj, err error) {

	var devfile devfileParser.DevfileObj
	devfile, err = devfileParser.ParseAndValidate(devfileLocation)
	return devfile, err
}
