// Package app_test exercises the core application logic.
//
// # Running tests
//
//	make test          # run all tests with the race detector
//	make test-cover    # run tests and open the HTML coverage report
//
// # Table-driven pattern
//
// All tests that exercise multiple cases follow the standard Go table-driven
// pattern:
//
//	testCases := []struct {
//	    name     string
//	    input    T
//	    expected T
//	}{...}
//	for _, tc := range testCases {
//	    t.Run(tc.name, func(t *testing.T) { ... })
//	}
//
// Each case is a named sub-test so that `go test -run TestFoo/case_name`
// can target it individually.
package app_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wallfacer/oblivion/internal/app"
)

// TestRun_Smoke verifies that Run completes without error when given a valid
// Config and that it writes at least some output.
func TestRun_Smoke(t *testing.T) {
	var buf bytes.Buffer
	err := app.Run(app.Config{Output: &buf})
	require.NoError(t, err, "Run should not return an error with a valid config")
	assert.NotEmpty(t, buf.String(), "Run should write output")
}

// TestRun_NilOutput verifies that Run returns an error when Output is nil,
// preventing a nil-pointer panic at runtime.
func TestRun_NilOutput(t *testing.T) {
	err := app.Run(app.Config{Output: nil})
	require.Error(t, err, "Run must return an error when Output is nil")
}

// TestRun_OutputContent uses a table-driven pattern to verify the content
// written by Run under different Config values.
//
// Add new rows here whenever behaviour changes; no other code is needed.
func TestRun_OutputContent(t *testing.T) {
	testCases := []struct {
		name           string
		cfg            app.Config
		containsAll    []string
		wantErr        bool
	}{
		{
			name:        "default output contains program name",
			cfg:         app.Config{Output: new(bytes.Buffer)},
			containsAll: []string{"oblivion"},
			wantErr:     false,
		},
		{
			name:        "verbose flag does not cause error",
			cfg:         app.Config{Output: new(bytes.Buffer), Verbose: true},
			containsAll: []string{"oblivion"},
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Re-use the buffer stored in cfg so we can inspect output.
			buf, ok := tc.cfg.Output.(*bytes.Buffer)
			require.True(t, ok, "test bug: Output must be *bytes.Buffer")

			err := app.Run(tc.cfg)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			output := buf.String()
			for _, want := range tc.containsAll {
				assert.True(t,
					strings.Contains(output, want),
					"output %q should contain %q", output, want,
				)
			}
		})
	}
}
