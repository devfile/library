package parser

import "github.com/devfile/devfile-parser/pkg/testingutil/filesystem"

// GetFs returns the filesystem object
func (d *DevfileCtx) GetFs() filesystem.Filesystem {
	return d.Fs
}
