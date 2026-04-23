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
