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
	"strings"
)

func ReplaceSchemaFile() {
	if len(os.Args) != 7 {
		printErr(fmt.Errorf("ReplaceSchemaFile() expect 7 arguments"))
		os.Exit(1)
	}
	originalSchema := os.Args[2]
	schemaURL := os.Args[3]
	packageVersion := os.Args[4]
	jsonSchemaVersion := os.Args[5]
	filePath := os.Args[6]

	// replace all ` with ' to convert schema content from json file format to json format in golang
	newSchema := strings.ReplaceAll(originalSchema, "`", "'")
	fmt.Printf("Writing to file: %s\n", filePath)
	fileContent := fmt.Sprintf("package %s\n\n// %s\nconst %s = `%s\n`\n", packageVersion, schemaURL, jsonSchemaVersion, newSchema)

	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		printErr(err)
		os.Exit(1)
	}
}

func printErr(err error) {
	// prints error in red
	colorRed := "\033[31m"
	colorReset := "\033[0m"

	fmt.Println(string(colorRed), err, string(colorReset))
}
