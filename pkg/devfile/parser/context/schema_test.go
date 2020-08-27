package parser

import (
	"testing"

	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
)

const (
	validJson200 = `{
		"schemaVersion": "2.0.0",
		"metadata": {
		   "name": "nodejs-stack"
		},
		"projects": [
		   {
			  "name": "project",
			  "git": {
				 "location": "https://github.com/che-samples/web-nodejs-sample.git"
			  }
		   }
		],
		"components": [
		   {
			  "plugin": {
				 "id": "eclipse/che-theia/7.1.0"
			  }
		   },
		   {
			  "plugin": {
				 "id": "eclipse/che-machine-exec-plugin/7.1.0"
			  }
		   },
		   {
			  "plugin": {
				 "name": "typescript-plugin",
				 "id": "che-incubator/typescript/1.30.2",
				 "components": [
					{
					   "container": {
						  "name": "??",
						  "memoryLimit": "512Mi"
					   }
					}
				 ]
			  }
		   },
		   {
			  "container": {
				 "name": "nodejs",
				 "image": "quay.io/eclipse/che-nodejs10-ubi:nightly",
				 "memoryLimit": "512Mi",
				 "endpoints": [
					{
					   "name": "nodejs",
					   "protocol": "http",
					   "targetPort": 3000
					}
				 ],
				 "mountSources": true
			  }
		   }
		],
		"commands": [
		   {
			  "exec": {
				 "id": "download dependencies",
				 "component": "nodejs",
				 "commandLine": "npm install",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "build"
				 }
			  }
		   },
		   {
			  "exec": {
				 "id": "run the app",
				 "component": "nodejs",
				 "commandLine": "nodemon app.js",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "run",
					"isDefault": true
				 }
			  }
		   },
		   {
			  "exec": {
				 "id": "run the app (debugging enabled)",
				 "component": "nodejs",
				 "commandLine": "nodemon --inspect app.js",
				 "workingDir": "${PROJECTS_ROOT}/project/app",
				 "group": {
					"kind": "run"
				 }
			  }
		   },
		   {
			  "exec": {
				 "id": "stop the app",
				 "component": "nodejs",
				 "commandLine": "node_server_pids=$(pgrep -fx '.*nodemon (--inspect )?app.js' | tr \"\\\\n\" \" \") && echo \"Stopping node server with PIDs: ${node_server_pids}\" &&  kill -15 ${node_server_pids} &>/dev/null && echo 'Done.'"
			  }
		   },
		   {
			  "vscodeLaunch": {
				 "id": "Attach remote debugger",
				 "inlined": "{\n  \"version\": \"0.2.0\",\n  \"configurations\": [\n    {\n      \"type\": \"node\",\n      \"request\": \"attach\",\n      \"name\": \"Attach to Remote\",\n      \"address\": \"localhost\",\n      \"port\": 9229,\n      \"localRoot\": \"${workspaceFolder}\",\n      \"remoteRoot\": \"${workspaceFolder}\"\n    }\n  ]\n}\n"
			  }
		   }
		]
	 }`
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
