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
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadAndParseKubernetesYaml(t *testing.T) {
	const serverIP = "127.0.0.1:9080"
	var data []byte

	fs := afero.Afero{Fs: afero.NewOsFs()}
	absPath, err := util.GetAbsPath("../../../tests/yamls/resources.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	data, err = fs.ReadFile(absPath)
	if err != nil {
		t.Error(err)
		return
	}

	// Mocking the YAML file endpoint on a very basic level
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write(data)
		if err != nil {
			t.Errorf("Unexpected error while writing data: %v", err)
		}
	}))
	// create a listener with the desired port.
	l, err := net.Listen("tcp", serverIP)
	if err != nil {
		t.Errorf("Unexpected error while creating listener: %v", err)
		return
	}

	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	testServer.Listener.Close()
	testServer.Listener = l

	testServer.Start()
	defer testServer.Close()

	badData := append(data, 59)

	tests := []struct {
		name                string
		src                 YamlSrc
		fs                  afero.Afero
		wantErr             bool
		wantDeploymentNames []string
		wantServiceNames    []string
		wantRouteNames      []string
		wantIngressNames    []string
		wantOtherNames      []string
	}{
		{
			name: "Read the YAML from the URL",
			src: YamlSrc{
				URL: "http://" + serverIP,
			},
			fs:                  fs,
			wantDeploymentNames: []string{"deploy-sample", "deploy-sample-2"},
			wantServiceNames:    []string{"service-sample", "service-sample-2"},
			wantRouteNames:      []string{"route-sample", "route-sample-2"},
			wantIngressNames:    []string{"ingress-sample", "ingress-sample-2"},
			wantOtherNames:      []string{"pvc-sample", "pvc-sample-2"},
		},
		{
			name: "Read the YAML from the Path",
			src: YamlSrc{
				Path: "../../../tests/yamls/resources.yaml",
			},
			fs:                  fs,
			wantDeploymentNames: []string{"deploy-sample", "deploy-sample-2"},
			wantServiceNames:    []string{"service-sample", "service-sample-2"},
			wantRouteNames:      []string{"route-sample", "route-sample-2"},
			wantIngressNames:    []string{"ingress-sample", "ingress-sample-2"},
			wantOtherNames:      []string{"pvc-sample", "pvc-sample-2"},
		},
		{
			name: "Read the YAML from the Data",
			src: YamlSrc{
				Data: data,
			},
			fs:                  fs,
			wantDeploymentNames: []string{"deploy-sample", "deploy-sample-2"},
			wantServiceNames:    []string{"service-sample", "service-sample-2"},
			wantRouteNames:      []string{"route-sample", "route-sample-2"},
			wantIngressNames:    []string{"ingress-sample", "ingress-sample-2"},
			wantOtherNames:      []string{"pvc-sample", "pvc-sample-2"},
		},
		{
			name: "Bad URL",
			src: YamlSrc{
				URL: "http://badurl",
			},
			fs:      fs,
			wantErr: true,
		},
		{
			name: "Bad Path",
			src: YamlSrc{
				Path: "$%^&",
			},
			fs:      fs,
			wantErr: true,
		},
		{
			name: "Bad Data",
			src: YamlSrc{
				Data: badData,
			},
			fs:      fs,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := ReadKubernetesYaml(tt.src, tt.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for _, value := range values {
				kubernetesMap := value.(map[string]interface{})

				kind := kubernetesMap["kind"]
				metadataMap := kubernetesMap["metadata"].(map[string]interface{})
				name := metadataMap["name"]

				switch kind {
				case "Deployment":
					assert.Contains(t, tt.wantDeploymentNames, name)
				case "Service":
					assert.Contains(t, tt.wantServiceNames, name)
				case "Route":
					assert.Contains(t, tt.wantRouteNames, name)
				case "Ingress":
					assert.Contains(t, tt.wantIngressNames, name)
				default:
					assert.Contains(t, tt.wantOtherNames, name)
				}
			}

			if len(values) > 0 {
				resources, err := ParseKubernetesYaml(values)
				if err != nil {
					t.Error(err)
					return
				}

				if reflect.DeepEqual(resources, KubernetesResources{}) {
					t.Error("Kubernetes resources is empty, expected to contain some resources")
				} else {
					deployments := resources.Deployments
					services := resources.Services
					routes := resources.Routes
					ingresses := resources.Ingresses

					for _, deploy := range deployments {
						assert.Contains(t, tt.wantDeploymentNames, deploy.Name)
					}
					for _, svc := range services {
						assert.Contains(t, tt.wantServiceNames, svc.Name)
					}
					for _, route := range routes {
						assert.Contains(t, tt.wantRouteNames, route.Name)
					}
					for _, ingress := range ingresses {
						assert.Contains(t, tt.wantIngressNames, ingress.Name)
					}
				}
			}
		})
	}
}
