package parser

import (
	"testing"

	devfileCtx "github.com/devfile/parser/pkg/devfile/parser/context"
	v200 "github.com/devfile/parser/pkg/devfile/parser/data/2.0.0"
	"github.com/devfile/parser/pkg/devfile/parser/data/common"
	"github.com/devfile/parser/pkg/testingutil/filesystem"
)

func TestWriteJsonDevfile(t *testing.T) {

	var (
		devfileTempPath = "devfile.yaml"
		apiVersion      = "2.0.0"
		testName        = "TestName"
	)

	t.Run("write json devfile", func(t *testing.T) {

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
			Data: &v200.Devfile200{
				SchemaVersion: apiVersion,
				Metadata: common.DevfileMetadata{
					Name: testName,
				},
			},
		}

		// Use fakeFs
		fs := filesystem.NewFakeFs()
		devfileObj.Ctx.Fs = fs

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

		// DevfileObj
		devfileObj := DevfileObj{
			Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
			Data: &v200.Devfile200{
				SchemaVersion: apiVersion,
				Metadata: common.DevfileMetadata{
					Name: testName,
				},
			},
		}

		// Use fakeFs
		fs := filesystem.NewFakeFs()
		devfileObj.Ctx.Fs = fs

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
