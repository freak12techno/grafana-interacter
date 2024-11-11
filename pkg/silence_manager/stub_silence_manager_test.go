package silence_manager

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStubSilenceManager(t *testing.T) {
	t.Parallel()

	silenceManager := NewStubSilenceManager()
	require.NotEmpty(t, silenceManager.Prefixes())
	require.NotEmpty(t, silenceManager.Name())
	require.NoError(t, silenceManager.DeleteSilence("123"))
	require.True(t, silenceManager.Enabled())
	require.Empty(t, silenceManager.GetMutesDurations())

	silence, err := silenceManager.GetSilence("123")
	require.Error(t, err)
	require.NotNil(t, silence)
}
