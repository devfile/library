# Devfile Library

<div id="header">

![Go](https://img.shields.io/badge/Go-1.19-blue)
[![Apache2.0 License](https://img.shields.io/badge/license-Apache2.0-brightgreen.svg)](LICENSE)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8231/badge)](https://www.bestpractices.dev/projects/8231)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/devfile/library/badge)](https://securityscorecards.dev/viewer/?uri=github.com/devfile/library)
</div>

## About

The Devfile Parser library is a Golang module that:
1. parses a devfile as specified by the [api](https://devfile.io/docs/2.2.1/devfile-schema) & [schema](https://github.com/devfile/api/tree/main/schemas/latest).
2. writes to the specified devfile with the updated data.
3. generates Kubernetes objects for the various devfile resources.
4. defines util functions for the devfile.
5. downloads resources from a parent devfile if specified in the devfile.

## Private repository support

Tokens are required to be set in the following cases:
1. parsing a devfile from a private repository
2. parsing a devfile containing a parent devfile from a private repository [1]
3. parsing a devfile from a private repository containing a parent devfile from a public repository [2]

Set the token for the repository:
```go
parser.ParserArgs{
	...
	// URL must point to a devfile.yaml
	URL: <url-to-devfile-on-supported-git-provider-repo>/devfile.yaml
	Token: <repo-personal-access-token>
	...
}
```
Note: The url must also be set with a supported git provider repo url.

Minimum token scope required:
1. GitHub: Read access to code
2. GitLab: Read repository
3. Bitbucket: Read repository

Note: To select token scopes for GitHub, a fine-grained token is required.

For more information about personal access tokens:
1. [GitHub docs](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
2. [GitLab docs](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token)
3. [Bitbucket docs](https://support.atlassian.com/bitbucket-cloud/docs/repository-access-tokens/)

[1] Currently, this works under the assumption that the token can authenticate the devfile and the parent devfile; both devfiles are in the same repository.

[2] In this scenario, the token will be used to authenticate the main devfile.

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

8. By default, the parser will set all unset boolean properties to their spec defined default values.  Clients can override this behaviour by specifiying the parser argument `SetBooleanDefaults` to false
   ```go
   setDefaults := false
   parserArgs := parser.ParserArgs{
		SetBooleanDefaults:               &setDefaults,
   }
   ```

9. When parsing a devfile that contains a parent reference, if the parent uri is a supported git provider repo url with the correct personal access token, all resources from the parent git repo excluding the parent devfile.yaml will be downloaded to the location of the devfile being parsed. **Note: The URL must point to a devfile.yaml**
   ```yaml
   schemaVersion: 2.2.0
   ...
   parent:
      uri: <uri-to-parent-devfile>/devfile.yaml
   ...
   ```

10. By default, the library downloads the Git repository resources associated with the Git URL that is mentioned in a devfile uri field. To turn off the download, pass in the `DownloadGitResources` property in the parser argument
   ```go
   downloadGitResources := false
   parserArgs := parser.ParserArgs{
		DownloadGitResources:               &downloadGitResources,
   }
   ```

11. To download/access files from a private repository like a private GitHub use the `Token` property
   ```go
   parserArgs := parser.ParserArgs{
		Token: "my-PAT",
   }
   ```

   ```go
   src: YamlSrc{
		URL: "http://github.com/my-private-repo",
		Token: "my-PAT",
   }
   values, err := ReadKubernetesYaml(src, fs, nil)
   ```

   If you would like to use the mock implementation for the `DevfileUtils` interface method defined in [pkg/devfile/parser/util/interface.go](pkg/devfile/parser/util/interface.go), then use 
   ```go
   var devfileUtilsClient DevfileUtils
   devfileUtilsClient = NewMockDevfileUtilsClient()
   devfileUtilsClient.DownloadInMemory(params)
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

The devfile/library releases are created annually or on demand. For existing devfile/library releases, please check the release [page](https://github.com/devfile/library/releases).

### Create a New Release

The steps to create a new release are:

* Create a separate branch for the particular release, for example, `v2.0.x`

* Run this [script](https://github.com/devfile/library/blob/main/scripts/changelog-script.sh) to generate release changelog.
```bash
# generate a changelog for all the changes since release v2.0.0
./changelog-script.sh v2.0.0
```

* Create a new release [here](https://github.com/devfile/library/releases/new) with a new tag (having the same name with the above branch - e.g. `v2.0.x`) and copy the generated changelog to the details

## Contributing

Please see our [contributing.md](./CONTRIBUTING.md).
