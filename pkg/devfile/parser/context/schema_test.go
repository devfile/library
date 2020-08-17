package parser

import (
	"testing"

	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
)

const (
	validJson200 = `{"apiVersion":"2.0.0","metadata":{"name":"java-web-spring"},"projects":[{"name":"java-web-spring","source":{"type":"git","location":"https://github.com/spring-projects/spring-petclinic.git"}}],"components":[{"type":"chePlugin","id":"redhat/java/latest","memoryLimit":"1512Mi"},{"alias":"tools","type":"dockerimage","image":"quay.io/eclipse/che-java8-maven:nightly","memoryLimit":"768Mi"}],"commands":[{"actions":[{"command":"mvn clean install","component":"tools","type":"build","workdir":"${CHE_PROJECTS_ROOT}/java-web-spring"}],"name":"maven build"},{"actions":[{"command":"java -jar -Xdebug -Xrunjdwp:transport=dt_socket,server=y,suspend=n,address=5005 \\\ntarget/*.jar\n","component":"tools","type":"run","workdir":"${CHE_PROJECTS_ROOT}/java-web-spring"}],"name":"run webapp"}]}`
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
