package parser

import (
	"testing"

	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
)

const (
	validJson200 = `{ "schemaVersion": "2.0.0", "metadata": { "name": "nodejs", "version": "1.0.0", "alpha.build-dockerfile": "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/build/Dockerfile", "alpha.deployment-manifest": "https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/deploy/deployment-manifest.yaml" }, "projects": [ { "name": "nodejs-starter", "git": { "location": "https://github.com/odo-devfiles/nodejs-ex.git" } } ], "components": [ { "container": { "name": "runtime", "image": "registry.access.redhat.com/ubi8/nodejs-12:1-45", "memoryLimit": "1024Mi", "mountSources": true, "sourceMapping": "/project", "endpoints": [ { "name": "http-3000", "targetPort": 3000 } ] } } ], "commands": [ { "exec": { "id": "install", "component": "runtime", "commandLine": "npm install", "workingDir": "/project", "group": { "kind": "build", "isDefault": true } } }, { "exec": { "id": "run", "component": "runtime", "commandLine": "npm start", "workingDir": "/project", "group": { "kind": "run", "isDefault": true } } }, { "exec": { "id": "debug", "component": "runtime", "commandLine": "npm run debug", "workingDir": "/project", "group": { "kind": "debug", "isDefault": true } } }, { "exec": { "id": "test", "component": "runtime", "commandLine": "npm test", "workingDir": "/project", "group": { "kind": "test", "isDefault": true } } } ] }`
)

func TestValidateDevfileSchema(t *testing.T) {

	t.Run("valid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: validJsonRawContent200(),
			}
		)

		err := d.ValidateDevfileSchema()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})

	t.Run("invalid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: []byte("{}"),
			}
		)

		err := d.ValidateDevfileSchema()
		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})
}

func validJsonRawContent200() []byte {
	return []byte(validJson200)
}
