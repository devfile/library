package parserV200

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

func AddVolume(numVols int) []schema.VolumeMount {
	commandVols := make([]schema.VolumeMount, numVols)
	for i := 0; i < numVols; i++ {
		commandVols[i].Name = "Name_" + GetRandomString(5, false)
		commandVols[i].Path = "/Path_" + GetRandomString(5, false)
		LogMessage(fmt.Sprintf("   ....... Add Volume: %s", commandVols[i]))
	}
	return commandVols
}

func getSchemaComponent(components []schema.Component, name string) (*schema.Component, bool) {
	found := false
	var schemaComponent schema.Component
	for _, component := range components {
		if component.Name == name {
			schemaComponent = component
			found = true
			break
		}
	}
	return &schemaComponent, found
}

func (devfile *TestDevfile) AddComponent(componentType schema.ComponentType) string {

	var components []schema.Component
	index := 0
	if devfile.SchemaDevFile.Components != nil {
		// Commands already exist so expand the current slice
		currentSize := len(devfile.SchemaDevFile.Components)
		components = make([]schema.Component, currentSize+1)
		for _, component := range devfile.SchemaDevFile.Components {
			components[index] = component
			index++
		}
	} else {
		components = make([]schema.Component, 1)
	}

	generateComponent(&components[index], componentType)
	devfile.SchemaDevFile.Components = components

	return components[index].Name

}

func generateComponent(component *schema.Component, componentType schema.ComponentType) {

	component.Name = GetRandomUniqueString(8, true)
	LogMessage(fmt.Sprintf("   ....... Name: %s", component.Name))

	if componentType == schema.ContainerComponentType {
		component.Container = createContainerComponent()
	} else if componentType == schema.VolumeComponentType {
		component.Volume = createVolumeComponent()
	}
}

func createContainerComponent() *schema.ContainerComponent {

	LogMessage("Create a container component :")

	containerComponent := schema.ContainerComponent{}
	setContainerComponentValues(&containerComponent)

	return &containerComponent

}

func createVolumeComponent() *schema.VolumeComponent {

	LogMessage("Create a volume component :")

	volumeComponent := schema.VolumeComponent{}
	setVolumeComponentValues(&volumeComponent)

	return &volumeComponent

}

func setContainerComponentValues(containerComponent *schema.ContainerComponent) {

	containerComponent.Image = GetRandomUniqueString(8+GetRandomNumber(10), false)

	if GetBinaryDecision() {
		numCommands := GetRandomNumber(3)
		containerComponent.Command = make([]string, numCommands)
		for i := 0; i < numCommands; i++ {
			containerComponent.Command[i] = GetRandomString(4+GetRandomNumber(10), false)
			LogMessage(fmt.Sprintf("   ....... command %d of %d : %s", i, numCommands, containerComponent.Command[i]))
		}
	}

	if GetBinaryDecision() {
		numArgs := GetRandomNumber(3)
		containerComponent.Args = make([]string, numArgs)
		for i := 0; i < numArgs; i++ {
			containerComponent.Args[i] = GetRandomString(8+GetRandomNumber(10), false)
			LogMessage(fmt.Sprintf("   ....... arg %d of %d : %s", i, numArgs, containerComponent.Args[i]))
		}
	}

	containerComponent.DedicatedPod = GetBinaryDecision()
	LogMessage(fmt.Sprintf("   ....... DedicatedPod: %t", containerComponent.DedicatedPod))

	if GetBinaryDecision() {
		containerComponent.MemoryLimit = strconv.Itoa(4+GetRandomNumber(124)) + "M"
		LogMessage(fmt.Sprintf("   ....... MemoryLimit: %s", containerComponent.MemoryLimit))
	}

	if GetBinaryDecision() {
		setMountSources := GetBinaryDecision()
		containerComponent.MountSources = &setMountSources
		LogMessage(fmt.Sprintf("   ....... MountSources: %t", *containerComponent.MountSources))

		if setMountSources {
			containerComponent.SourceMapping = "/" + GetRandomString(8, false)
			LogMessage(fmt.Sprintf("   ....... SourceMapping: %s", containerComponent.SourceMapping))
		}
	}

	if GetBinaryDecision() {
		containerComponent.Env = AddEnv(GetRandomNumber(4))
	} else {
		containerComponent.Env = nil
	}

	if GetBinaryDecision() {
		containerComponent.VolumeMounts = AddVolume(GetRandomNumber(4))
	} else {
		containerComponent.Env = nil
	}

	if GetBinaryDecision() {
		containerComponent.Endpoints = CreateEndpoints()
	}

}

func setVolumeComponentValues(volumeComponent *schema.VolumeComponent) {

	if GetRandomDecision(5, 1) {
		volumeComponent.Size = strconv.Itoa(4+GetRandomNumber(252)) + "G"
		LogMessage(fmt.Sprintf("   ....... volumeComponent.Size: %s", volumeComponent.Size))
	}

}

func (devfile *TestDevfile) UpdateComponent(component *schema.Component) error {

	errorString := ""
	testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, component.Name)
	if found {
		LogMessage(fmt.Sprintf(" ....... Updating component name: %s", component.Name))
		if testComponent.ComponentType == schema.ContainerComponentType {
			setContainerComponentValues(component.Container)
		} else if testComponent.ComponentType == schema.VolumeComponentType {
			setVolumeComponentValues(component.Volume)
		}
	} else {
		errorString += LogMessage(fmt.Sprintf(" ....... Component not found in test : %s", component.Name))
	}
	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}

func (devfile TestDevfile) VerifyComponents(parserComponents []schema.Component) error {

	LogMessage("Enter VerifyComponents")
	errorString := ""

	// Compare entire array of commands
	if !cmp.Equal(parserComponents, devfile.SchemaDevFile.Components) {
		errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component array compare failed."))
		for _, component := range parserComponents {
			if testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, component.Name); found {
				if !cmp.Equal(component, *testComponent) {
					parserFilename := AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Parser")
					testFilename := AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Test")
					LogMessage(fmt.Sprintf("   .......marshall and write devfile %s", parserFilename))
					c, err := yaml.Marshal(component)
					if err == nil {
						err = ioutil.WriteFile(parserFilename, c, 0644)
					}
					LogMessage(fmt.Sprintf("   .......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testComponent)
					if err == nil {
						err = ioutil.WriteFile(testFilename, c, 0644)
					}
					errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component %s did not match, see files : %s and %s", component.Name, parserFilename, testFilename))
				} else {
					LogMessage(fmt.Sprintf(" --> Component matched : %s", component.Name))
				}
			} else {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component from parser not known to test - id : %s ", component.Name))
			}
		}
		for _, component := range devfile.SchemaDevFile.Components {
			if _, found := getSchemaComponent(parserComponents, component.Name); !found {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component from test not returned by parser : %s ", component.Name))
			}
		}
	} else {
		LogMessage(fmt.Sprintf(" --> Component structures matched"))
	}

	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}
