//go:build e2e

package tests

import "testing"

func TestConfigGetContextsSuccess(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_kubeconfig_with_dev_and_prod_contexts()

	when.
		config_get_contexts_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_sorted_context_list()
}

func TestConfigGetContextsMissingKubeconfig(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_missing_kubeconfig_path()

	when.
		config_get_contexts_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_status_failure_with_code(400).and().
		the_output_message_contains("no such file")
}

func TestConfigGetContextsInvalidOptions(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_invalid_options_input()

	when.
		config_get_contexts_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_status_failure_with_code(400).and().
		the_output_message_contains("options must be an object")
}

func TestConfigCurrentContextSuccess(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_kubeconfig_with_current_context_set_to("prod")

	when.
		config_current_context_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_the_current_context("prod")
}

func TestConfigCurrentContextMissingKubeconfig(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_missing_kubeconfig_path_for_current_context()

	when.
		config_current_context_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_status_failure_with_code(400).and().
		the_output_message_contains("no such file")
}

func TestConfigCurrentContextUnset(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_kubeconfig_without_current_context()

	when.
		config_current_context_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_status_failure_with_code(400).and().
		the_output_message_contains("current-context is not set")
}

func TestConfigCurrentContextInvalidOptions(t *testing.T) {
	given, when, then := scenario(t)

	given.
		an_invalid_options_input_for_current_context()

	when.
		config_current_context_is_invoked()

	then.
		the_call_has_no_error().and().
		the_output_is_a_status_failure_with_code(400).and().
		the_output_message_contains("options must be an object")
}
