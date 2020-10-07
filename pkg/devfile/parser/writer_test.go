package parser

import (
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
	"github.com/devfile/parser/pkg/testingutil/filesystem"
)

func TestWriteJsonDevfile(t *testing.T) {

	var (
		schemaVersion = "2.0.0"
		testName      = "TestName"
	)

	t.Run("write json devfile", func(t *testing.T) {

		// Use fakeFs
		fs := filesystem.NewFakeFs()

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.FakeContext(fs, OutputDevfileJsonPath),
			Data: &v200.Devfile200{
				v1.Devfile{
					SchemaVersion: schemaVersion,
					Metadata: v1.DevfileMetadata{
						Name: testName,
					},
				},
			},
		}

		// test func()
		err := devfileObj.WriteJsonDevfile()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}

		if _, err := fs.Stat(OutputDevfileJsonPath); err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})

	t.Run("write yaml devfile", func(t *testing.T) {

		// Use fakeFs
		fs := filesystem.NewFakeFs()

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
			Data: &v200.Devfile200{
				v1.Devfile{
					SchemaVersion: schemaVersion,
					Metadata: v1.DevfileMetadata{
						Name: testName,
					},
				},
			},
		}

		// test func()
		err := devfileObj.WriteYamlDevfile()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}

		if _, err := fs.Stat(OutputDevfileYamlPath); err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})
}
