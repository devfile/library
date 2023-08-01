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

package parser

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/dockercompose"
)

// GetDeployComponents gets the default deploy command associated components
func GetDeployComponents(devfileData data.DevfileData) (map[string]string, error) {
	deployCommandFilter := common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: devfilev1.DeployCommandGroupKind,
		},
	}
	deployCommands, err := devfileData.GetCommands(deployCommandFilter)
	if err != nil {
		return nil, err
	}

	deployAssociatedComponents := make(map[string]string)
	var deployAssociatedSubCommands []string

	for _, command := range deployCommands {
		if command.Apply != nil {
			if len(deployCommands) > 1 && command.Apply.Group.IsDefault != nil && !*command.Apply.Group.IsDefault {
				continue
			}
			deployAssociatedComponents[command.Apply.Component] = command.Apply.Component
		} else if command.Composite != nil {
			if len(deployCommands) > 1 && command.Composite.Group.IsDefault != nil && !*command.Composite.Group.IsDefault {
				continue
			}
			deployAssociatedSubCommands = append(deployAssociatedSubCommands, command.Composite.Commands...)
		}
	}

	applyCommandFilter := common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: devfilev1.ApplyCommandType,
		},
	}
	applyCommands, err := devfileData.GetCommands(applyCommandFilter)
	if err != nil {
		return nil, err
	}

	for _, command := range applyCommands {
		if command.Apply != nil {
			for _, deployCommand := range deployAssociatedSubCommands {
				if deployCommand == command.Id {
					deployAssociatedComponents[command.Apply.Component] = command.Apply.Component
				}
			}

		}
	}

	return deployAssociatedComponents, nil
}

// GetImageBuildComponent gets the image build component from the deploy associated components
func GetImageBuildComponent(devfileData data.DevfileData, deployAssociatedComponents map[string]string) (devfilev1.Component, error) {
	imageComponentFilter := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: devfilev1.ImageComponentType,
		},
	}

	imageComponents, err := devfileData.GetComponents(imageComponentFilter)
	if err != nil {
		return devfilev1.Component{}, err
	}

	var imageBuildComponent devfilev1.Component
	for _, component := range imageComponents {
		if _, ok := deployAssociatedComponents[component.Name]; ok && component.Image != nil {
			if reflect.DeepEqual(imageBuildComponent, devfilev1.Component{}) {
				imageBuildComponent = component
			} else {
				errMsg := "expected to find one devfile image component with a deploy command for build. Currently there is more than one image component"
				return devfilev1.Component{}, fmt.Errorf(errMsg)
			}
		}
	}

	// If there is not one image component defined in the deploy command, err out
	if reflect.DeepEqual(imageBuildComponent, devfilev1.Component{}) {
		errMsg := "expected to find one devfile image component with a deploy command for build. Currently there is no image component"
		return devfilev1.Component{}, fmt.Errorf(errMsg)
	}

	return imageBuildComponent, nil
}

//ConvertDockerComposeToK8s convert Docker Compose to k8

func convertDockerComposeToK8s(filepath string) (error, string) {
	currentDir := path.Dir(filepath)
	client, err := dockercompose.NewClient(dockercompose.WithErrorOnWarning())
	if err != nil {
		return err, ""
	}

	outputDir := path.Join(currentDir, "services/")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err, ""
	}

	_, err = client.Convert(dockercompose.ConvertOptions{
		OutFile:    outputDir,
		InputFiles: []string{filepath},
	})
	if err != nil {
		return err, ""
	}
	return err, outputDir

}

//Convert k8s representation to k8s components

func createk8sComponents(devObj *DevfileObj, tempDir string) error {
	files, err := devObj.Ctx.GetFs().ReadDir(tempDir)
	if err != nil {
		return err
	}
	var components []devfilev1.Component

	for _, file := range files {
		filepath := path.Join(tempDir, file.Name())
		content, err := devObj.Ctx.GetFs().ReadFile(filepath)
		if err != nil {
			return err
		}
		component := devfilev1.Component{
			Name: strings.TrimSuffix(file.Name(), path.Ext(file.Name())),
			ComponentUnion: devfilev1.ComponentUnion{
				Kubernetes: &devfilev1.KubernetesComponent{
					K8sLikeComponent: devfilev1.K8sLikeComponent{
						K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
							Inlined: string(content),
						},
					},
				},
			},
		}
		if component.Attributes == nil {
			component.Attributes = attributes.Attributes{}
		}
		component.Attributes.PutString(ComposeGeneratedComponentKey, "true")
		components = append(components, component)

	}

	return devObj.Data.AddComponents(components)
}
