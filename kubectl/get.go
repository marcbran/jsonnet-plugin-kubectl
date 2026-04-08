package kubectl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func Get() jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "get",
		Params: ast.Identifiers{"options", "resource", "name"},
		Func: func(input []any) (any, error) {
			opts, res, name, err := parseGetInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			out, err := runGet(context.Background(), opts, res, name)
			if err != nil {
				return nil, err
			}
			return out, nil
		},
	}
}

func parseGetInput(input []any) (map[string]any, string, *string, error) {
	if len(input) != 3 {
		return nil, "", nil, fmt.Errorf("expected options, resource, and name")
	}
	var opts map[string]any
	if input[0] == nil {
		opts = map[string]any{}
	} else {
		m, ok := input[0].(map[string]any)
		if !ok {
			return nil, "", nil, fmt.Errorf("options must be an object")
		}
		opts = m
	}
	res, ok := input[1].(string)
	if !ok || res == "" {
		return nil, "", nil, fmt.Errorf("resource must be a non-empty string")
	}
	if input[2] == nil {
		return opts, res, nil, nil
	}
	s, ok := input[2].(string)
	if !ok {
		return nil, "", nil, fmt.Errorf("name must be a string or null")
	}
	if s == "" {
		return nil, "", nil, fmt.Errorf("name must not be empty")
	}
	return opts, res, &s, nil
}

func clientFailureStatus(code int32, msg string) map[string]any {
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Status",
		"status":     "Failure",
		"code":       float64(code),
		"message":    msg,
		"reason":     "BadRequest",
	}
}

func statusToMap(st *metav1.Status) (map[string]any, error) {
	b, err := json.Marshal(st)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func apiErrToMap(err error) (map[string]any, error) {
	var se *apierrors.StatusError
	if errors.As(err, &se) {
		return statusToMap(&se.ErrStatus)
	}
	return clientFailureStatus(500, err.Error()), nil
}

func restConfigFromOptions(opts map[string]any) (*rest.Config, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if s, ok := opts["context"].(string); ok && s != "" {
		overrides.CurrentContext = s
	}
	if s, ok := opts["namespace"].(string); ok && s != "" {
		overrides.Context.Namespace = s
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
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

func runGet(ctx context.Context, opts map[string]any, resource string, name *string) (map[string]any, error) {
	restCfg, ns, err := restConfigFromOptions(opts)
	if err != nil {
		return clientFailureStatus(400, err.Error()), nil
	}
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return clientFailureStatus(500, err.Error()), nil
	}
	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return clientFailureStatus(500, err.Error()), nil
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return clientFailureStatus(404, err.Error()), nil
	}
	gvk, err := mapper.KindFor(gvr)
	if err != nil {
		return clientFailureStatus(404, err.Error()), nil
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return clientFailureStatus(404, err.Error()), nil
	}
	allNamespaces, _ := opts["allNamespaces"].(bool)
	if allNamespaces {
		ns = ""
	}
	var ri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ri = dyn.Resource(gvr).Namespace(ns)
	} else {
		ri = dyn.Resource(gvr)
	}
	if name != nil {
		obj, err := ri.Get(ctx, *name, metav1.GetOptions{})
		if err != nil {
			return apiErrToMap(err)
		}
		return unstructuredObjectToMap(obj)
	}
	list, err := ri.List(ctx, metav1.ListOptions{})
	if err != nil {
		return apiErrToMap(err)
	}
	return unstructuredListToMap(list)
}

func unstructuredObjectToMap(u *unstructured.Unstructured) (map[string]any, error) {
	b, err := u.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func unstructuredListToMap(u *unstructured.UnstructuredList) (map[string]any, error) {
	b, err := u.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
