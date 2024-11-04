package pkg

import (
	"main/pkg/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadFail(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}
	LoadConfig(filesystem, "not-existing.yml")
}

func TestLoadConfigInvalidYaml(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}
	LoadConfig(filesystem, "invalid.yml")
}

func TestLoadConfigValid(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}
	config := LoadConfig(filesystem, "config-valid.yml")
	require.NotNil(t, config)
}
