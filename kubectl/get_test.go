package kubectl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGetInput(t *testing.T) {
	tests := []struct {
		name      string
		input     []any
		want      GetInput
		wantError string
	}{
		{
			name:  "list with empty options",
			input: []any{map[string]any{}, "pods", nil},
			want:  GetInput{Opts: GetOptions{}, Res: "pods", Name: nil},
		},
		{
			name:  "list with null options",
			input: []any{nil, "pods", nil},
			want:  GetInput{Opts: GetOptions{}, Res: "pods", Name: nil},
		},
		{
			name:  "get with context and namespace",
			input: []any{map[string]any{"context": "prod", "namespace": "ns1"}, "pods", "p1"},
			want: GetInput{
				Opts: GetOptions{RestOptions: RestOptions{Context: "prod", Namespace: "ns1"}},
				Res:  "pods",
				Name: strPtr("p1"),
			},
		},
		{
			name:  "list with allNamespaces",
			input: []any{map[string]any{"context": "prod", "allNamespaces": true}, "deployments", nil},
			want: GetInput{
				Opts: GetOptions{RestOptions: RestOptions{Context: "prod"}, AllNamespaces: true},
				Res:  "deployments",
				Name: nil,
			},
		},
		{
			name:  "list with labelSelector",
			input: []any{map[string]any{"labelSelector": "app=foo,env=prod"}, "pods", nil},
			want: GetInput{
				Opts: GetOptions{LabelSelector: "app=foo,env=prod"},
				Res:  "pods",
				Name: nil,
			},
		},
		{
			name:  "list with fieldSelector",
			input: []any{map[string]any{"fieldSelector": "status.phase=Running"}, "pods", nil},
			want: GetInput{
				Opts: GetOptions{FieldSelector: "status.phase=Running"},
				Res:  "pods",
				Name: nil,
			},
		},
		{
			name:  "list with kubeconfig",
			input: []any{map[string]any{"kubeconfig": "/home/user/.kube/config"}, "pods", nil},
			want: GetInput{
				Opts: GetOptions{RestOptions: RestOptions{Kubeconfig: "/home/user/.kube/config"}},
				Res:  "pods",
				Name: nil,
			},
		},
		{
			name:  "list with env",
			input: []any{map[string]any{"env": map[string]any{"FOO": "bar", "BAZ": "qux"}}, "pods", nil},
			want: GetInput{
				Opts: GetOptions{RestOptions: RestOptions{Env: map[string]string{"FOO": "bar", "BAZ": "qux"}}},
				Res:  "pods",
				Name: nil,
			},
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
			gi, err := parseGetInput(tt.input)
			if tt.wantError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.Opts, gi.Opts)
			require.Equal(t, tt.want.Res, gi.Res)
			if tt.want.Name == nil {
				require.Nil(t, gi.Name)
				return
			}
			require.NotNil(t, gi.Name)
			require.Equal(t, *tt.want.Name, *gi.Name)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
