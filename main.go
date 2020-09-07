package main

import (
	"fmt"
	"reflect"

	"github.com/devfile/parser/pkg/devfile/parser"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
)

func main() {
	devfile, err := ParseDevfile("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		devdata := devfile.Data
		if (reflect.TypeOf(devdata) == reflect.TypeOf(&v200.Devfile200{})) {
			d := devdata.(*v200.Devfile200)
			fmt.Println(d.SchemaVersion)
		}

		for _, component := range devfile.Data.GetComponents() {
			/*
				if component.Dockerfile != nil {
							fmt.Println(component.Dockerfile.DockerfileLocation)
						}
			*/
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
	/*
		var devfile parser.DevfileObj
		devfile, err = parser.ParseAndValidate(devfileLocation)
		return devfile, err
	*/
	return
}
