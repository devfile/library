//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	devfilepkg "github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	v2 "github.com/devfile/library/v2/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "updateSchema" {
		ReplaceSchemaFile()
	} else {
		parserTest()
	}
}

func parserTest() {
	var args parser.ParserArgs
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "http") {
			args = parser.ParserArgs{
				URL: os.Args[1],
			}
		} else {
			args = parser.ParserArgs{
				Path: os.Args[1],
			}
		}
		fmt.Println("parsing devfile from " + os.Args[1])

	} else {
		args = parser.ParserArgs{
			Path: "devfile.yaml",
		}
		fmt.Println("parsing devfile from ./devfile.yaml")
	}
	devfile, warning, err := devfilepkg.ParseDevfileAndValidate(args)
	if err != nil {
		fmt.Println(err)
	} else {
		if len(warning.Commands) > 0 || len(warning.Components) > 0 || len(warning.Projects) > 0 || len(warning.StarterProjects) > 0 {
			fmt.Printf("top-level variables were not substituted successfully %+v\n", warning)
		}
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
				fmt.Printf("command %s is with kind: %s\n", command.Id, command.Exec.Group.Kind)
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
