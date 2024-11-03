package normalize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeString(t *testing.T) {
	t.Parallel()
	require.Equal(t, "abc", NormalizeString("abcабв"))
}
