package main

import (
	"os"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // disabled
func TestValidateConfigNoConfigProvided(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd", "validate-config"}
	main()
}

//nolint:paralleltest // disabled
func TestValidateConfigFailedToLoad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd", "validate-config", "--config", "../assets/config-not-found.yml"}
	main()
}

//nolint:paralleltest // disabled
func TestValidateConfigInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd", "validate-config", "--config", "../assets/config-invalid.yml"}
	main()
}

//nolint:paralleltest // disabled
func TestValidateConfigValid(t *testing.T) {
	os.Args = []string{"cmd", "validate-config", "--config", "../assets/config-valid.yml"}
	main()
}

//nolint:paralleltest // disabled
func TestStartNoConfigProvided(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd"}
	main()
}

//nolint:paralleltest // disabled
func TestStartConfigInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd", "--config", "../assets/config-invalid.yml"}
	main()
}

//nolint:paralleltest // disabled
func TestStartConfigProvided(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	os.Args = []string{"cmd", "--config", "../assets/config-valid.yml"}
	main()
}
