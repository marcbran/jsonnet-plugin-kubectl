package kubectl

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseApiResourcesInput(t *testing.T) {
	nsTrue := true
	tests := []struct {
		name      string
		input     []any
		want      ApiResourcesOptions
		wantError string
	}{
		{
			name:  "empty options",
			input: []any{map[string]any{}},
			want:  ApiResourcesOptions{},
		},
		{
			name:  "null options",
			input: []any{nil},
			want:  ApiResourcesOptions{},
		},
		{
			name:  "full options",
			input: []any{map[string]any{
				"context":     "prod",
				"kubeconfig":  "/a/config",
				"env":         map[string]any{"A": "b"},
				"apiGroup":    "apps",
				"namespaced":  true,
				"verbs":       []any{"get", "list"},
				"cached":      true,
				"sortBy":      "kind",
				"output":      "wide",
				"noHeaders":   true,
			}},
			want: ApiResourcesOptions{
				RestOptions: RestOptions{
					Context:    "prod",
					Kubeconfig: "/a/config",
					Env:        map[string]string{"A": "b"},
				},
				APIGroup:   "apps",
				Namespaced: &nsTrue,
				Verbs:      []string{"get", "list"},
				Cached:     true,
				SortBy:     "kind",
				Output:     "wide",
				NoHeaders:  true,
			},
		},
		{
			name:      "wrong arity",
			input:     []any{},
			wantError: "expected options",
		},
		{
			name:      "options not object",
			input:     []any{"x"},
			wantError: "options must be an object",
		},
		{
			name:      "verbs non-strings",
			input:     []any{map[string]any{"verbs": []any{"get", 1}}},
			wantError: "verbs must be an array of strings",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseApiResourcesInput(tt.input)
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

func TestBuildApiResourceRows(t *testing.T) {
	lists := []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "pods",
					SingularName: "pod",
					Namespaced:   true,
					Kind:         "Pod",
					Verbs:        []string{"get", "list", "watch", "create"},
					ShortNames:   []string{"po"},
				},
			},
		},
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{
					Group:        "apps",
					Version:      "v1",
					Name:         "deployments",
					Namespaced:   true,
					Kind:         "Deployment",
					Verbs:        []string{"get", "list", "delete"},
					ShortNames:   []string{"deploy"},
					Categories:   []string{"all"},
				},
				{Name: "deployments/scale", Kind: "Scale", Verbs: []string{"get", "update"}},
			},
		},
	}
	t.Run("filter subresources and apiGroup", func(t *testing.T) {
		out, err := buildApiResourceRows(lists, ApiResourcesOptions{APIGroup: "apps", SortBy: "name"})
		require.NoError(t, err)
		require.Len(t, out, 1)
		m := out[0].(map[string]any)
		require.Equal(t, "deployments", m["name"])
	})
	t.Run("verbs filter", func(t *testing.T) {
		out, err := buildApiResourceRows(lists, ApiResourcesOptions{Verbs: []string{"get", "list", "watch"}})
		require.NoError(t, err)
		names := make([]string, 0, len(out))
		for _, row := range out {
			m := row.(map[string]any)
			names = append(names, m["name"].(string))
		}
		require.Contains(t, names, "pods")
		require.NotContains(t, names, "deployments")
	})
	t.Run("namespaced filter", func(t *testing.T) {
		f := false
		out, err := buildApiResourceRows([]*metav1.APIResourceList{lists[0]}, ApiResourcesOptions{Namespaced: &f})
		require.NoError(t, err)
		require.Empty(t, out)
	})
	t.Run("output wide", func(t *testing.T) {
		out, err := buildApiResourceRows(lists, ApiResourcesOptions{APIGroup: "apps", Output: "wide"})
		require.NoError(t, err)
		m := out[0].(map[string]any)
		require.NotNil(t, m["verbs"])
		require.NotNil(t, m["categories"])
	})
	t.Run("sort by kind", func(t *testing.T) {
		out, err := buildApiResourceRows(lists, ApiResourcesOptions{SortBy: "kind"})
		require.NoError(t, err)
		kinds := make([]string, 0, len(out))
		for _, row := range out {
			kinds = append(kinds, row.(map[string]any)["kind"].(string))
		}
		require.Equal(t, []string{"Deployment", "Pod"}, kinds)
	})
	t.Run("invalid sortBy", func(t *testing.T) {
		_, err := buildApiResourceRows(lists, ApiResourcesOptions{SortBy: "nope"})
		require.Error(t, err)
	})
	t.Run("invalid output", func(t *testing.T) {
		_, err := buildApiResourceRows(lists, ApiResourcesOptions{Output: "json"})
		require.Error(t, err)
	})
}
