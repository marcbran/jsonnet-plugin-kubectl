package kubectl

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type RestOptions struct {
	Context    string
	Kubeconfig string
	Env        map[string]string
	Namespace  string
}

func restOptionsFromMap(raw map[string]any) RestOptions {
	var o RestOptions
	if s, ok := raw["context"].(string); ok {
		o.Context = s
	}
	if s, ok := raw["kubeconfig"].(string); ok {
		o.Kubeconfig = s
	}
	if s, ok := raw["namespace"].(string); ok {
		o.Namespace = s
	}
	switch m := raw["env"].(type) {
	case map[string]string:
		o.Env = m
	case map[string]any:
		o.Env = make(map[string]string, len(m))
		for k, v := range m {
			if s, ok := v.(string); ok {
				o.Env[k] = s
			}
		}
	}
	return o
}

func restConfigFromRestOptions(opts RestOptions) (*rest.Config, string, error) {
	rawCfg, err := loadingRules(opts.Kubeconfig).Load()
	if err != nil {
		return nil, "", err
	}
	contextName := opts.Context
	if contextName == "" {
		contextName = rawCfg.CurrentContext
	}
	if len(opts.Env) > 0 {
		err = injectExecEnv(rawCfg, contextName, opts.Env)
		if err != nil {
			return nil, "", err
		}
	}
	overrides := &clientcmd.ConfigOverrides{}
	if opts.Context != "" {
		overrides.CurrentContext = opts.Context
	}
	if opts.Namespace != "" {
		overrides.Context.Namespace = opts.Namespace
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*rawCfg, overrides)
	restCfg, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", err
	}
	ns, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", err
	}
	return restCfg, ns, nil
}

func injectExecEnv(cfg *clientcmdapi.Config, contextName string, env map[string]string) error {
	ctx, ok := cfg.Contexts[contextName]
	if !ok {
		return fmt.Errorf("context %q not found", contextName)
	}
	authInfo, ok := cfg.AuthInfos[ctx.AuthInfo]
	if !ok || authInfo.Exec == nil {
		return nil
	}
	for k, v := range env {
		authInfo.Exec.Env = append(authInfo.Exec.Env, clientcmdapi.ExecEnvVar{Name: k, Value: v})
	}
	return nil
}
