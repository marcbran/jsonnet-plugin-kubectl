package kubectl

import (
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func ConfigCurrentContext() jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "configCurrentContext",
		Params: ast.Identifiers{"options"},
		Func: func(input []any) (any, error) {
			opts, err := parseConfigCurrentContextInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			currentContext, err := runConfigCurrentContext(opts)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			return currentContext, nil
		},
	}
}

type ConfigCurrentContextInput struct {
	Kubeconfig string
}

func parseConfigCurrentContextInput(input []any) (ConfigCurrentContextInput, error) {
	if len(input) != 1 {
		return ConfigCurrentContextInput{}, fmt.Errorf("expected options")
	}
	if input[0] == nil {
		return ConfigCurrentContextInput{}, nil
	}
	rawOpts, ok := input[0].(map[string]any)
	if !ok {
		return ConfigCurrentContextInput{}, fmt.Errorf("options must be an object")
	}
	opts := ConfigCurrentContextInput{}
	if s, ok := rawOpts["kubeconfig"].(string); ok {
		opts.Kubeconfig = s
	}
	return opts, nil
}

func runConfigCurrentContext(input ConfigCurrentContextInput) (string, error) {
	cfg, err := loadingRules(input.Kubeconfig).Load()
	if err != nil {
		return "", err
	}
	if cfg.CurrentContext == "" {
		return "", fmt.Errorf("current-context is not set")
	}
	return cfg.CurrentContext, nil
}
