package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	testDir       = "../"
	jsonDir       = "../json/"
	schemaVersion = "2.0.0"
)

// TestToRun contains details of an individual test
type TestToRun struct {
	SchemaVersion string
	FileDirectory string
	TestError     []string
	FileName      string   `json:"FileName"`
	Disabled      bool     `json:"Disabled"`
	ExpectOutcome string   `json:"ExpectOutcome"`
	Files         []string `json:"Files"`
}

type TestJsonFile struct {
	FileInfo      os.FileInfo
	TempDirectory string
	SchemaVersion string      `json:"SchemaVersion"`
	Tests         []TestToRun `json:"Tests"`
}

// GetJsonFile returns an array of TestJsonFile objects, one for each json file containing tests
func GetJsonFiles(directory string) ([]TestJsonFile, error) {

	var jsonFiles []TestJsonFile

	// Read the content of the json directory to find test files
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		LogErrorMessage(fmt.Sprintf("Error finding test json files in : %s :  %v", jsonDir, err))
	} else {
		for _, testJsonFile := range files {
			// if the file ends with -test.json it can be processed
			if strings.HasSuffix(testJsonFile.Name(), "-tests.json") {
				testJson := TestJsonFile{}
				testJson.FileInfo = testJsonFile
				dirName := testJsonFile.Name()[0:strings.LastIndex(testJsonFile.Name(), ".json")]
				testJson.TempDirectory = CreateTempDir(dirName)
				jsonFiles = append(jsonFiles, testJson)
			}
		}
	}
	return jsonFiles, err
}

// GetTests returns an array of TestToRun objects, one for each test contained in a json file
func (testJsonFile *TestJsonFile) GetTests() ([]TestToRun, error) {

	var err error
	if len(testJsonFile.Tests) < 1 {
		// Open the json file which defines the tests to run
		var testJson *os.File
		testJson, err = os.Open(filepath.Join(jsonDir, testJsonFile.FileInfo.Name()))
		if err != nil {
			LogErrorMessage(fmt.Sprintf("Failed to open %s : %s", testJsonFile.FileInfo.Name(), err))
		} else {
			// Read contents of the json file which defines the tests to run
			var byteValue []byte
			byteValue, err = ioutil.ReadAll(testJson)
			if err != nil {
				LogErrorMessage(fmt.Sprintf("Failed to read : %s : %v", testJsonFile.FileInfo.Name(), err))
			} else {
				// Unmarshall the contents of the json file which defines the tests to run for each test
				err = json.Unmarshal(byteValue, &testJsonFile)
				if err != nil {
					LogErrorMessage(fmt.Sprintf("Failed to unmarshal : %s : %v", testJsonFile.FileInfo.Name(), err))
				}
			}
		}
		testJson.Close()

		for testNum, _ := range testJsonFile.Tests {
			(&testJsonFile.Tests[testNum]).FileDirectory = testJsonFile.TempDirectory
			(&testJsonFile.Tests[testNum]).SchemaVersion = testJsonFile.SchemaVersion
		}
	}
	return testJsonFile.Tests, err
}

// CreatTestYaml creates a devfile.yaml file on disk as required by a test
func (testToRun *TestToRun) CreateTestYaml() (string, error) {

	yamlFileName := filepath.Join(testToRun.FileDirectory, testToRun.FileName)
	// Open the file to contain the generated test yaml'

	yamlFile, err := os.OpenFile(yamlFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		LogErrorMessage(fmt.Sprintf("FAIL : Failed to open %s : %v", yamlFileName, err))
	} else {

		yamlFile.WriteString("schemaVersion: \"" + testToRun.SchemaVersion + "\"\n")

		// Now add each of the yaml sippets used the make the yaml file for test
		for _, testSnippet := range testToRun.Files {
			// Read the snippet
			data, readErr := ioutil.ReadFile(filepath.Join(testDir, testSnippet))
			if readErr != nil {
				err = readErr
				LogErrorMessage(fmt.Sprintf("FAIL: failed reading %s: %v", filepath.Join(testDir, testSnippet), err))
				break
			} else {
				// Add snippet to yaml file
				yamlFile.Write(data)
				// Ensure approproate line breaks
				yamlFile.WriteString("\n")
			}
		}
		yamlFile.Close()
	}
	return yamlFileName, err
}

// GetAlTests returns all tests from all json files in the specified directory
func GetAllTests(directory string) ([]TestToRun, bool) {

	var testsToRun []TestToRun
	errorOccurred := false
	testJsonFiles, err := GetJsonFiles(directory)
	if err != nil {
		errorOccurred = true
	}

	for _, testJsonFile := range testJsonFiles {
		jsonFileTests, err := testJsonFile.GetTests()
		if err != nil {
			errorOccurred = true
		}
		for _, jsonFileTest := range jsonFileTests {
			testsToRun = append(testsToRun, jsonFileTest)
		}
	}

	return testsToRun, errorOccurred
}
