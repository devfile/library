package main

import (
	"fmt"
	"io/ioutil"
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

	if err := ioutil.WriteFile(filePath, []byte(fileContent), 0755); err != nil {
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
