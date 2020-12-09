package tests

import (
	"errors"
	"fmt"
	"strconv"

	schema "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
)

type GenericComponent struct {
	Name                     string
	Verified                 bool
	ComponentType            schema.ComponentType
	ContainerComponentSchema *schema.ContainerComponent
	VolumeComponentSchema    *schema.VolumeComponent
}

func (genericComponent *GenericComponent) setVerified() {
	genericComponent.Verified = true
}

func (genericComponent *GenericComponent) setName(name string) {
	genericComponent.Name = name
}

func (genericComponent *GenericComponent) checkId(coomponent schema.Component) bool {
	return genericComponent.Name == coomponent.Name
}

func AddVolume(numVols int) []schema.VolumeMount {
	commandVols := make([]schema.VolumeMount, numVols)
	for i := 0; i < numVols; i++ {
		commandVols[i].Name = "Name_" + GetRandomString(5, false)
		commandVols[i].Path = "/Path_" + GetRandomString(5, false)
		LogMessage(fmt.Sprintf("   ....... Add Volume: %s", commandVols[i]))
	}
	return commandVols
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

	genericComponent := GenericComponent{}
	genericComponent.ComponentType = componentType

	generateComponent(&components[index], &genericComponent)

	devfile.MapComponent(genericComponent)
	devfile.SchemaDevFile.Components = components

	return components[index].Name

}

func generateComponent(component *schema.Component, genericComponent *GenericComponent) {

	component.Name = GetRandomUniqueString(8, true)
	genericComponent.setName(component.Name)
	LogMessage(fmt.Sprintf("   ....... Name: %s", component.Name))

	if genericComponent.ComponentType == schema.ContainerComponentType {
		component.Container = createContainerComponent()
		genericComponent.ContainerComponentSchema = component.Container
	} else if genericComponent.ComponentType == schema.VolumeComponentType {
		component.Volume = createVolumeComponent()
		genericComponent.VolumeComponentSchema = component.Volume
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
	genericComponent := devfile.GetComponent(component.Name)
	if genericComponent != nil {
		LogMessage(fmt.Sprintf(" ....... Updating component name: %s", component.Name))
		if genericComponent.ComponentType == schema.ContainerComponentType {
			setContainerComponentValues(component.Container)
			genericComponent.ContainerComponentSchema = component.Container
		} else if genericComponent.ComponentType == schema.VolumeComponentType {
			setVolumeComponentValues(component.Volume)
			genericComponent.VolumeComponentSchema = component.Volume
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

func (devfile TestDevfile) VerifyComponents(components []schema.Component) error {

	LogMessage("Enter VerifyComponents")
	errorString := ""

	if devfile.ComponentMap != nil {
		for _, component := range components {

			LogMessage(fmt.Sprintf(" --> Volume Component structures matched - name : %s ", component.Name))
			if matchedComponent, found := devfile.ComponentMap[component.Name]; found {
				matchedComponent.setVerified()
				if matchedComponent.ComponentType == schema.ContainerComponentType {
					if !cmp.Equal(*component.Container, *matchedComponent.ContainerComponentSchema) {
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Container Component %s from parser: %v", component.Name, *component.Container))
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Container Component %s from tester: %v", matchedComponent.Name, matchedComponent.ContainerComponentSchema))
					} else {
						LogMessage(fmt.Sprintf(" --> Container Component structures matched - name : %s ", component.Name))
					}
				}
				if matchedComponent.ComponentType == schema.VolumeComponentType {
					if !cmp.Equal(*component.Volume, *matchedComponent.VolumeComponentSchema) {
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Volume Component %s from parser: %v", component.Name, *component.Volume))
						errorString += LogMessage(fmt.Sprintf(" ---> ERROR: Volume Component %s from tester: %v", matchedComponent.Name, matchedComponent.VolumeComponentSchema))
					} else {
						LogMessage(fmt.Sprintf(" --> Volume Component structures matched - name : %s ", component.Name))
					}
				}

			} else {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component from parser not known to test - - name : %s ", component.Name))
			}
		}

		for _, genericComponent := range devfile.ComponentMap {
			if !genericComponent.Verified {
				errorString += LogMessage(fmt.Sprintf(" --> ERROR: Component not returned by parser - name : %s", genericComponent.Name))
			}
		}

	} else {
		if components != nil {
			errorString += LogMessage(" --> ERROR: Parser returned components but Test does not include any.")
		}
	}
	var err error
	if errorString != "" {
		err = errors.New(errorString)
	}
	return err
}
