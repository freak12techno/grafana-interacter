package fs

import (
	"main/assets"
	"os"
)

type TestFS struct{}

func (fs *TestFS) ReadFile(name string) ([]byte, error) {
	return assets.EmbedFS.ReadFile(name)
}

func (fs *TestFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return nil
}
