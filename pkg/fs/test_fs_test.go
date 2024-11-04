package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsRead(t *testing.T) {
	t.Parallel()

	fs := &TestFS{}
	file, err := fs.ReadFile("config-valid.yml")
	assert.NotEmpty(t, file)
	require.NoError(t, err)
}
