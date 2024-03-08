//
// Copyright Red Hat
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
	"strings"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	v2 "github.com/devfile/library/v2/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
)

func TestDevfileObj_WriteYamlDevfile(t *testing.T) {

	var (
		schemaVersion = "2.2.0"
		uri           = "./relative/path/deploy.yaml"
		uri2          = "./relative/path/deploy2.yaml"
	)

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "write devfile with .yaml extension",
			fileName: OutputDevfileYamlPath,
		},
		{
			name:     "write .devfile with .yaml extension",
			fileName: ".devfile.yaml",
		},
		{
			name:     "write devfile with .yml extension",
			fileName: "devfile.yml",
		},
		{
			name:     "write .devfile with .yml extension",
			fileName: ".devfile.yml",
		},
		{
			name:     "write any file, regardless of name and extension",
			fileName: "some-random-file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				// Use fakeFs
				fs          = filesystem.NewFakeFs()
				attributes  = apiAttributes.Attributes{}.PutString(K8sLikeComponentOriginalURIKey, uri)
				attributes2 = apiAttributes.Attributes{}.PutString(K8sLikeComponentOriginalURIKey, uri2)
			)

			// DevfileObj
			devfileObj := DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, tt.fileName),
				Data: &v2.DevfileV2{
					Devfile: v1.Devfile{
						DevfileHeader: devfilepkg.DevfileHeader{
							SchemaVersion: schemaVersion,
							Metadata: devfilepkg.DevfileMetadata{
								Name: tt.name,
							},
						},
						DevWorkspaceTemplateSpec: v1.DevWorkspaceTemplateSpec{
							DevWorkspaceTemplateSpecContent: v1.DevWorkspaceTemplateSpecContent{
								Components: []v1.Component{
									{
										Name:       "kubeComp",
										Attributes: attributes,
										ComponentUnion: v1.ComponentUnion{
											Kubernetes: &v1.KubernetesComponent{
												K8sLikeComponent: v1.K8sLikeComponent{
													K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
														Inlined: "placeholder",
													},
												},
											},
										},
									},
									{
										Name:       "openshiftComp",
										Attributes: attributes2,
										ComponentUnion: v1.ComponentUnion{
											Openshift: &v1.OpenshiftComponent{
												K8sLikeComponent: v1.K8sLikeComponent{
													K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
														Inlined: "placeholder",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			devfileObj.Ctx.SetConvertUriToInlined(true)

			// test func()
			err := devfileObj.WriteYamlDevfile()
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWriteYamlDevfile() unexpected error: '%v', wantErr=%v", err, tt.wantErr)
				return
			}

			if _, err := fs.Stat(tt.fileName); err != nil {
				t.Errorf("TestWriteYamlDevfile() unexpected error: '%v'", err)
			}

			data, err := fs.ReadFile(tt.fileName)
			if err != nil {
				t.Errorf("TestWriteYamlDevfile() unexpected error: '%v'", err)
				return
			}

			content := string(data)
			if strings.Contains(content, "inlined") || strings.Contains(content, K8sLikeComponentOriginalURIKey) {
				t.Errorf("TestWriteYamlDevfile() failed: kubernetes component should not contain inlined or %s", K8sLikeComponentOriginalURIKey)
			}
			if !strings.Contains(content, fmt.Sprintf("uri: %s", uri)) {
				t.Errorf("TestWriteYamlDevfile() failed: kubernetes component does not contain uri")
			}
			if !strings.Contains(content, fmt.Sprintf("uri: %s", uri2)) {
				t.Errorf("TestWriteYamlDevfile() failed: openshift component does not contain uri")
			}
		})
	}
}
