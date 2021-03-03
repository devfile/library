package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	devfilepkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "updateSchema" {
		ReplaceSchemaFile()
	} else {
		parserTest()
	}
}

//ParseDevfile to parse devfile from library
func ParseDevfile(devfileLocation string) (parser.DevfileObj, error) {

	devfile, err := devfilepkg.ParseAndValidate(devfileLocation)
	return devfile, err
}

func parserTest() {
	var devfile parser.DevfileObj
	var err error
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "http") {
			devfile, err = devfilepkg.ParseFromURLAndValidate(os.Args[1])
		} else {
			devfile, err = ParseDevfile(os.Args[1])
		}
		fmt.Println("parsing devfile from " + os.Args[1])

	} else {
		devfile, err = ParseDevfile("devfile.yaml")
		fmt.Println("parsing devfile from " + devfile.Ctx.GetAbsPath())
	}
	if err != nil {
		fmt.Println(err)
	} else {
		devdata := devfile.Data
		if (reflect.TypeOf(devdata) == reflect.TypeOf(&v2.DevfileV2{})) {
			d := devdata.(*v2.DevfileV2)
			fmt.Printf("schema version: %s\n", d.SchemaVersion)
		}

		components, e := devfile.Data.GetComponents(common.DevfileOptions{})
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}
		fmt.Printf("All component: \n")
		for _, component := range components {
			fmt.Printf("%s\n", component.Name)
		}

		fmt.Printf("All Exec commands: \n")
		commands, e := devfile.Data.GetCommands(common.DevfileOptions{})
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}
		for _, command := range commands {
			if command.Exec != nil {
				fmt.Printf("command %s is with kind: %s", command.Id, command.Exec.Group.Kind)
				fmt.Printf("workingDir is: %s\n", command.Exec.WorkingDir)
			}
		}

		fmt.Println("=========================================================")

		compOptions := common.DevfileOptions{
			Filter: map[string]interface{}{
				"tool": "console-import",
				"import": map[string]interface{}{
					"strategy": "Dockerfile",
				},
			},
		}

		components, e = devfile.Data.GetComponents(compOptions)
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}
		fmt.Printf("Container components applied filter: \n")
		for _, component := range components {
			if component.Container != nil {
				fmt.Printf("%s\n", component.Name)
			}
		}

		cmdOptions := common.DevfileOptions{
			Filter: map[string]interface{}{
				"tool": "odo",
			},
		}

		fmt.Printf("Exec commands applied filter: \n")
		commands, e = devfile.Data.GetCommands(cmdOptions)
		if e != nil {
			fmt.Printf("err: %v\n", err)
		}
		for _, command := range commands {
			if command.Exec != nil {
				fmt.Printf("command %s is with kind: %s", command.Id, command.Exec.Group.Kind)
				fmt.Printf("workingDir is: %s\n", command.Exec.WorkingDir)
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
