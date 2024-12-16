package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsFsRead(t *testing.T) {
	t.Parallel()

	fs := &OsFS{}
	file, err := fs.ReadFile("not-found.test")
	assert.Empty(t, file)
	require.Error(t, err)
}

func TestOsFsWrite(t *testing.T) {
	t.Parallel()

	fs := &OsFS{}
	require.Error(t, fs.WriteFile("/etc/etc/etc/etc/etc", []byte{}, 0o755))
	require.NoError(t, fs.WriteFile("/tmp/file.txt", []byte{}, 0o755))
}
