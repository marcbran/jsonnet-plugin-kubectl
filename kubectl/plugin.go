package kubectl

import (
	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/pkg/jpoet"
)

func Plugin(opts ...jpoet.PluginOption) *jpoet.Plugin {
	return jpoet.NewPlugin("kubectl", []jsonnet.NativeFunction{
		ConfigCurrentContext(),
		ConfigGetContexts(),
		ApiResources(),
		Get(),
	}, opts...)
}
