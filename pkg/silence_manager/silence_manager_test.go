package silence_manager

import (
	"errors"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetSilencesWithAlertsFetchSilencesError(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	manager.GetSilencesError = errors.New("custom error")

	silences, _, _, err := GetSilencesWithAlerts(manager, 0, 100)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silences)
}

func TestGetSilencesWithAlertsFetchSilenceError(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	manager.GetSilenceMatchingAlertsError = errors.New("custom error")
	_, err := manager.CreateSilence(types.Silence{
		ID:     "silence",
		EndsAt: time.Now().Add(time.Hour),
		Status: types.SilenceStatus{State: "active"},
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)

	silences, _, _, err := GetSilencesWithAlerts(manager, 0, 100)
	require.Error(t, err)
	require.ErrorContains(t, err, "Error getting alerts for silence on 1 silences!")
	require.Empty(t, silences)
}

func TestGetSilencesWithAlertsOk(t *testing.T) {
	t.Parallel()

	manager := NewStubSilenceManager()
	_, err := manager.CreateSilence(types.Silence{
		ID:     "silence",
		EndsAt: time.Now().Add(time.Hour),
		Status: types.SilenceStatus{State: "active"},
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)

	silences, _, _, err := GetSilencesWithAlerts(manager, 0, 100)
	require.NoError(t, err)
	require.NotEmpty(t, silences)
}
