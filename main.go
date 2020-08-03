package main

import (
	"fmt"

	"github.com/devfile/parser/pkg/devfile/parser"
)

func main() {
	devfile, err := ParseDevfile("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, component := range devfile.Data.GetAliasedComponents() {
			if component.Dockerfile != nil {
				fmt.Println(component.Dockerfile.DockerfileLocation)
			}
			if component.Container != nil {
				fmt.Println(component.Container.Image)
			}
		}

		for _, command := range devfile.Data.GetCommands() {
			if command.Exec != nil {
				fmt.Println(command.Exec.Group.Kind)
			}
		}
	}

}

//ParseDevfile to parse devfile from library
func ParseDevfile(devfileLocation string) (devfileoj parser.DevfileObj, err error) {

	var devfile parser.DevfileObj
	devfile, err = parser.ParseAndValidate(devfileLocation)
	return devfile, err
}
