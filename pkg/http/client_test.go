package http

import (
	"main/assets"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestHttpClientErrorCreating(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	client := NewClient(logger, "querier")
	result := map[string]string{}
	err := client.Get("://test", &result, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "missing protocol scheme")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientNotOkHTTPCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewBytesResponder(503, assets.GetBytesOrPanic("empty.json")),
	)
	logger := loggerPkg.GetNopLogger()
	client := NewClient(logger, "querier")
	result := map[string]string{}
	err := client.Get("https://example.com", &result, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "Could not fetch request. Status code: 503")
}
