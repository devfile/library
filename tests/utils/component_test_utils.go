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

// addVolume returns volumeMounts in a schema structure based on a specified number of volumes
func addVolume(numVols int) []schema.VolumeMount {
	commandVols := make([]schema.VolumeMount, numVols)
	for i := 0; i < numVols; i++ {
		commandVols[i].Name = "name-" + GetRandomString(5, true)
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
func (devfile *TestDevfile) AddComponent(componentType schema.ComponentType) string {
	component := generateComponent(componentType)
	devfile.SchemaDevFile.Components = append(devfile.SchemaDevFile.Components, component)
	return component.Name
}

// generateComponent generates a component in a schema structure of the specified type
func generateComponent(componentType schema.ComponentType) schema.Component {

	component := schema.Component{}
	component.Name = GetRandomUniqueString(8, true)
	LogInfoMessage(fmt.Sprintf("....... Name: %s", component.Name))

	if componentType == schema.ContainerComponentType {
		component.Container = createContainerComponent()
	} else if componentType == schema.VolumeComponentType {
		component.Volume = createVolumeComponent()
	}
	return component
}

// createContainerComponent creates a container component and set its attribute values
func createContainerComponent() *schema.ContainerComponent {

	LogInfoMessage("Create a container component :")

	containerComponent := schema.ContainerComponent{}
	setContainerComponentValues(&containerComponent)

	return &containerComponent

}

// createVolumeComponent creates a volume component and set its attribute values
func createVolumeComponent() *schema.VolumeComponent {

	LogInfoMessage("Create a volume component :")

	volumeComponent := schema.VolumeComponent{}
	setVolumeComponentValues(&volumeComponent)

	return &volumeComponent

}

// setContainerComponentValues randomly sets container component attributes to random values
func setContainerComponentValues(containerComponent *schema.ContainerComponent) {

	containerComponent.Image = GetRandomUniqueString(8+GetRandomNumber(10), false)

	if GetBinaryDecision() {
		numCommands := GetRandomNumber(3)
		containerComponent.Command = make([]string, numCommands)
		for i := 0; i < numCommands; i++ {
			containerComponent.Command[i] = GetRandomString(4+GetRandomNumber(10), false)
			LogInfoMessage(fmt.Sprintf("....... command %d of %d : %s", i, numCommands, containerComponent.Command[i]))
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

	if GetBinaryDecision() {
		containerComponent.VolumeMounts = addVolume(GetRandomNumber(4))
	} else {
		containerComponent.VolumeMounts = nil
	}

	if GetBinaryDecision() {
		containerComponent.Endpoints = CreateEndpoints()
	}

}

// setVolumeComponentValues randomly sets volume component attributes to random values
func setVolumeComponentValues(volumeComponent *schema.VolumeComponent) {

	if GetRandomDecision(5, 1) {
		volumeComponent.Size = strconv.Itoa(4+GetRandomNumber(252)) + "G"
		LogInfoMessage(fmt.Sprintf("....... volumeComponent.Size: %s", volumeComponent.Size))
	}

}

// UpdateComponent randomly updates the attribute values of a specified component
func (devfile *TestDevfile) UpdateComponent(component *schema.Component) error {

	var errorString []string
	testComponent, found := getSchemaComponent(devfile.SchemaDevFile.Components, component.Name)
	if found {
		LogInfoMessage(fmt.Sprintf("....... Updating component name: %s", component.Name))
		if testComponent.ComponentType == schema.ContainerComponentType {
			setContainerComponentValues(component.Container)
		} else if testComponent.ComponentType == schema.VolumeComponentType {
			setVolumeComponentValues(component.Volume)
		}
	} else {
		errorString = append(errorString, LogInfoMessage(fmt.Sprintf("....... Component not found in test : %s", component.Name)))
	}
	var err error
	if len(errorString) > 0 {
		err = errors.New(fmt.Sprint(errorString))
	}
	return err
}

// VerifyComponents verifies components returned by the parser are the same as those saved in the devfile schema
func (devfile TestDevfile) VerifyComponents(parserComponents []schema.Component) error {

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
