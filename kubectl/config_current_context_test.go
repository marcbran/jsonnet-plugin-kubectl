package kubectl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfigCurrentContextInput(t *testing.T) {
	tests := []struct {
		name      string
		input     []any
		want      ConfigCurrentContextInput
		wantError string
	}{
		{
			name:  "null options",
			input: []any{nil},
			want:  ConfigCurrentContextInput{},
		},
		{
			name:  "empty options",
			input: []any{map[string]any{}},
			want:  ConfigCurrentContextInput{},
		},
		{
			name:  "kubeconfig option",
			input: []any{map[string]any{"kubeconfig": "/home/user/.kube/config"}},
			want:  ConfigCurrentContextInput{Kubeconfig: "/home/user/.kube/config"},
		},

		{
			name:      "wrong arity none",
			input:     []any{},
			wantError: "expected options",
		},
		{
			name:      "wrong arity too many",
			input:     []any{nil, nil},
			wantError: "expected options",
		},
		{
			name:      "options not object",
			input:     []any{"x"},
			wantError: "options must be an object",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConfigCurrentContextInput(tt.input)
			if tt.wantError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
