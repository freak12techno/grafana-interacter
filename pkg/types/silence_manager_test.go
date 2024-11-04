package types

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetSilencesWithAlertsFetchSilencesError(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	manager.GetSilencesError = errors.New("custom error")

	silences, err := GetSilencesWithAlerts(manager)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silences)
}

func TestGetSilencesWithAlertsFetchSilenceError(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	manager.GetSilenceMatchingAlertsError = errors.New("custom error")
	_, err := manager.CreateSilence(Silence{
		ID:     "silence",
		EndsAt: time.Now().Add(time.Hour),
		Matchers: SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)

	silences, err := GetSilencesWithAlerts(manager)
	require.Error(t, err)
	require.ErrorContains(t, err, "Error getting alerts for silence on 1 silences!")
	require.Empty(t, silences)
}

func TestGetSilencesWithAlertsOk(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	_, err := manager.CreateSilence(Silence{
		ID:     "silence",
		EndsAt: time.Now().Add(time.Hour),
		Matchers: SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)

	silences, err := GetSilencesWithAlerts(manager)
	require.NoError(t, err)
	require.NotEmpty(t, silences)
}
