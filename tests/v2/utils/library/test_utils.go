package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	header "github.com/devfile/api/v2/pkg/devfile"
	devfilepkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"sigs.k8s.io/yaml"

	commonUtils "github.com/devfile/library/tests/v2/utils/common"
)

const (
	// numDevfiles : the number of devfiles to create for each test
	numDevfiles = 5
	// numThreads :  Number of threads used by multi-thread tests
	numThreads = 5
)

type DevfileValidator struct{}

type DevfileFollower struct {
	LibraryData devfileData.DevfileData
}

// WriteAndVerify implements Saved.DevfileValidator interface.
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

func (devfileFollower DevfileFollower) AddCommand(command schema.Command) error {
	return devfileFollower.LibraryData.AddCommands(command)
}

func (devfileFollower DevfileFollower) UpdateCommand(command schema.Command) {
	devfileFollower.LibraryData.UpdateCommand(command)
}

func (devfileFollower DevfileFollower) AddComponent(component schema.Component) error {
	var components []schema.Component
	components = append(components, component)
	return devfileFollower.LibraryData.AddComponents(components)
}

func (devfileFollower DevfileFollower) UpdateComponent(component schema.Component) {
	devfileFollower.LibraryData.UpdateComponent(component)
}

func (devfileFollower DevfileFollower) AddProject(project schema.Project) error {
	var projects []schema.Project
	projects = append(projects, project)
	return devfileFollower.LibraryData.AddProjects(projects)
}

func (devfileFollower DevfileFollower)UpdateProject(project schema.Project) {
	devfileFollower.LibraryData.UpdateProject(project)
}

func (devfileFollower DevfileFollower)AddStarterProject(starterProject schema.StarterProject) error {
	var starterProjects []schema.StarterProject
	starterProjects = append(starterProjects, starterProject)
	return devfileFollower.LibraryData.AddStarterProjects(starterProjects)
}

func (devfileFollower DevfileFollower)UpdateStarterProject(starterProject schema.StarterProject) {
	devfileFollower.LibraryData.UpdateStarterProject(starterProject)
}

func (devfileFollower DevfileFollower)AddEvent(event schema.Events) error {
	return devfileFollower.LibraryData.AddEvents(event)
}

func (devfileFollower DevfileFollower)UpdateEvent(event schema.Events) {
	devfileFollower.LibraryData.UpdateEvents(event.PreStart,event.PostStart,event.PreStop,event.PostStop)
}

func (devfileFollower DevfileFollower)SetParent(parent schema.Parent) error {
	devfileFollower.LibraryData.SetParent(&parent)
	return nil
}

func (devfileFollower DevfileFollower)UpdateParent(parent schema.Parent) {
	devfileFollower.LibraryData.SetParent(&parent)
}

func (devfileFollower DevfileFollower)SetMetaData(metaData header.DevfileMetadata) error {
	devfileFollower.LibraryData.SetMetadata(metaData.Name,metaData.Version)
	return nil
}

func (devfileFollower DevfileFollower)UpdateMetaData(metadata header.DevfileMetadata) {
	devfileFollower.LibraryData.SetMetadata(metadata.Name,metadata.Version)
}

func (devfileFollower DevfileFollower)SetSchemaVersion(schemaVersion string) {
	devfileFollower.LibraryData.SetSchemaVersion(schemaVersion)
}

// WriteDevfile create a devifle on disk for use in a test.
// If useParser is true the parser library is used to generate the file, otherwise "sigs.k8s.io/yaml" is used.
func writeDevfile(devfile *commonUtils.TestDevfile) error {
	var err error

	fileName := devfile.FileName
	if !strings.HasSuffix(fileName, ".yaml") {
		fileName += ".yaml"
	}

	if commonUtils.GetBinaryDecision() {
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

	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("Marshall and write devfile %s", devfile.FileName))
		c, marshallErr := yaml.Marshal(&(devfile.SchemaDevFile))

		if marshallErr != nil {
			err = errors.New(commonUtils.LogErrorMessage(fmt.Sprintf("Marshall devfile %s : %v", devfile.FileName, marshallErr)))
		} else {
			err = ioutil.WriteFile(fileName, c, 0644)
			if err != nil {
				commonUtils.LogErrorMessage(fmt.Sprintf("Write devfile %s : %v", devfile.FileName, err))
			}
		}
	}
	return err
}

// parseSchema uses the parser to parse a devfile on disk
func validateDevfile(devfile *commonUtils.TestDevfile) error {

	var err error

	commonUtils.LogInfoMessage(fmt.Sprintf("Parse and Validate %s : ", devfile.FileName))
	libraryObj, err := devfilepkg.ParseAndValidate(devfile.FileName)
	if err != nil {
		commonUtils.LogErrorMessage(fmt.Sprintf("From ParseAndValidate %v : ", err))
	} else {
		follower := devfile.Follower.(DevfileFollower)
		follower.LibraryData = libraryObj.Data
	}

	return err
}

// RunMultiThreadTest : Runs the same test on multiple threads, the test is based on the content of the specified TestContent
func RunMultiThreadTest(testContent commonUtils.TestContent, t *testing.T) {

	commonUtils.LogMessage(fmt.Sprintf("Start Threaded test for %s", testContent.FileName))

	devfileName := testContent.FileName
	var i int
	for i = 1; i < numThreads; i++ {
		testContent.FileName = commonUtils.AddSuffixToFileName(devfileName, strconv.Itoa(i))
		go RunTest(testContent, t)
	}
	testContent.FileName = commonUtils.AddSuffixToFileName(devfileName, strconv.Itoa(i))
	RunTest(testContent, t)

	commonUtils.LogMessage(fmt.Sprintf("Sleep 3 seconds to allow all threads to complete : %s", devfileName))
	time.Sleep(3 * time.Second)
	commonUtils.LogMessage(fmt.Sprintf("Sleep complete : %s", devfileName))

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
		libraryData, err := devfileData.NewDevfileData("2.0.0")
		if err != nil {
			t.Fatalf(commonUtils.LogMessage(fmt.Sprintf("Error creating parser data : %v", err)))
		}
		libraryData.SetSchemaVersion("2.0.0")
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

			validator.WriteAndValidate(&testDevfile)
		}
	}
}

// Verify verifies the contents of the specified devfile with the expected content
func verify(devfile *commonUtils.TestDevfile) error {

	commonUtils.LogInfoMessage(fmt.Sprintf("Verify %s : ", devfile.FileName))

	var errorString []string

	libraryData := devfile.Follower.(DevfileFollower).LibraryData
	commonUtils.LogInfoMessage(fmt.Sprintf("Get commands %s : ", devfile.FileName))
	commands, _ := libraryData.GetCommands(common.DevfileOptions{})
	if commands != nil && len(commands) > 0 {
		err := VerifyCommands(devfile, commands)
		if err != nil {
			errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verfify Commands %s : %v", devfile.FileName, err)))
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("No command found in %s : ", devfile.FileName))
	}

	commonUtils.LogInfoMessage(fmt.Sprintf("Get components %s : ", devfile.FileName))
	components, _ := libraryData.GetComponents(common.DevfileOptions{})
	if components != nil && len(components) > 0 {
		err := VerifyComponents(devfile, components)
		if err != nil {
			errorString = append(errorString, commonUtils.LogErrorMessage(fmt.Sprintf("Verfify Commands %s : %v", devfile.FileName, err)))
		}
	} else {
		commonUtils.LogInfoMessage(fmt.Sprintf("No components found in %s : ", devfile.FileName))
	}

	var returnError error
	if len(errorString) > 0 {
		returnError = errors.New(fmt.Sprint(errorString))
	}
	return returnError

}

// EditCommands modifies random attributes for each of the commands in the devfile.
func editCommands(devfile *commonUtils.TestDevfile) error {

	commonUtils.LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	var err error
	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
	commands, _ := devfile.Follower.(DevfileFollower).LibraryData.GetCommands(common.DevfileOptions{})
	for _, command := range commands {
		err = UpdateCommand(devfile, command.Id)
		if err != nil {
			commonUtils.LogErrorMessage(fmt.Sprintf("Updating command : %v", err))
		}
	}

	return err
}

// EditComponents modifies random attributes for each of the components in the devfile.
func editComponents(devfile *commonUtils.TestDevfile) error {

	commonUtils.LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	var err error
	commonUtils.LogInfoMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
	components, _ := devfile.Follower.(DevfileFollower).LibraryData.GetComponents(common.DevfileOptions{})
	for _, component := range components {
		err = UpdateComponent(devfile, component.Name)
		if err != nil {
			commonUtils.LogErrorMessage(fmt.Sprintf("Updating component : %v", err))
		}
	}
	return err
}
