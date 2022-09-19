# Devfile Library

<div id="header">

[![Apache2.0 License](https://img.shields.io/badge/license-Apache2.0-brightgreen.svg)](LICENSE)
</div>

## About

The Devfile Parser library is a Golang module that:
1. parses the devfile.yaml as specified by the [api](https://devfile.github.io/devfile/api-reference.html) & [schema](https://github.com/devfile/api/tree/main/schemas/latest).
2. writes to the devfile.yaml with the updated data.
3. generates Kubernetes objects for the various devfile resources.
4. defines util functions for the devfile.

## Usage

The function documentation can be accessed via [pkg.go.dev](https://pkg.go.dev/github.com/devfile/library). 
1. To parse a devfile, visit [parse.go source file](pkg/devfile/parse.go)
   ```go
   // ParserArgs is the struct to pass into parser functions which contains required info for parsing devfile.
   parserArgs := parser.ParserArgs{
		Path:                           path,
		FlattenedDevfile:               &flattenedDevfile,
		ConvertKubernetesContentInUri:  &convertKubernetesContentInUri
		RegistryURLs:                   registryURLs,
		DefaultNamespace:               defaultNamespace,
		Context:                        context,
		K8sClient:                      client,
   		ExternalVariables:              externalVariables,
	}

   // Parses the devfile and validates the devfile data
   // if top-level variables are not substituted successfully, the warnings can be logged by parsing variableWarning
   devfile, variableWarning, err := devfilePkg.ParseDevfileAndValidate(parserArgs)
   ```


2. To override the HTTP request and response timeouts for a devfile with a parent reference from a registry URL, specify the HTTPTimeout value in the parser arguments
   ```go
      // specify the timeout in seconds  
      httpTimeout := 20 
      parserArgs := parser.ParserArgs{
         HTTPTimeout: &httpTimeout
	  }
   ```

   
3. To get specific content from devfile
   ```go
   // To get all the components from the devfile
   components, err := devfile.Data.GetComponents(DevfileOptions{})

   // To get all the components from the devfile with attributes tagged - tool: console-import
   // & import: {strategy: Dockerfile}
   components, err := devfile.Data.GetComponents(DevfileOptions{
      Filter: map[string]interface{}{
			"tool": "console-import",
			"import": map[string]interface{}{
				"strategy": "Dockerfile",
			},
		},
   })

   // To get all the volume components
   components, err := devfile.Data.GetComponents(DevfileOptions{
		ComponentOptions: ComponentOptions{
			ComponentType: v1.VolumeComponentType,
		},
   })

   // To get all the exec commands that belong to the build group
   commands, err := devfile.Data.GetCommands(DevfileOptions{
		CommandOptions: CommandOptions{
			CommandType: v1.ExecCommandType,
			CommandGroupKind: v1.BuildCommandGroupKind,
		},
   })
   ```
   
4. To get the Kubernetes objects from the devfile, visit [generators.go source file](pkg/devfile/generator/generators.go)
   ```go
    // To get a slice of Kubernetes containers of type corev1.Container from the devfile component containers
    containers, err := generator.GetContainers(devfile)

    // To generate a Kubernetes deployment of type v1.Deployment
    deployParams := generator.DeploymentParams{
		TypeMeta:          generator.GetTypeMeta(deploymentKind, deploymentAPIVersion),
		ObjectMeta:        generator.GetObjectMeta(name, namespace, labels, annotations),
		InitContainers:    initContainers,
		Containers:        containers,
		Volumes:           volumes,
		PodSelectorLabels: labels,
	}
	deployment := generator.GetDeployment(deployParams)
   ```
   
5. To update devfile content
   ```go
   // To update an existing component in devfile object
   err := devfile.Data.UpdateComponent(v1.Component{
   	    Name: "component1",
   	    ComponentUnion: v1.ComponentUnion{
   	    	Container: &v1.ContainerComponent{
   	    		Container: v1.Container{
   	    			Image: "image1",
                },
            },
        },
   })

   // To add a new component to devfile object
   err := devfile.Data.AddComponents([]v1.Component{
        {
           Name: "component2",
           ComponentUnion: v1.ComponentUnion{
               Container: &v1.ContainerComponent{
                   Container: v1.Container{
                       Image: "image2",
                   },
               },
           },
        },
   })

   // To delete a component from the devfile object
   err := devfile.Data.DeleteComponent(componentName)
   ```

6. To write to a devfile, visit [writer.go source file](pkg/devfile/parser/writer.go)
   ```go
   // If the devfile object has been created with devfile path already set, can simply call WriteYamlDevfile to writes the devfile
   err := devfile.WriteYamlDevfile()
   
   
   // To write to a devfile from scratch
   // create a new DevfileData with a specific devfile version
   devfileData, err := data.NewDevfileData(devfileVersion)

   // set schema version
   devfileData.SetSchemaVersion(devfileVersion)
   
   // add devfile content use library APIs
   devfileData.AddComponents([]v1.Component{...})
   devfileData.AddCommands([]v1.Commands{...})
   ......
   
   // create a new DevfileCtx
   ctx := devfileCtx.NewDevfileCtx(devfilePath)
   err = ctx.SetAbsPath()

   // create devfile object with the new DevfileCtx and DevfileData
   devfile := parser.DevfileObj{
		Ctx:  ctx,
		Data: devfileData,
   }
    
   // write to the devfile on disk
   err = devfile.WriteYamlDevfile()
   ```
7. To parse the outerloop Kubernetes/OpenShift component's uri or inline content, call the read and parse functions
   ```go
   // Read the YAML content
   values, err := ReadKubernetesYaml(src, fs)

   // Get the Kubernetes resources
   resources, err := ParseKubernetesYaml(values)
   ```


## Projects using devfile/library

The following projects are consuming this library as a Golang dependency

* [odo](https://github.com/openshift/odo)
* [OpenShift Console](https://github.com/openshift/console)

## Tests

To run unit tests and api tests. Visit [library tests](tests/README.md) to find out more information on tests
```
make test
```

## Issues

Issues are tracked in the [devfile/api](https://github.com/devfile/api) repo with the label [area/library](https://github.com/devfile/api/issues?q=is%3Aopen+is%3Aissue+label%3Aarea%2Flibrary) 

## Releases

For devfile/library releases, please check the release [page](https://github.com/devfile/library/releases).

## Contributing

Please see our [contributing.md](./CONTRIBUTING.md).
