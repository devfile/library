package main

import (
	"fmt"
	"reflect"

	devfilepkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

func main() {
	devfile, err := ParseDevfile("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		devdata := devfile.Data
		if (reflect.TypeOf(devdata) == reflect.TypeOf(&v2.DevfileV2{})) {
			d := devdata.(*v2.DevfileV2)
			fmt.Printf("schema version: %s\n", d.SchemaVersion)
		}

		compOptions := common.DevfileOptions{
			Filter: map[string]interface{}{
				"tool": "console-import",
				"import": map[string]interface{}{
					"strategy": "Dockerfile",
				},
			},
		}

		components, e := devfile.Data.GetComponents(compOptions)
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}

		for _, component := range components {
			if component.Container != nil {
				fmt.Printf("component container: %s\n", component.Name)
			}
		}

		cmdOptions := common.DevfileOptions{
			Filter: map[string]interface{}{
				"tool": "odo",
			},
		}

		commands, e := devfile.Data.GetCommands(cmdOptions)
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}
		for _, command := range commands {
			if command.Exec != nil {
				fmt.Printf("exec command kind: %s\n", command.Exec.Group.Kind)
			}
		}

		var err error
		metadataAttr := devfile.Data.GetMetadata().Attributes
		dockerfilePath := metadataAttr.GetString("alpha.build-dockerfile", &err)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		fmt.Printf("dockerfilePath: %s\n", dockerfilePath)
	}

}

//ParseDevfile to parse devfile from library
func ParseDevfile(devfileLocation string) (parser.DevfileObj, error) {

	devfile, err := devfilepkg.ParseAndValidate(devfileLocation)
	return devfile, err
}
