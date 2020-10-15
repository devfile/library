package parser

import (
	"testing"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/pkg/devfile"
	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v2 "github.com/devfile/parser/pkg/devfile/parser/data/v2"
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
			Data: &v2.DevfileV2{
				Devfile: v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: schemaVersion,
						Metadata: devfilepkg.DevfileMetadata{
							Name: testName,
						},
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
			Data: &v2.DevfileV2{
				Devfile: v1.Devfile{
					DevfileHeader: devfilepkg.DevfileHeader{
						SchemaVersion: schemaVersion,
						Metadata: devfilepkg.DevfileMetadata{
							Name: testName,
						},
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
