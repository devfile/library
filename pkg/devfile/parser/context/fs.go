package parser

import "github.com/redhat-developer/devfile-parser/pkg/testingutil/filesystem"

// GetFs returns the filesystem object
func (d *DevfileCtx) GetFs() filesystem.Filesystem {
	return d.Fs
}
