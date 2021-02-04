package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

// componentAdded adds a new component to the test schema data and to the parser data
func (devfile *TestDevfile)componentAdded(component schema.Component) {
	LogInfoMessage(fmt.Sprintf("component added Name: %s", component.Name))
	devfile.SchemaDevFile.Components = append(devfile.SchemaDevFile.Components, component)
	devfile.ParserData.AddComponents([]schema.Component{component})
}

// componetUpdated updates a component in the parser data
func (devfile *TestDevfile) componentUpdated(component schema.Component) {
	LogInfoMessage(fmt.Sprintf("component updated Name: %s", component.Name))
	devfile.ParserData.UpdateComponent(component)
}

// addVolume returns volumeMounts in a schema structure based on a specified number of volumes
func (devfile *TestDevfile) addVolume(numVols int) []schema.VolumeMount {
	commandVols := make([]schema.VolumeMount, numVols)
	for i := 0; i < numVols; i++ {
		volumeComponent := devfile.AddComponent(schema.VolumeComponentType)
		commandVols[i].Name = volumeComponent.Name
		commandVols[i].Path = "/Path_" + GetRandomString(5, false)
		LogInfoMessage(fmt.Sprintf("....... Add Volume: %s", commandVols[i]))
	}
	return commandVols
}

// getSchemaComponent returns a named component from an array of components
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


// AddComponent adds a component of the specified type, with random attributes, to the devfile schema
func (devfile *TestDevfile) AddComponent(componentType schema.ComponentType) schema.Component {

	var component schema.Component
	if componentType == schema.ContainerComponentType {
		component = devfile.createContainerComponent()
		devfile.setContainerComponentValues(&component)
	} else if componentType == schema.VolumeComponentType {
		component = devfile.createVolumeComponent()
		devfile.setVolumeComponentValues(&component)
	}
	return component
}

// createContainerComponent creates a container component, ready for attribute setting
func (devfile *TestDevfile) createContainerComponent() schema.Component {

	LogInfoMessage("Create a container component :")
	component := schema.Component{}
	component.Name = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("....... Name: %s", component.Name))
	component.Container = &schema.ContainerComponent{}
	devfile.componentAdded(component)
	return component

}

// createVolumeComponent creates a volume component , ready for attribute setting
func (devfile *TestDevfile) createVolumeComponent() schema.Component {

	LogInfoMessage("Create a volume component :")
	component := schema.Component{}
	component.Name = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("....... Name: %s", component.Name))
	component.Volume = &schema.VolumeComponent{}
	devfile.componentAdded(component)
	return component

}

// AddCommandToContainer adds a command id to a container, creating one if necessary.
func (devfile *TestDevfile) AddCommandToContainer(commandId string) string {

	LogInfoMessage(fmt.Sprintf("add command %s to a container.",commandId))
	componentName := ""
	for _,currentComponent := range devfile.SchemaDevFile.Components {
		if currentComponent.Container != nil {
			currentComponent.Container.Command = append(currentComponent.Container.Command,commandId)
			componentName = currentComponent.Name
			LogInfoMessage(fmt.Sprintf("add command to existing container : %s",componentName))
			devfile.componentUpdated(currentComponent)
			break;
		}
	}

	if componentName == "" {
		component := devfile.createContainerComponent()
		component.Container.Command = append(component.Container.Command,commandId)
		componentName = component.Name
		LogInfoMessage(fmt.Sprintf("add command to a new container : %s",componentName))
		devfile.componentUpdated(component)
	}

	return componentName
}

// setContainerComponentValues randomly sets container component attributes to random values
func (devfile *TestDevfile) setContainerComponentValues(component *schema.Component) {

	containerComponent := component.Container

	containerComponent.Image = GetRandomUniqueString(8+GetRandomNumber(10), false)

	if GetBinaryDecision() {
		numCommands := GetRandomNumber(3)
		for i := 0; i < numCommands; i++ {
			commandId := GetRandomString(4+GetRandomNumber(10),true)
			containerComponent.Command = append(containerComponent.Command,commandId)
			LogInfoMessage(fmt.Sprintf("....... command %d added : %s", len(containerComponent.Command),commandId))
		}
	}

	if GetBinaryDecision() {
		numArgs := GetRandomNumber(3)
		containerComponent.Args = make([]string, numArgs)
		for i := 0; i < numArgs; i++ {
			containerComponent.Args[i] = GetRandomString(8+GetRandomNumber(10), false)
			LogInfoMessage(fmt.Sprintf("....... arg %d of %d : %s", i, numArgs, containerComponent.Args[i]))
		}
	}

	containerComponent.DedicatedPod = GetBinaryDecision()
	LogInfoMessage(fmt.Sprintf("....... DedicatedPod: %t", containerComponent.DedicatedPod))

	if GetBinaryDecision() {
		containerComponent.MemoryLimit = strconv.Itoa(4+GetRandomNumber(124)) + "M"
		LogInfoMessage(fmt.Sprintf("....... MemoryLimit: %s", containerComponent.MemoryLimit))
	}

	if GetBinaryDecision() {
		setMountSources := GetBinaryDecision()
		containerComponent.MountSources = &setMountSources
		LogInfoMessage(fmt.Sprintf("....... MountSources: %t", *containerComponent.MountSources))

		if setMountSources {
			containerComponent.SourceMapping = "/" + GetRandomString(8, false)
			LogInfoMessage(fmt.Sprintf("....... SourceMapping: %s", containerComponent.SourceMapping))
		}
	}

	if GetBinaryDecision() {
		containerComponent.Env = addEnv(GetRandomNumber(4))
	} else {
		containerComponent.Env = nil
	}

	if len(containerComponent.VolumeMounts) == 0 {
		if GetBinaryDecision() {
			containerComponent.VolumeMounts = devfile.addVolume(GetRandomNumber(4))
		}
	}

	if GetBinaryDecision() {
		containerComponent.Endpoints = devfile.CreateEndpoints()
	}

	devfile.componentUpdated(*component)

}

// setVolumeComponentValues randomly sets volume component attributes to random values
func (devfile *TestDevfile) setVolumeComponentValues(component *schema.Component) {

	component.Volume.Size = strconv.Itoa(4+GetRandomNumber(252)) + "G"
	LogInfoMessage(fmt.Sprintf("....... volumeComponent.Size: %s", component.Volume.Size))
	devfile.componentUpdated(*component)

}

// UpdateComponent randomly updates the attribute values of a specified component
func (devfile *TestDevfile) UpdateComponent(componentName string) error {

	var errorString []string
	testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, componentName)
	if found {
		LogInfoMessage(fmt.Sprintf("....... Updating component name: %s", componentName))
		if testComponent.Container != nil {
			devfile.setContainerComponentValues(testComponent)
		} else if testComponent.Volume != nil {
			devfile.setVolumeComponentValues(testComponent)
		} else {
			errorString = append(errorString, LogInfoMessage(fmt.Sprintf("....... Component is not of expected type.")))
		}
	} else {
		errorString = append(errorString, LogInfoMessage(fmt.Sprintf("....... Component not found in test : %s", componentName)))
	}
	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// VerifyComponents verifies components returned by the parser are the same as those saved in the devfile schema
func (devfile *TestDevfile) VerifyComponents(parserComponents []schema.Component) error {

	LogInfoMessage("Enter VerifyComponents")
	var errorString []string

	// Compare entire array of components
	if !cmp.Equal(parserComponents, devfile.SchemaDevFile.Components) {
		errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Component array compare failed.")))
		for _, component := range parserComponents {
			if testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, component.Name); found {
				if !cmp.Equal(component, *testComponent) {
					parserFilename := AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Parser")
					testFilename := AddSuffixToFileName(devfile.FileName, "_"+component.Name+"_Test")
					LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", parserFilename))
					c, err := yaml.Marshal(component)
					if err != nil {
						errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", parserFilename)))
					} else {
						err = ioutil.WriteFile(parserFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......write devfile %s", parserFilename)))
						}
					}
					LogInfoMessage(fmt.Sprintf(".......marshall and write devfile %s", testFilename))
					c, err = yaml.Marshal(testComponent)
					if err != nil {
						errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......marshall devfile %s", testFilename)))
					} else {
						err = ioutil.WriteFile(testFilename, c, 0644)
						if err != nil {
							errorString = append(errorString, LogErrorMessage(fmt.Sprintf(".......write devfile %s", testFilename)))
						}
					}
					errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Component %s did not match, see files : %s and %s", component.Name, parserFilename, testFilename)))
				} else {
					LogInfoMessage(fmt.Sprintf(" --> Component matched : %s", component.Name))
				}
			} else {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Component from parser not known to test - id : %s ", component.Name)))
			}
		}
		for _, component := range devfile.SchemaDevFile.Components {
			if _, found := getSchemaComponent(parserComponents, component.Name); !found {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Component from test not returned by parser : %s ", component.Name)))
			}
		}
	} else {
		LogInfoMessage(fmt.Sprintf("Component structures matched"))
	}

	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}
