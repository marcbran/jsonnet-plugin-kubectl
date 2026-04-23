package kubectl

import "k8s.io/client-go/tools/clientcmd"

func loadingRules(kubeconfigPath string) *clientcmd.ClientConfigLoadingRules {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		rules.ExplicitPath = kubeconfigPath
	}
	return rules
}
