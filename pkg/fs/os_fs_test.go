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
