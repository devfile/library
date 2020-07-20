package main

import (
	"fmt"

	devfileParser "github.com/redhat-developer/devfile-parser/pkg/devfile/parser"
)

func main() {

	var devfile devfileParser.DevfileObj
	devfile, err := devfileParser.ParseAndValidate("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, component := range devfile.Data.GetAliasedComponents() {
			if component.Dockerfile != nil {
				fmt.Println(component.Dockerfile.Destination)
			}
		}
	}

}
