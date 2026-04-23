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

type GetInput struct {
	Opts GetOptions
	Res  string
	Name *string
}

type GetOptions struct {
	Context       string
	Namespace     string
	AllNamespaces bool
	LabelSelector string
	FieldSelector string
	Kubeconfig    string
}

func Get() jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "get",
		Params: ast.Identifiers{"options", "resource", "name"},
		Func: func(input []any) (any, error) {
			gi, err := parseGetInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			out, err := runGet(context.Background(), gi)
			if err != nil {
				return nil, err
			}
			return out, nil
		},
	}
}

func parseGetInput(input []any) (GetInput, error) {
	if len(input) != 3 {
		return GetInput{}, fmt.Errorf("expected options, resource, and name")
	}
	var rawOpts map[string]any
	if input[0] == nil {
		rawOpts = map[string]any{}
	} else {
		m, ok := input[0].(map[string]any)
		if !ok {
			return GetInput{}, fmt.Errorf("options must be an object")
		}
		rawOpts = m
	}
	res, ok := input[1].(string)
	if !ok || res == "" {
		return GetInput{}, fmt.Errorf("resource must be a non-empty string")
	}
	var name *string
	if input[2] != nil {
		s, ok := input[2].(string)
		if !ok {
			return GetInput{}, fmt.Errorf("name must be a string or null")
		}
		if s == "" {
			return GetInput{}, fmt.Errorf("name must not be empty")
		}
		name = &s
	}
	opts := GetOptions{}
	if s, ok := rawOpts["context"].(string); ok {
		opts.Context = s
	}
	if s, ok := rawOpts["namespace"].(string); ok {
		opts.Namespace = s
	}
	if b, ok := rawOpts["allNamespaces"].(bool); ok {
		opts.AllNamespaces = b
	}
	if s, ok := rawOpts["labelSelector"].(string); ok {
		opts.LabelSelector = s
	}
	if s, ok := rawOpts["fieldSelector"].(string); ok {
		opts.FieldSelector = s
	}
	if s, ok := rawOpts["kubeconfig"].(string); ok {
		opts.Kubeconfig = s
	}
	return GetInput{Opts: opts, Res: res, Name: name}, nil
}

func runGet(ctx context.Context, gi GetInput) (map[string]any, error) {
	restCfg, ns, err := restConfigFromOptions(gi.Opts)
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
	gvr, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: gi.Res})
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
	if gi.Opts.AllNamespaces {
		ns = ""
	}
	var ri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ri = dyn.Resource(gvr).Namespace(ns)
	} else {
		ri = dyn.Resource(gvr)
	}
	if gi.Name != nil {
		obj, err := ri.Get(ctx, *gi.Name, metav1.GetOptions{})
		if err != nil {
			return apiErrToMap(err)
		}
		return unstructuredObjectToMap(obj)
	}
	list, err := ri.List(ctx, metav1.ListOptions{
		LabelSelector: gi.Opts.LabelSelector,
		FieldSelector: gi.Opts.FieldSelector,
	})
	if err != nil {
		return apiErrToMap(err)
	}
	return unstructuredListToMap(list)
}

func restConfigFromOptions(opts GetOptions) (*rest.Config, string, error) {
	cfgLoadingRules := loadingRules(opts.Kubeconfig)
	overrides := &clientcmd.ConfigOverrides{}
	if opts.Context != "" {
		overrides.CurrentContext = opts.Context
	}
	if opts.Namespace != "" {
		overrides.Context.Namespace = opts.Namespace
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(cfgLoadingRules, overrides)
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

func apiErrToMap(err error) (map[string]any, error) {
	var se *apierrors.StatusError
	if errors.As(err, &se) {
		return statusToMap(&se.ErrStatus)
	}
	return clientFailureStatus(500, err.Error()), nil
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

func clientFailureStatus(code int32, msg string) map[string]any {
	var reason metav1.StatusReason
	switch code {
	case 400:
		reason = metav1.StatusReasonBadRequest
	case 404:
		reason = metav1.StatusReasonNotFound
	default:
		reason = metav1.StatusReasonInternalError
	}
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Status",
		"status":     "Failure",
		"code":       float64(code),
		"message":    msg,
		"reason":     string(reason),
	}
}
