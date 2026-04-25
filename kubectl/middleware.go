package kubectl

import "github.com/marcbran/jpoet/pkg/jpoet"

func EnvByContext(envFor func(context string) map[string]string) jpoet.Middleware {
	return jpoet.HookMiddleware(func(next jpoet.Invoker, funcName string, args []any) (any, error) {
		opts, _ := args[0].(map[string]any)
		contextName, _ := opts["context"].(string)
		args = injectEnvIntoArgs(funcName, args, envFor(contextName))
		return next.Invoke(funcName, args)
	})
}

func injectEnvIntoArgs(funcName string, args []any, env map[string]string) []any {
	if (funcName != "get" && funcName != "apiResources") || len(args) == 0 || len(env) == 0 {
		return args
	}
	opts, _ := args[0].(map[string]any)
	merged := make(map[string]any, len(opts)+1)
	for k, v := range opts {
		merged[k] = v
	}
	merged["env"] = env
	newArgs := make([]any, len(args))
	copy(newArgs, args)
	newArgs[0] = merged
	return newArgs
}
