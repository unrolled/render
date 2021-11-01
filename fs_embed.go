//go:build go1.16
// +build go1.16

package render

import (
	"embed"
	"io"
	"io/fs"
	"path/filepath"
)

// EmbedFileSystem implements FileSystem on top of an embed.FS
type EmbedFileSystem struct {
	embed.FS
}

var _ FileSystem = &EmbedFileSystem{}

func (e *EmbedFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return fs.WalkDir(e.FS, root, func(path string, d fs.DirEntry, _ error) error {
		if d == nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		return walkFn(path, info, err)
	})
}

type tmplFS struct {
	fs.FS
}

func (tfs tmplFS) Walk(root string, walkFn filepath.WalkFunc) error {
	return fs.WalkDir(tfs, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		return walkFn(path, info, err)
	})
}

func (tfs tmplFS) ReadFile(filename string) ([]byte, error) {
	f, err := tfs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// FS converts io/fs.FS to FileSystem
func FS(oriFS fs.FS) FileSystem {
	return tmplFS{oriFS}
}
