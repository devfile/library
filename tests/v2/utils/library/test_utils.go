//
// Copyright Red Hat
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

package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	header "github.com/devfile/api/v2/pkg/devfile"
	commonUtils "github.com/devfile/api/v2/test/v200/utils/common"
	devfilepkg "github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	devfileData "github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/util"
)

const (
	// numDevfiles : the number of devfiles to create for each test
	numDevfiles = 5
	// numThreads :  Number of threads used by multi-thread tests
	numThreads = 5
	// schemaVersion: Latest schemaVersion
	schemaVersion = "2.2.0"
)

// DevfileValidator struct for DevfileValidator interface.
// The DevfileValidator interface is test/v200/utils/common/test_utils.go of the devfile/api repository.
type DevfileValidator struct{}

var parserArgs = parser.ParserArgs{}

// directory where test devfiles are generated and/or copied
const destDir = "tmp/library_test/"

// WriteAndValidate implements DevfileValidator interface.
// writes to disk and validates the specified devfile
func (devfileValidator DevfileValidator) WriteAndValidate(devfile *commonUtils.TestDevfile) error {
	err := writeDevfile(devfile)
	if err != nil {
		commonUtils.LogErrorMessage(fmt.Sprintf("Error writing file : %s : %v", devfile.FileName, err))
	} else {
		err = validateDevfile(devfile)
		if err != nil {
			commonUtils.LogErrorMessage(fmt.Sprintf("Error vaidating file : %s : %v", devfile.FileName, err))
		} else {
			err = verify(devfile)
		}
	}
	return err
}

// DevfileFollower struct for DevfileFollower interface.
// The DevfileFollower interface is defined in test/v200/utils/common/test_utils.go of the devfile/api repository
type DevfileFollower struct {
	LibraryData devfileData.DevfileData
}

// AddCommand adds the specified command to the library data
func (devfileFollower DevfileFollower) AddCommand(command schema.Command) error {
	return devfileFollower.LibraryData.AddCommands([]schema.Command{command})
}

// UpdateCommand updates the specified command in the library data
func (devfileFollower DevfileFollower) UpdateCommand(command schema.Command) {
	devfileFollower.LibraryData.UpdateCommand(command)
}

// AddComponent adds the specified component to the library data
func (devfileFollower DevfileFollower) AddComponent(component schema.Component) error {
	var components []schema.Component
	components = append(components, component)
	return devfileFollower.LibraryData.AddComponents(components)
}

// UpdateComponent updates the specified component in the library data
func (devfileFollower DevfileFollower) UpdateComponent(component schema.Component) {
	devfileFollower.LibraryData.UpdateComponent(component)
}

// AddProject adds the specified project to the library data
func (devfileFollower DevfileFollower) AddProject(project schema.Project) error {
	var projects []schema.Project
	projects = append(projects, project)
	return devfileFollower.LibraryData.AddProjects(projects)
}

// UpdateProject updates the specified project in the library data
func (devfileFollower DevfileFollower) UpdateProject(project schema.Project) {
	devfileFollower.LibraryData.UpdateProject(project)
}

// AddStarterProject adds the specified starter project to the library data
func (devfileFollower DevfileFollower) AddStarterProject(starterProject schema.StarterProject) error {
	var starterProjects []schema.StarterProject
	starterProjects = append(starterProjects, starterProject)
	return devfileFollower.LibraryData.AddStarterProjects(starterProjects)
}

// UpdateStarterProject updates the specified starter project in the library data
func (devfileFollower DevfileFollower) UpdateStarterProject(starterProject schema.StarterProject) {
	devfileFollower.LibraryData.UpdateStarterProject(starterProject)
}

// AddEvent adds the specified event to the library data
func (devfileFollower DevfileFollower) AddEvent(event schema.Events) error {
	return devfileFollower.LibraryData.AddEvents(event)
}

// UpdateEvent updates the specified event in the library data
func (devfileFollower DevfileFollower) UpdateEvent(event schema.Events) {
	devfileFollower.LibraryData.UpdateEvents(event.PostStart, event.PostStop, event.PreStart, event.PreStop)
}

// SetParent sets the specified parent in the library data
func (devfileFollower DevfileFollower) SetParent(parent schema.Parent) error {
	devfileFollower.LibraryData.SetParent(&parent)
	return nil
}

// UpdateParent updates the specified parent in the library data
func (devfileFollower DevfileFollower) UpdateParent(parent schema.Parent) {
	devfileFollower.LibraryData.SetParent(&parent)
}

// SetMetaData sets the specified metaData in the library data
func (devfileFollower DevfileFollower) SetMetaData(metaData header.DevfileMetadata) error {
	devfileFollower.LibraryData.SetMetadata(metaData)
	return nil
}

// UpdateMetaData updates the specified UpdateMetaData in the library data
func (devfileFollower DevfileFollower) UpdateMetaData(updateMetaData header.DevfileMetadata) {
	devfileFollower.LibraryData.SetMetadata(updateMetaData)
}

// SetSchemaVersion  sets the specified schemaVersion in the library data
func (devfileFollower DevfileFollower) SetSchemaVersion(schemaVersion string) {
	devfileFollower.LibraryData.SetSchemaVersion(schemaVersion)
}

// WriteDevfile uses the library to create a devfile on disk for use in a test.
func writeDevfile(devfile *commonUtils.TestDevfile) error {
	var err error

	fileName := devfile.FileName
	if !strings.HasSuffix(fileName, ".yaml") {
		fileName += ".yaml"
	}

	commonUtils.LogInfoMessage(fmt.Sprintf("Use Parser to write devfile %s", fileName))

	ctx := devfileCtx.NewDevfileCtx(fileName)

	err = ctx.SetAbsPath()
	if err != nil {
		commonUtils.LogErrorMessage(fmt.Sprintf("Setting devfile path : %v", err))
	} else {
		devObj := parser.DevfileObj{
			Ctx:  ctx,
			Data: devfile.Follower.(DevfileFollower).LibraryData,
		}
		err = devObj.WriteYamlDevfile()
		if err != nil {
			commonUtils.LogErrorMessage(fmt.Sprintf("Writing devfile : %v", err))
		}
	}

	return err
}

// validateDevfile uses the library to parse and validate a devfile on disk
func validateDevfile(devfile *commonUtils.TestDevfile) error {
	var err error

	commonUtils.LogInfoMessage(fmt.Sprintf("Parse and Validate %s : ", devfile.FileName))
	parseK8sDefinitionFromURI := false
	parserArgs.Path = devfile.FileName
	parserArgs.ConvertKubernetesContentInUri = &parseK8sDefinitionFromURI
	libraryObj, warning, err := devfilepkg.ParseDevfileAndValidate(parserArgs)

	if len(warning.Commands) > 0 || len(warning.Components) > 0 || len(warning.Projects) > 0 || len(warning.StarterProjects) > 0 {
		commonUtils.LogWarningMessage(fmt.Sprintf("top-level variables were not substituted successfully %+v\n", warning))
	}

	if err != nil {
		commonUtils.LogErrorMessage(fmt.Sprintf("From ParseDevfileAndValidate %v : ", err))
		return err
	} else {
		follower := devfile.Follower.(DevfileFollower)
		follower.LibraryData = libraryObj.Data
	}

	err = verifyEphemeralUnset(libraryObj)
	if err != nil {
		return err
	}

	return err
}

// verifyEphemeralUnset  verifies volume.Ephemeral is not set on schema version 2.0.0
func verifyEphemeralUnset(libraryObj parser.DevfileObj) error {
	version := libraryObj.Data.GetSchemaVersion()

	//verify volume.Ephemeral is not set on schema version 2.0.0
	if version == string(devfileData.APISchemaVersion200) {
		volumes, err := libraryObj.Data.GetComponents(common.DevfileOptions{
			ComponentOptions: common.ComponentOptions{
				ComponentType: schema.VolumeComponentType,
			},
		})

		if err != nil {
			return err
		}

		for i := range volumes {
			volume := volumes[i].Volume
			if volume != nil && volume.Ephemeral != nil {
				return errors.New("ephemeral is not supported on schema version 2.0.0")
			}
		}
	}

	return nil
}

// RunMultiThreadTest : Runs the same test on multiple threads, the test is based on the content of the specified TestContent
func RunMultiThreadTest(testContent commonUtils.TestContent, t *testing.T) {

	commonUtils.LogMessage(fmt.Sprintf("Start Threaded test for %s", testContent.FileName))

	devfileName := testContent.FileName
	for i := 1; i < numThreads; i++ {
		testContent.FileName = commonUtils.AddSuffixToFileName(devfileName, "T"+strconv.Itoa(i)+"-")
		go RunTest(testContent, t)
	}

	commonUtils.LogMessage(fmt.Sprintf("Sleep 3 seconds to allow all threads to complete : %s", devfileName))
	time.Sleep(3 * time.Second)
	commonUtils.LogMessage(fmt.Sprintf("Sleep complete : %s", devfileName))

}

// RunMultiThreadedStaticTest : Runs the same test on multiple threads, the test is based on the content of the specified TestContent
func RunMultiThreadedStaticTest(testContent commonUtils.TestContent, t *testing.T) {

	commonUtils.LogMessage(fmt.Sprintf("Start Threaded static test for %s", testContent.FileName))
	devfileName := testContent.FileName
	for i := 1; i < numThreads+1; i++ {
		testContent.FileName = commonUtils.AddSuffixToFileName(devfileName, "T"+strconv.Itoa(i)+"-")
		duplicateDevfileSample(t, devfileName, testContent.FileName)
		go RunStaticTest(testContent, t)
	}

	commonUtils.LogMessage(fmt.Sprintf("Sleep 3 seconds to allow all threads to complete : %s", devfileName))
	time.Sleep(3 * time.Second)
	commonUtils.LogMessage(fmt.Sprintf("Sleep complete : %s", devfileName))

}

// SetParserArgs :  Used when parser args other than filename are set in the library tests
func SetParserArgs(args parser.ParserArgs) {
	parserArgs = args
}

// CopyDevfileSamples : Copies existing artifacts from the devfiles/samples directory to the tmp/library_test directory.  Used in parent tests
func CopyDevfileSamples(t *testing.T, testDevfiles []string) {

	srcDir := "../devfiles/samples/"
	dstDir := commonUtils.CreateTempDir("library_test")

	for i := range testDevfiles {
		srcPath := srcDir + testDevfiles[i]
		destPath := dstDir + testDevfiles[i]

		file, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("Error locating testDevfile %v ", err)))
		} else {
			commonUtils.LogMessage(fmt.Sprintf("copy file from %s to %s ", srcPath, destPath))
			util.CopyFile(srcPath, destPath, file)
		}
	}
}

// duplicateDevfileSample: Makes a copy of the parent devfile test artifact that is expected to exist in the tmp/library_test directory.
// This is used in the multi-threaded parent test scenarios
func duplicateDevfileSample(t *testing.T, src string, dst string) {
	srcPath := destDir + src
	destPath := destDir + dst
	file, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("Error locating testDevfile %v ", err)))
	} else {
		commonUtils.LogMessage(fmt.Sprintf("duplicate file %s to %s ", srcPath, destPath))
		util.CopyFile(srcPath, destDir+dst, file)
	}
}

// RunTest : Runs a test to create and verify a devfile based on the content of the specified TestContent
func RunTest(testContent commonUtils.TestContent, t *testing.T) {
	commonUtils.LogMessage(fmt.Sprintf("Start test for %s", testContent.FileName))
	devfileName := testContent.FileName
	for i := 1; i <= numDevfiles; i++ {

		testContent.FileName = commonUtils.AddSuffixToFileName(devfileName, strconv.Itoa(i))
		commonUtils.LogMessage(fmt.Sprintf("Start test for %s", testContent.FileName))

		validator := DevfileValidator{}
		follower := DevfileFollower{}
		libraryData, err := devfileData.NewDevfileData(schemaVersion)
		if err != nil {
			t.Fatalf(commonUtils.LogMessage(fmt.Sprintf("Error creating parser data : %v", err)))
		}
		libraryData.SetSchemaVersion(schemaVersion)
		follower.LibraryData = libraryData
		commonUtils.LogMessage(fmt.Sprintf("Parser data created with schema version : %s", follower.LibraryData.GetSchemaVersion()))

		testDevfile, err := commonUtils.GetDevfile(testContent.FileName, follower, validator)
		if err != nil {
			t.Fatalf(commonUtils.LogMessage(fmt.Sprintf("Error creating devfile : %v", err)))
		}

		testDevfile.RunTest(testContent, t)

		if testContent.EditContent {
			if len(testContent.CommandTypes) > 0 {
				err = editCommands(&testDevfile)
				if err != nil {
					t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("ERROR editing commands :  %s : %v", testContent.FileName, err)))
				}
			}
			if len(testContent.ComponentTypes) > 0 {
				err = editComponents(&testDevfile)
				if err != nil {
					t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("ERROR editing components :  %s : %v", testContent.FileName, err)))
				}
			}
			if len(testContent.ProjectTypes) > 0 {
				err = editProjects(&testDevfile)
				if err != nil {
					t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("ERROR editing projects :  %s : %v", testContent.FileName, err)))
				}
			}
			if len(testContent.StarterProjectTypes) > 0 {
				err = editStarterProjects(&testDevfile)
				if err != nil {
					t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("ERROR editing starter projects :  %s : %v", testContent.FileName, err)))
				}
			}

			validator.WriteAndValidate(&testDevfile)
		}
	}
}

// RunStaticTest : Runs fixed tests based on pre-existing artifacts
func RunStaticTest(testContent commonUtils.TestContent, t *testing.T) {
	commonUtils.LogMessage(fmt.Sprintf("Start test for %s", testContent.FileName))
	follower := DevfileFollower{}
	validator := DevfileValidator{}
	testfileName := destDir + testContent.FileName
	testDevfile, _ := commonUtils.GetDevfile(testfileName, follower, validator)
	testDevfile.SchemaDevFile.Parent = &schema.Parent{}
	err := validateDevfile(&testDevfile)
	if err != nil {
		t.Fatalf(commonUtils.LogErrorMessage(fmt.Sprintf("Error validating testDevfile %v ", err)))
	}
}

// verify verifies the library contents of the specified devfile with the expected content
func verify(devfile *commonUtils.TestDevfile) error {

	commonUtils.LogInfoMessage(fmt.Sprintf("Verify %s : ", devfile.FileName))

	var errorString []string

	libraryData := devfile.Follower.(DevfileFollower).LibraryData
	commonUtils.LogInfoMessage(fmt.Sprintf("Get commands %s : ", devfile.FileName))
	commands, err := libraryData.GetCommands(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting Commands from library : %s : %v", devfile.FileName, err)))
	} else {
		if commands != nil && len(commands) > 0 {
			err := VerifyCommands(devfile, commands)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verify Commands %s : %v", devfile.FileName, err)))
			}
		} else {
			commonUtils.LogInfoMessage(fmt.Sprintf("No commands found in %s : ", devfile.FileName))
		}
	}

	commonUtils.LogInfoMessage(fmt.Sprintf("Get components %s : ", devfile.FileName))
	components, err := libraryData.GetComponents(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting Components from library : %s : %v", devfile.FileName, err)))
	} else {
		if components != nil && len(components) > 0 {
			err := VerifyComponents(devfile, components)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verify Components %s : %v", devfile.FileName, err)))
			}
		} else {
			commonUtils.LogInfoMessage(fmt.Sprintf("No components found in %s : ", devfile.FileName))
		}
	}

	commonUtils.LogInfoMessage(fmt.Sprintf("Get projects %s : ", devfile.FileName))
	projects, err := libraryData.GetProjects(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting Projects from library : %s : %v", devfile.FileName, err)))
	} else {
		if projects != nil && len(projects) > 0 {
			err := VerifyProjects(devfile, projects)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verify Projects %s : %v", devfile.FileName, err)))
			}
		} else {
			commonUtils.LogInfoMessage(fmt.Sprintf("No projects found in %s : ", devfile.FileName))
		}
	}

	commonUtils.LogInfoMessage(fmt.Sprintf("Get starter projects %s : ", devfile.FileName))
	starterProjects, err := libraryData.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting Starter Projects from library : %s : %v", devfile.FileName, err)))
	} else {
		if starterProjects != nil && len(starterProjects) > 0 {
			err := VerifyStarterProjects(devfile, starterProjects)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verify Starter Projects %s : %v", devfile.FileName, err)))
			}
		} else {
			commonUtils.LogInfoMessage(fmt.Sprintf("No starter projects found in %s : ", devfile.FileName))
		}
	}

	var returnError error
	if len(errorString) > 0 {
		returnError = errors.New(fmt.Sprint(errorString))
	}
	return returnError

}

// editCommands modifies random attributes for each of the commands in the devfile.
func editCommands(devfile *commonUtils.TestDevfile) error {

	var errorString []string
	commonUtils.LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
	commands, err := devfile.Follower.(DevfileFollower).LibraryData.GetCommands(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting commands from library : %s : %v", devfile.FileName, err)))
	} else if len(commands) < 1 {
		errorString = append(errorString, commonUtils.LogErrorMessage("Updating commands : No commands returned"))
	} else {
		for _, command := range commands {
			err = UpdateCommand(devfile, command.Id)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Updating command : %s : %v", devfile.FileName, err)))
			}
		}
	}
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// editComponents modifies random attributes for each of the components in the devfile.
func editComponents(devfile *commonUtils.TestDevfile) error {

	var errorString []string
	commonUtils.LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get components %s : ", devfile.FileName))
	components, err := devfile.Follower.(DevfileFollower).LibraryData.GetComponents(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting components from library : %s : %v", devfile.FileName, err)))
	} else if len(components) < 1 {
		errorString = append(errorString, commonUtils.LogErrorMessage("Updating components : No components returned"))
	} else {
		for _, component := range components {
			err = UpdateComponent(devfile, component.Name)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Updating component : %s : %v", devfile.FileName, err)))
			}
		}
	}
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// editProjects modifies random attributes for each of the projects in the devfile.
func editProjects(devfile *commonUtils.TestDevfile) error {

	var errorString []string
	commonUtils.LogInfoMessage(fmt.Sprintf("Edit project %s : ", devfile.FileName))

	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get projects %s : ", devfile.FileName))
	projects, err := devfile.Follower.(DevfileFollower).LibraryData.GetProjects(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting projects from library : %s : %v", devfile.FileName, err)))
	} else if len(projects) < 1 {
		errorString = append(errorString, commonUtils.LogErrorMessage("Updating projects : No projects returned"))
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Updating projects : %d projects found.", len(projects)))
		for _, project := range projects {
			err = UpdateProject(devfile, project.Name)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Updating project : %v", err)))
			}
		}
	}
	return err
}

// editStarterProjects modifies random attributes for each of the starter projects in the devfile.
func editStarterProjects(devfile *commonUtils.TestDevfile) error {

	var errorString []string
	commonUtils.LogInfoMessage(fmt.Sprintf("Edit starter project %s : ", devfile.FileName))

	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get starter projects %s : ", devfile.FileName))
	starterProjects, err := devfile.Follower.(DevfileFollower).LibraryData.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Getting starter projects from library : %s : %v", devfile.FileName, err)))
	} else if len(starterProjects) < 1 {
		errorString = append(errorString, commonUtils.LogErrorMessage("Updating starter projects : No starter projects returned"))
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Updating starter projects : %d starter projects found.", len(starterProjects)))
		for _, starterProject := range starterProjects {
			err = UpdateStarterProject(devfile, starterProject.Name)
			if err != nil {
				errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Updating starter project : %v", err)))
			}
		}
	}
	return err
}
