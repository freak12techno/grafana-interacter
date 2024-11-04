package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigInvalidTimezone(t *testing.T) {
	t.Parallel()

	config := &Config{Timezone: "invalid"}
	err := config.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "error parsing timezone")
}

func TestLoadConfigOk(t *testing.T) {
	t.Parallel()

	config := &Config{Timezone: "Etc/GMT"}
	err := config.Validate()
	require.NoError(t, err)
}
