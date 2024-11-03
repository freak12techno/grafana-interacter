package normalize

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizeString(t *testing.T) {
	t.Parallel()
	require.Equal(t, "abc", NormalizeString("abcабв"))
}
