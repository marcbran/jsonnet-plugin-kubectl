package kubectl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGetInput(t *testing.T) {
	tests := []struct {
		name      string
		input     []any
		wantOpts  map[string]any
		wantRes   string
		wantName  *string
		wantError string
	}{
		{
			name:     "list with empty options",
			input:    []any{map[string]any{}, "pods", nil},
			wantOpts: map[string]any{},
			wantRes:  "pods",
			wantName: nil,
		},
		{
			name:     "list with null options",
			input:    []any{nil, "pods", nil},
			wantOpts: map[string]any{},
			wantRes:  "pods",
			wantName: nil,
		},
		{
			name:  "get with context and namespace",
			input: []any{map[string]any{"context": "prod", "namespace": "ns1"}, "pods", "p1"},
			wantOpts: map[string]any{
				"context":   "prod",
				"namespace": "ns1",
			},
			wantRes:  "pods",
			wantName: strPtr("p1"),
		},
		{
			name:      "wrong arity",
			input:     []any{map[string]any{}, "pods"},
			wantError: "expected options, resource, and name",
		},
		{
			name:      "options not object",
			input:     []any{"x", "pods", nil},
			wantError: "options must be an object",
		},
		{
			name:      "resource not string",
			input:     []any{map[string]any{}, 1, nil},
			wantError: "resource must be a non-empty string",
		},
		{
			name:      "resource empty",
			input:     []any{map[string]any{}, "", nil},
			wantError: "resource must be a non-empty string",
		},
		{
			name:      "name not string",
			input:     []any{map[string]any{}, "pods", 1},
			wantError: "name must be a string or null",
		},
		{
			name:      "name empty string",
			input:     []any{map[string]any{}, "pods", ""},
			wantError: "name must not be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, res, name, err := parseGetInput(tt.input)
			if tt.wantError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantOpts, opts)
			require.Equal(t, tt.wantRes, res)
			if tt.wantName == nil {
				require.Nil(t, name)
				return
			}
			require.NotNil(t, name)
			require.Equal(t, *tt.wantName, *name)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
