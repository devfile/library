//
// Copyright 2022-2023 Red Hat, Inc.
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

package devfile

import (
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
)

func TestParseDevfileAndValidate(t *testing.T) {
	falseValue := false
	trueValue := true
	convertUriToInline := false
	K8sLikeComponentOriginalURIKey := "api.devfile.io/k8sLikeComponent-originalURI"
	outerloopDeployContent := `
kind: Deployment
apiVersion: apps/v1
metadata:
  name: my-python
spec:
  replicas: 1
  selector:
    matchLabels:
      app: python-app
  template:
    metadata:
      labels:
        app: python-app
    spec:
      containers:
        - name: my-python
          image: my-python-image:{{ PARAMS }}
          ports:
            - name: http
              containerPort: 8081
              protocol: TCP
          resources:
            limits:
              memory: "128Mi"
              cpu: "500m"
`
	outerloopServiceContent := `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: python-app
  name: python-app-svc
spec:
  ports:
    - name: http-8081
      port: 8081
      protocol: TCP
      targetPort: 8081
  selector:
    app: python-app
	variable: {{ PARAMS }}
  type: LoadBalancer
`
	uri := "127.0.0.1:8080"
	var testServer *httptest.Server
	testServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		if strings.Contains(r.URL.Path, "/outerloop-deploy.yaml") {
			_, err = w.Write([]byte(outerloopDeployContent))
		} else if strings.Contains(r.URL.Path, "/outerloop-service.yaml") {
			_, err = w.Write([]byte(outerloopServiceContent))
		}
		if err != nil {
			t.Errorf("unexpected error while writing yaml: %v", err)
		}

	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", uri)
	if err != nil {
		t.Errorf("Test_parseParentAndPluginFromURI() unexpected error while creating listener: %v", err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	devfileContent := `commands:
- exec:
    commandLine: ./main {{ PARAMS }}
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
components:
- container:
    endpoints:
    - name: http
      targetPort: 8080
    image: golang:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
- kubernetes:
    uri: http://127.0.0.1:8080/outerloop-deploy.yaml
  name: outerloop-deploy
- openshift:
    uri: http://127.0.0.1:8080/outerloop-service.yaml
  name: outerloop-deploy2
metadata:
  description: Stack with the latest Go version
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: go
  name: my-go-app
  projectType: go
  tags:
  - Go
  version: 1.0.0
schemaVersion: 2.2.0
`

	devfileContentWithVariable := devfileContent + `variables:
  PARAMS: foo`
	type args struct {
		args parser.ParserArgs
	}
	tests := []struct {
		name                 string
		args                 args
		wantVarWarning       variables.VariableWarning
		wantCommandLine      string
		wantKubernetesInline string
		wantOpenshiftInline  string
		wantVariables        map[string]string
	}{
		{
			name: "with external overriding variables",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"PARAMS": "bar",
					},
					Data: []byte(devfileContentWithVariable),
				},
			},
			wantKubernetesInline: "image: my-python-image:bar",
			wantOpenshiftInline:  "variable: bar",
			wantCommandLine:      "./main bar",
			wantVariables: map[string]string{
				"PARAMS": "bar",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		},
		{
			name: "with new external variables",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"OTHER": "other",
					},
					Data: []byte(devfileContentWithVariable),
				},
			},
			wantKubernetesInline: "image: my-python-image:foo",
			wantOpenshiftInline:  "variable: foo",
			wantCommandLine:      "./main foo",
			wantVariables: map[string]string{
				"PARAMS": "foo",
				"OTHER":  "other",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		}, {
			name: "with new external variables",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"PARAMS": "baz",
					},
					Data: []byte(devfileContent),
				},
			},
			wantKubernetesInline: "image: my-python-image:baz",
			wantOpenshiftInline:  "variable: baz",
			wantCommandLine:      "./main baz",
			wantVariables: map[string]string{
				"PARAMS": "baz",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		}, {
			name: "with external variables and covertUriToInline is false",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"PARAMS": "baz",
					},
					ConvertKubernetesContentInUri: &convertUriToInline,
					Data:                          []byte(devfileContent),
				},
			},
			wantCommandLine: "./main baz",
			wantVariables: map[string]string{
				"PARAMS": "baz",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		},
		{
			name: "with flattening set to false and setBooleanDefaults to true",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"PARAMS": "baz",
					},
					FlattenedDevfile: &falseValue,
					Data:             []byte(devfileContent),
				},
			},
			wantCommandLine: "./main baz",
			wantVariables: map[string]string{
				"PARAMS": "baz",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		},
		{
			name: "with setBooleanDefaults to false",
			args: args{
				args: parser.ParserArgs{
					ExternalVariables: map[string]string{
						"PARAMS": "baz",
					},
					SetBooleanDefaults: &falseValue,
					Data:               []byte(devfileContent),
				},
			},
			wantCommandLine: "./main baz",
			wantVariables: map[string]string{
				"PARAMS": "baz",
			},
			wantVarWarning: variables.VariableWarning{
				Commands:        map[string][]string{},
				Components:      map[string][]string{},
				Projects:        map[string][]string{},
				StarterProjects: map[string][]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, gotVarWarning, err := ParseDevfileAndValidate(tt.args.args)
			if err != nil {
				t.Errorf("ParseDevfileAndValidate() error = %v, wantErr nil", err)
				return
			}
			commands, err := gotD.Data.GetCommands(common.DevfileOptions{})
			if err != nil {
				t.Errorf("unexpected error getting commands")
			}
			expectedCommandLine := commands[0].Exec.CommandLine
			if expectedCommandLine != tt.wantCommandLine {
				t.Errorf("command line is %q, should be %q", expectedCommandLine, tt.wantCommandLine)
			}

			getKubeCompOptions := common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.KubernetesComponentType,
				},
			}
			kubeComponents, err := gotD.Data.GetComponents(getKubeCompOptions)
			if err != nil {
				t.Errorf("unexpected error getting kubernetes component")
			}
			kubenetesComponent := kubeComponents[0]

			// check openshift component uri -> inline conversion and value substitution
			getOpenshiftCompOptions := common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.OpenshiftComponentType,
				},
			}
			openshiftComponents, err := gotD.Data.GetComponents(getOpenshiftCompOptions)
			if err != nil {
				t.Errorf("unexpected error getting openshift component")
			}
			openshiftComponent := openshiftComponents[0]

			if tt.args.args.ConvertKubernetesContentInUri == nil || *tt.args.args.ConvertKubernetesContentInUri != false {
				// check kubernetes component uri -> inline conversion and value substitution
				if kubenetesComponent.Kubernetes.Uri != "" || kubenetesComponent.Kubernetes.Inlined == "" ||
					!strings.Contains(kubenetesComponent.Kubernetes.Inlined, tt.wantKubernetesInline) {
					t.Errorf("unexpected kubenetes component inlined, got %s, want include %s", kubenetesComponent.Kubernetes.Inlined, tt.wantKubernetesInline)
				}

				if kubenetesComponent.Attributes != nil {
					if originalUri := kubenetesComponent.Attributes.GetString(K8sLikeComponentOriginalURIKey, &err); err != nil || originalUri != "http://127.0.0.1:8080/outerloop-deploy.yaml" {
						t.Errorf("ParseDevfileAndValidate() should set kubenetesComponent.Attributes, '%s', expected http://127.0.0.1:8080/outerloop-deploy.yaml, got %s",
							K8sLikeComponentOriginalURIKey, originalUri)
					}
				} else {
					t.Error("ParseDevfileAndValidate() should set kubenetesComponent.Attributes, but got empty Attributes")
				}

				// check openshift component uri -> inline conversion and value substitution
				if openshiftComponent.Openshift.Uri != "" || openshiftComponent.Openshift.Inlined == "" ||
					!strings.Contains(openshiftComponent.Openshift.Inlined, tt.wantOpenshiftInline) {
					t.Errorf("unexpected openshift component inlined, got %s, want include %s", openshiftComponent.Openshift.Inlined, tt.wantOpenshiftInline)
				}

				if openshiftComponent.Attributes != nil {
					if originalUri := openshiftComponent.Attributes.GetString(K8sLikeComponentOriginalURIKey, &err); err != nil || originalUri != "http://127.0.0.1:8080/outerloop-service.yaml" {
						t.Errorf("ParseDevfileAndValidate() should set openshiftComponent.Attributes, '%s', expected http://127.0.0.1:8080/outerloop-service.yaml, got %s",
							K8sLikeComponentOriginalURIKey, originalUri)
					}
				} else {
					t.Error("ParseDevfileAndValidate() should set openshiftComponent.Attributes, but got empty Attributes")
				}
			} else {
				if kubenetesComponent.Kubernetes.Uri == "" || kubenetesComponent.Kubernetes.Inlined != "" {
					t.Errorf("unexpected Kubernetes component inlined, got %s, want empty", kubenetesComponent.Kubernetes.Inlined)
				}
				if kubenetesComponent.Attributes != nil {
					t.Errorf("unexpected Kubernetes component attribute, got %v, want empty", kubenetesComponent.Attributes)
				}

				if openshiftComponent.Openshift.Uri == "" || openshiftComponent.Openshift.Inlined != "" {
					t.Errorf("unexpected Openshift component inlined, got %s, want empty", openshiftComponent.Openshift.Inlined)
				}
				if kubenetesComponent.Attributes != nil {
					t.Errorf("unexpected Openshift component attribute, got %v, want empty", openshiftComponent.Attributes)
				}
			}

			getContainerCompOptions := common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{
					ComponentType: v1.ContainerComponentType,
				},
			}

			containerComponents, err := gotD.Data.GetComponents(getContainerCompOptions)
			if err != nil {
				t.Errorf("unexpected error getting container component")
			}
			containerComponent := containerComponents[0]
			dedicatedPod := containerComponent.Container.DedicatedPod
			//check that unset booleans are set to defaults if flattenedDevfile is false
			if tt.args.args.SetBooleanDefaults == nil {
				if tt.args.args.FlattenedDevfile != nil && *tt.args.args.FlattenedDevfile == false {
					if dedicatedPod == nil || dedicatedPod == &trueValue {
						t.Errorf("unset property dedicatedPod is expected to have a default value of false")
					}
				}
			} else {
				if *tt.args.args.SetBooleanDefaults == false {
					if dedicatedPod != nil {
						t.Errorf("unset property dedicatedPod should be set to nil")
					}
				}
			}

			if !reflect.DeepEqual(gotVarWarning, tt.wantVarWarning) {
				t.Errorf("ParseDevfileAndValidate() gotVarWarning = %v, want %v", gotVarWarning, tt.wantVarWarning)
			}
			variables := gotD.Data.GetDevfileWorkspaceSpec().Variables
			if !reflect.DeepEqual(variables, tt.wantVariables) {
				t.Errorf("variables are %+v, expected %+v", variables, tt.wantVariables)
			}
		})
	}
}
