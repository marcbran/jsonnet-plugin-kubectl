//go:build e2e

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/marcbran/jsonnet-plugin-kubectl/kubectl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Stage struct {
	t require.TestingT

	tempDir string
	plugin  *jpoet.Plugin
	snippet string

	lastOutput any
	lastErr    error
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	s := &Stage{
		t:       t,
		tempDir: t.TempDir(),
		plugin:  kubectl.Plugin(),
	}
	return s, s, s
}

func (s *Stage) and() *Stage {
	return s
}

func (s *Stage) a_kubeconfig_with_dev_and_prod_contexts() *Stage {
	path := filepath.Join(s.tempDir, "config.yaml")
	kubeconfig := `apiVersion: v1
kind: Config
current-context: prod
contexts:
- name: prod
  context:
    cluster: cluster-prod
    user: user-prod
    namespace: ns-prod
- name: dev
  context:
    cluster: cluster-dev
    user: user-dev
clusters:
- name: cluster-prod
  cluster:
    server: https://prod.example
- name: cluster-dev
  cluster:
    server: https://dev.example
users:
- name: user-prod
  user: {}
- name: user-dev
  user: {}
`
	err := os.WriteFile(path, []byte(kubeconfig), 0o600)
	require.NoError(s.t, err)
	s.snippet = fmt.Sprintf(
		`std.native('invoke:kubectl')('configGetContexts', [{kubeconfig: %q}])`,
		path,
	)
	return s
}

func (s *Stage) a_kubeconfig_without_current_context() *Stage {
	path := filepath.Join(s.tempDir, "no-current-context.yaml")
	kubeconfig := `apiVersion: v1
kind: Config
contexts:
- name: dev
  context:
    cluster: cluster-dev
    user: user-dev
clusters:
- name: cluster-dev
  cluster:
    server: https://dev.example
users:
- name: user-dev
  user: {}
`
	err := os.WriteFile(path, []byte(kubeconfig), 0o600)
	require.NoError(s.t, err)
	s.snippet = fmt.Sprintf(
		`std.native('invoke:kubectl')('configCurrentContext', [{kubeconfig: %q}])`,
		path,
	)
	return s
}

func (s *Stage) a_kubeconfig_with_current_context_set_to(contextName string) *Stage {
	path := filepath.Join(s.tempDir, "current-context.yaml")
	kubeconfig := fmt.Sprintf(`apiVersion: v1
kind: Config
current-context: %s
contexts:
- name: %s
  context:
    cluster: cluster-%s
    user: user-%s
clusters:
- name: cluster-%s
  cluster:
    server: https://%s.example
users:
- name: user-%s
  user: {}
`, contextName, contextName, contextName, contextName, contextName, contextName, contextName)
	err := os.WriteFile(path, []byte(kubeconfig), 0o600)
	require.NoError(s.t, err)
	s.snippet = fmt.Sprintf(
		`std.native('invoke:kubectl')('configCurrentContext', [{kubeconfig: %q}])`,
		path,
	)
	return s
}

func (s *Stage) a_missing_kubeconfig_path() *Stage {
	path := filepath.Join(s.tempDir, "missing-config.yaml")
	s.snippet = fmt.Sprintf(
		`std.native('invoke:kubectl')('configGetContexts', [{kubeconfig: %q}])`,
		path,
	)
	return s
}

func (s *Stage) a_missing_kubeconfig_path_for_current_context() *Stage {
	path := filepath.Join(s.tempDir, "missing-current-context.yaml")
	s.snippet = fmt.Sprintf(
		`std.native('invoke:kubectl')('configCurrentContext', [{kubeconfig: %q}])`,
		path,
	)
	return s
}

func (s *Stage) an_invalid_options_input() *Stage {
	s.snippet = `std.native('invoke:kubectl')('configGetContexts', ['x'])`
	return s
}

func (s *Stage) an_invalid_options_input_for_current_context() *Stage {
	s.snippet = `std.native('invoke:kubectl')('configCurrentContext', ['x'])`
	return s
}

func (s *Stage) config_get_contexts_is_invoked() *Stage {
	s.lastErr = jpoet.Eval(
		jpoet.WithPlugin(s.plugin),
		jpoet.SnippetInput("test.jsonnet", s.snippet),
		jpoet.ValueOutput(&s.lastOutput),
		jpoet.Serialize(false),
	)
	return s
}

func (s *Stage) config_current_context_is_invoked() *Stage {
	s.lastErr = jpoet.Eval(
		jpoet.WithPlugin(s.plugin),
		jpoet.SnippetInput("test.jsonnet", s.snippet),
		jpoet.ValueOutput(&s.lastOutput),
		jpoet.Serialize(false),
	)
	return s
}

func (s *Stage) the_call_has_no_error() *Stage {
	require.NoError(s.t, s.lastErr)
	return s
}

func (s *Stage) the_output_is_a_sorted_context_list() *Stage {
	rows, ok := s.lastOutput.([]any)
	require.True(s.t, ok)
	require.Len(s.t, rows, 2)
	dev, ok := rows[0].(map[string]any)
	require.True(s.t, ok)
	prod, ok := rows[1].(map[string]any)
	require.True(s.t, ok)

	assert.Equal(s.t, "dev", dev["name"])
	assert.Equal(s.t, false, dev["current"])
	assert.Equal(s.t, "cluster-dev", dev["cluster"])
	assert.Equal(s.t, "user-dev", dev["authInfo"])
	assert.Equal(s.t, "", dev["namespace"])

	assert.Equal(s.t, "prod", prod["name"])
	assert.Equal(s.t, true, prod["current"])
	assert.Equal(s.t, "cluster-prod", prod["cluster"])
	assert.Equal(s.t, "user-prod", prod["authInfo"])
	assert.Equal(s.t, "ns-prod", prod["namespace"])
	return s
}

func (s *Stage) the_output_is_a_status_failure_with_code(code float64) *Stage {
	status, ok := s.lastOutput.(map[string]any)
	require.True(s.t, ok)
	assert.Equal(s.t, "Status", status["kind"])
	assert.Equal(s.t, "Failure", status["status"])
	assert.Equal(s.t, code, status["code"])
	return s
}

func (s *Stage) the_output_message_contains(msg string) *Stage {
	status, ok := s.lastOutput.(map[string]any)
	require.True(s.t, ok)
	outMsg, ok := status["message"].(string)
	require.True(s.t, ok)
	assert.Contains(s.t, outMsg, msg)
	return s
}

func (s *Stage) the_output_is_the_current_context(expected string) *Stage {
	out, ok := s.lastOutput.(string)
	require.True(s.t, ok)
	assert.Equal(s.t, expected, out)
	return s
}
