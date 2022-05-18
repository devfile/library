package devfile

import (
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

func TestParseDevfileAndValidate(t *testing.T) {
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
schemaVersion: 2.1.0
`

	devfileContentWithVariable := devfileContent + `variables:
  PARAMS: foo`
	type args struct {
		args parser.ParserArgs
	}
	tests := []struct {
		name            string
		args            args
		wantVarWarning  variables.VariableWarning
		wantCommandLine string
		wantVariables   map[string]string
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

			wantCommandLine: "./main bar",
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

			wantCommandLine: "./main foo",
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
