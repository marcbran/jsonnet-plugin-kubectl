package kubectl

import (
	"fmt"
	"sort"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func ConfigGetContexts() jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "configGetContexts",
		Params: ast.Identifiers{"options"},
		Func: func(input []any) (any, error) {
			opts, err := parseConfigGetContextsInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			rows, err := runConfigGetContexts(opts)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			return rows, nil
		},
	}
}

type ConfigGetContextsInput struct {
	Kubeconfig string
}

func parseConfigGetContextsInput(input []any) (ConfigGetContextsInput, error) {
	if len(input) != 1 {
		return ConfigGetContextsInput{}, fmt.Errorf("expected options")
	}
	if input[0] == nil {
		return ConfigGetContextsInput{}, nil
	}
	rawOpts, ok := input[0].(map[string]any)
	if !ok {
		return ConfigGetContextsInput{}, fmt.Errorf("options must be an object")
	}
	opts := ConfigGetContextsInput{}
	if s, ok := rawOpts["kubeconfig"].(string); ok {
		opts.Kubeconfig = s
	}
	return opts, nil
}

func runConfigGetContexts(input ConfigGetContextsInput) ([]any, error) {
	cfg, err := loadingRules(input.Kubeconfig).Load()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	rows := make([]any, 0, len(names))
	for _, name := range names {
		ctx := cfg.Contexts[name]
		row := map[string]any{
			"name":      name,
			"current":   name == cfg.CurrentContext,
			"cluster":   ctx.Cluster,
			"authInfo":  ctx.AuthInfo,
			"namespace": ctx.Namespace,
		}
		rows = append(rows, row)
	}
	return rows, nil
}
