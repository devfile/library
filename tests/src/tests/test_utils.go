package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"sigs.k8s.io/yaml"
)

const tmpDir = "./tmp/"
const logErrorOnly = false
const logFileName = "test.log"
const logToFileOnly = true // If set to false the log output will also be output to the console

var (
	testLogger *log.Logger
)

func init() {
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		os.RemoveAll(tmpDir)
	}
	os.Mkdir(tmpDir, 0755)

	f, err := os.OpenFile(filepath.Join(tmpDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error creating Log file : %v", err)
	} else {
		if logToFileOnly {
			testLogger = log.New(f, "", log.LstdFlags|log.Lmicroseconds)
		} else {
			writer := io.MultiWriter(f, os.Stdout)
			testLogger = log.New(writer, "", log.LstdFlags|log.Lmicroseconds)
		}
		testLogger.Println("Test Starting:")
	}
}

func GetTempDir() string {
	_, fn, _, ok := runtime.Caller(1)
	if !ok {
		return tmpDir
	}
	testFile := filepath.Base(fn)
	testFileExtension := filepath.Ext(testFile)
	subdir := testFile[0 : len(testFile)-len(testFileExtension)]
	return CreateTempDir(subdir)
}

func CreateTempDir(subdir string) string {
	tempDir := tmpDir + subdir + "/"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}
	return tempDir
}

func GetDevFileName() string {
	pc, fn, _, ok := runtime.Caller(1)
	if !ok {
		return tmpDir + "DefaultDevfile"
	}

	testFile := filepath.Base(fn)
	testFileExtension := filepath.Ext(testFile)
	subdir := testFile[0 : len(testFile)-len(testFileExtension)]
	destDir := CreateTempDir(subdir)
	callerName := runtime.FuncForPC(pc).Name()

	pos1 := strings.LastIndex(callerName, "/tests.") + len("/tests.")
	devfileName := callerName[pos1:len(callerName)]

	LogMessage(fmt.Sprintf("GetDevFileName : %s", destDir+devfileName))

	return destDir + devfileName
}

func LogMessage(message string) string {
	testLogger.Println(message)
	return message
}

type TestDevfile struct {
	SchemaDevFile   schema.Devfile
	FileName        string
	ParsedSchemaObj parser.DevfileObj
	SchemaParsed    bool
	CommandMap      map[string]*GenericCommand
	ComponentMap    map[string]*GenericComponent
}

var StringCount int = 0

var RndSeed int64 = time.Now().UnixNano()

func GetRandomUniqueString(n int, lower bool) string {
	StringCount++
	return fmt.Sprintf("%s%04d", GetRandomString(n, lower), StringCount)
}

func setRandSeed() {
	RndSeed++
	rand.Seed(RndSeed)
}

const schemaBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GetRandomString(n int, lower bool) string {
	setRandSeed()
	b := make([]byte, n)
	for i := range b {
		b[i] = schemaBytes[rand.Intn(len(schemaBytes)-1)]
	}
	randomString := string(b)
	if lower {
		randomString = strings.ToLower(randomString)
	}
	return randomString
}

func GetGroupKinds() []schema.CommandGroupKind {
	setRandSeed()
	return []schema.CommandGroupKind{schema.BuildCommandGroupKind, schema.RunCommandGroupKind, schema.TestCommandGroupKind, schema.DebugCommandGroupKind}
}

func GetRandomGroupKind() schema.CommandGroupKind {
	setRandSeed()
	return GetGroupKinds()[rand.Intn(len(GetGroupKinds()))]
}

func GetBinaryDecision() bool {
	return GetRandomDecision(1, 1)
}

func GetRandomDecision(success int, failure int) bool {
	setRandSeed()
	return rand.Intn(success+failure) > failure-1
}

func GetRandomNumber(max int) int {
	setRandSeed()
	return rand.Intn(max) + 1
}

func GetDevfile(fileName string) TestDevfile {
	testDevfile := TestDevfile{}
	testDevfile.SchemaDevFile = schema.Devfile{}
	testDevfile.FileName = fileName
	testDevfile.SchemaDevFile.SchemaVersion = "2.0.0"
	testDevfile.SchemaParsed = false
	return testDevfile
}

func (devfile *TestDevfile) MapCommand(command GenericCommand) {

	if devfile.CommandMap == nil {
		devfile.CommandMap = make(map[string]*GenericCommand)
	}
	LogMessage(fmt.Sprintf("   .......Add command to commandMap : %s", command.Id))
	devfile.CommandMap[command.Id] = &command
}

func (devfile *TestDevfile) GetCommand(id string) *GenericCommand {

	genericCommand := devfile.CommandMap[id]
	return genericCommand

}

func (devfile *TestDevfile) MapComponent(component GenericComponent) {

	if devfile.ComponentMap == nil {
		devfile.ComponentMap = make(map[string]*GenericComponent)
	}
	LogMessage(fmt.Sprintf("   .......Add component to componentMap : %s", component.Name))
	devfile.ComponentMap[component.Name] = &component
}

func (devfile *TestDevfile) GetComponent(name string) *GenericComponent {
	return devfile.ComponentMap[name]
}

func (devfile *TestDevfile) CreateDevfile(useParser bool) error {
	var err error
	if useParser {
		LogMessage(fmt.Sprintf("   .......use Parser to write devfile %s", devfile.FileName))
		newDevfile, err := devfileData.NewDevfileData(devfile.SchemaDevFile.SchemaVersion)
		if err != nil {
			LogMessage(fmt.Sprintf(" ..... ERROR: creating new devfile : %v", err))
		} else {
			newDevfile.SetSchemaVersion(devfile.SchemaDevFile.SchemaVersion)

			// add the commands to new devfile
			for _, command := range devfile.SchemaDevFile.Commands {
				newDevfile.AddCommands(command)
			}
			// add components to the new devfile
			newDevfile.AddComponents(devfile.SchemaDevFile.Components)

			ctx := devfileCtx.NewDevfileCtx(devfile.FileName)

			err = ctx.SetAbsPath()
			if err != nil {
				LogMessage(fmt.Sprintf(" ..... ERROR: setting devfile path : %v", err))
			} else {
				devObj := parser.DevfileObj{
					Ctx:  ctx,
					Data: newDevfile,
				}
				err = devObj.WriteYamlDevfile()
				if err != nil {
					LogMessage(fmt.Sprintf(" ..... ERROR: wriring devfile : %v", err))
				}
			}

		}
	} else {
		LogMessage(fmt.Sprintf("   .......marshall and write devfile %s", devfile.FileName))
		c, err := yaml.Marshal(&(devfile.SchemaDevFile))

		if err == nil {
			err = ioutil.WriteFile(devfile.FileName, c, 0644)
		}
	}
	if err == nil {
		devfile.SchemaParsed = false
	}
	return err
}

func (devfile *TestDevfile) ParseSchema() error {

	var err error
	if !devfile.SchemaParsed {
		LogMessage(fmt.Sprintf(" -> Parse and Validate %s : ", devfile.FileName))
		devfile.ParsedSchemaObj, err = devfilepkg.ParseAndValidate(devfile.FileName)
		if err != nil {
			LogMessage(fmt.Sprintf(" ......ERROR from ParseAndValidate %v : ", err))
		}
		devfile.SchemaParsed = true
	}
	return err
}

func (devfile TestDevfile) Verify() error {

	LogMessage(fmt.Sprintf("Verify %s : ", devfile.FileName))

	err := devfile.ParseSchema()

	if err == nil {
		LogMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
		commands, _ := devfile.ParsedSchemaObj.Data.GetCommands(common.DevfileOptions{})
		if commands != nil && len(commands) > 0 {
			err = devfile.VerifyCommands(commands)
		} else {
			LogMessage(fmt.Sprintf("  No command found in %s : ", devfile.FileName))
		}
	}

	if err == nil {
		LogMessage(fmt.Sprintf(" -> Get components %s : ", devfile.FileName))
		components, _ := devfile.ParsedSchemaObj.Data.GetComponents(common.DevfileOptions{})
		if components != nil && len(components) > 0 {
			err = devfile.VerifyComponents(components)
		} else {
			LogMessage(fmt.Sprintf("  No components found in %s : ", devfile.FileName))
		}
	}

	return err

}

func (devfile TestDevfile) EditCommands() error {

	LogMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	err := devfile.ParseSchema()
	if err == nil {
		LogMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
		commands, _ := devfile.ParsedSchemaObj.Data.GetCommands(common.DevfileOptions{})
		for _, command := range commands {
			err = devfile.UpdateCommand(&command)
			if err != nil {
				LogMessage(fmt.Sprintf(" ..... ERROR: updating command : %v", err))
			} else {
				LogMessage(fmt.Sprintf(" ..... Update command in Parser : %s", command.Id))
				devfile.ParsedSchemaObj.Data.UpdateCommand(command)
			}
		}
		LogMessage(fmt.Sprintf(" ..... Write updated file to yaml : %s", devfile.FileName))
		devfile.ParsedSchemaObj.WriteYamlDevfile()
		devfile.SchemaParsed = false
	} else {
		LogMessage(fmt.Sprintf(" ..... ERROR: from parser : %v", err))
	}
	return err

}

func (devfile TestDevfile) EditComponents() error {

	LogMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	err := devfile.ParseSchema()
	if err == nil {
		LogMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
		components, _ := devfile.ParsedSchemaObj.Data.GetComponents(common.DevfileOptions{})
		for _, component := range components {
			err = devfile.UpdateComponent(&component)
			if err != nil {
				LogMessage(fmt.Sprintf(" ..... ERROR: updating component : %v", err))
			} else {
				LogMessage(fmt.Sprintf(" ..... Update component in Parser : %s", component.Name))
				devfile.ParsedSchemaObj.Data.UpdateComponent(component)
			}
		}
		LogMessage(fmt.Sprintf(" ..... Write updated file to yaml : %s", devfile.FileName))
		devfile.ParsedSchemaObj.WriteYamlDevfile()
		devfile.SchemaParsed = false
	} else {
		LogMessage(fmt.Sprintf(" ..... ERROR: from parser : %v", err))
	}
	return err

}
