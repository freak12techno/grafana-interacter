package silence_manager

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/types"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestAlertmanagerBasic(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "http://localhost:9090", MutesDurations: []string{"1h"}}
	client := InitAlertmanager(config, logger)

	require.True(t, client.Enabled())
	require.Equal(t, "Alertmanager", client.Name())
	require.Equal(t, constants.AlertmanagerSilencePrefix, client.GetSilencePrefix())
	require.Equal(t, constants.AlertmanagerUnsilencePrefix, client.GetUnsilencePrefix())
	require.Equal(t, constants.AlertmanagerPaginatedSilencesList, client.GetPaginatedSilencesListPrefix())
	require.Equal(t, []string{"1h"}, client.GetMutesDurations())
}

//nolint:paralleltest
func TestAlertmanagerGetSilencesFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	silences, err := client.GetSilences()
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silences)
}

//nolint:paralleltest
func TestAlertmanagerGetSilencesOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com", User: "admin", Password: "admin"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silences-ok.json")))

	silences, err := client.GetSilences()
	require.NoError(t, err)
	require.NotEmpty(t, silences)
}

//nolint:paralleltest
func TestAlertmanagerGetSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/silence/id",
		httpmock.NewErrorResponder(errors.New("custom error")))

	silence, err := client.GetSilence("id")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, silence)
}

//nolint:paralleltest
func TestAlertmanagerGetSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/silence/id",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-silence-ok.json")))

	silence, err := client.GetSilence("id")
	require.NoError(t, err)
	require.NotEmpty(t, silence)
}

//nolint:paralleltest
func TestAlertmanagerCreateSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/v2/silences",
		httpmock.NewErrorResponder(errors.New("custom error")))

	_, err := client.CreateSilence(types.Silence{})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest
func TestAlertmanagerCreateSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/api/v2/silences",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-create-silence-ok.json")))

	silence, err := client.CreateSilence(types.Silence{ID: "test"})
	require.NoError(t, err)
	require.NotEmpty(t, silence)
}

//nolint:paralleltest
func TestAlertmanagerDeleteSilenceFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/v2/silence/id",
		httpmock.NewErrorResponder(errors.New("custom error")))

	err := client.DeleteSilence("id")
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest
func TestAlertmanagerDeleteSilenceOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/api/v2/silence/id",
		httpmock.NewBytesResponder(200, []byte{}))

	err := client.DeleteSilence("id")
	require.NoError(t, err)
}

//nolint:paralleltest
func TestAlertmanagerListSilenceAlertsFailed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/alerts?filter=key%3D%22value%22&silenced=true&inhibited=true&active=true",
		httpmock.NewErrorResponder(errors.New("custom error")))

	alerts, err := client.GetSilenceMatchingAlerts(types.Silence{
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Empty(t, alerts)
}

//nolint:paralleltest
func TestAlertmanagerListSilenceAlertsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	config := &configPkg.AlertmanagerConfig{URL: "https://example.com"}
	client := InitAlertmanager(config, logger)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/api/v2/alerts?filter=key%3D%22value%22&silenced=true&inhibited=true&active=true",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("alertmanager-alerts.json")))

	alerts, err := client.GetSilenceMatchingAlerts(types.Silence{
		Matchers: types.SilenceMatchers{
			{IsEqual: true, IsRegex: false, Name: "key", Value: "value"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, alerts)
}
