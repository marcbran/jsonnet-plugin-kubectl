package kubectl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectEnvIntoArgs(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []any
		env      map[string]string
		want     []any
	}{
		{
			name:     "non-get function is unchanged",
			funcName: "configGetContexts",
			args:     []any{map[string]any{"context": "prod"}},
			env:      map[string]string{"FOO": "bar"},
			want:     []any{map[string]any{"context": "prod"}},
		},
		{
			name:     "empty env leaves args unchanged",
			funcName: "get",
			args:     []any{map[string]any{"context": "prod"}, "pods", nil},
			env:      map[string]string{},
			want:     []any{map[string]any{"context": "prod"}, "pods", nil},
		},
		{
			name:     "nil env leaves args unchanged",
			funcName: "get",
			args:     []any{map[string]any{"context": "prod"}, "pods", nil},
			env:      nil,
			want:     []any{map[string]any{"context": "prod"}, "pods", nil},
		},
		{
			name:     "env is injected into get options",
			funcName: "get",
			args:     []any{map[string]any{"context": "prod"}, "pods", nil},
			env:      map[string]string{"FOO": "bar"},
			want:     []any{map[string]any{"context": "prod", "env": map[string]string{"FOO": "bar"}}, "pods", nil},
		},
		{
			name:     "existing options are preserved alongside env",
			funcName: "get",
			args:     []any{map[string]any{"context": "prod", "namespace": "ns1"}, "pods", nil},
			env:      map[string]string{"FOO": "bar"},
			want:     []any{map[string]any{"context": "prod", "namespace": "ns1", "env": map[string]string{"FOO": "bar"}}, "pods", nil},
		},
		{
			name:     "nil options treated as empty",
			funcName: "get",
			args:     []any{nil, "pods", nil},
			env:      map[string]string{"FOO": "bar"},
			want:     []any{map[string]any{"env": map[string]string{"FOO": "bar"}}, "pods", nil},
		},
		{
			name:     "trailing args beyond options are preserved",
			funcName: "get",
			args:     []any{map[string]any{"context": "dev"}, "nodes", "my-node"},
			env:      map[string]string{"TOKEN": "secret"},
			want:     []any{map[string]any{"context": "dev", "env": map[string]string{"TOKEN": "secret"}}, "nodes", "my-node"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectEnvIntoArgs(tt.funcName, tt.args, tt.env)
			require.Equal(t, tt.want, got)
		})
	}
}
