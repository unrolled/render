package render

import (
	"io/ioutil"
	"path/filepath"
)

type FileSystem interface {
	Walk(root string, walkFn filepath.WalkFunc) error
	ReadFile(filename string) ([]byte, error)
}

type osFileSystem struct{}

func (osFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (osFileSystem) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
