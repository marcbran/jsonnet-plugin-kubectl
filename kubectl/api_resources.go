package kubectl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
)

type ApiResourcesOptions struct {
	RestOptions
	APIGroup   string
	Namespaced *bool
	Verbs      []string
	Cached     bool
	SortBy     string
	Output     string
	NoHeaders  bool
}

func ApiResources() jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "apiResources",
		Params: ast.Identifiers{"options"},
		Func: func(input []any) (any, error) {
			opts, err := parseApiResourcesInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			out, err := runApiResources(context.Background(), opts)
			if err != nil {
				return nil, err
			}
			return out, nil
		},
	}
}

func parseApiResourcesInput(input []any) (ApiResourcesOptions, error) {
	if len(input) != 1 {
		return ApiResourcesOptions{}, fmt.Errorf("expected options")
	}
	if input[0] == nil {
		return ApiResourcesOptions{}, nil
	}
	rawOpts, ok := input[0].(map[string]any)
	if !ok {
		return ApiResourcesOptions{}, fmt.Errorf("options must be an object")
	}
	opts := ApiResourcesOptions{RestOptions: restOptionsFromMap(rawOpts)}
	if s, ok := rawOpts["apiGroup"].(string); ok {
		opts.APIGroup = s
	}
	if b, ok := rawOpts["namespaced"].(bool); ok {
		opts.Namespaced = &b
	}
	if rawVerbs, ok := rawOpts["verbs"].([]any); ok {
		for _, v := range rawVerbs {
			s, ok := v.(string)
			if !ok {
				return ApiResourcesOptions{}, fmt.Errorf("verbs must be an array of strings")
			}
			opts.Verbs = append(opts.Verbs, s)
		}
	}
	if b, ok := rawOpts["cached"].(bool); ok {
		opts.Cached = b
	}
	if s, ok := rawOpts["sortBy"].(string); ok {
		opts.SortBy = s
	}
	if s, ok := rawOpts["output"].(string); ok {
		opts.Output = s
	}
	if b, ok := rawOpts["noHeaders"].(bool); ok {
		opts.NoHeaders = b
	}
	return opts, nil
}

func runApiResources(_ context.Context, opts ApiResourcesOptions) (any, error) {
	restCfg, _, err := restConfigFromRestOptions(opts.RestOptions)
	if err != nil {
		return clientFailureStatus(400, err.Error()), nil
	}
	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return clientFailureStatus(500, err.Error()), nil
	}
	var disc discovery.DiscoveryInterface = dc
	if opts.Cached {
		disc = memory.NewMemCacheClient(dc)
	}
	lists, err := disc.ServerPreferredResources()
	if err != nil {
		if len(lists) == 0 {
			return clientFailureStatus(500, err.Error()), nil
		}
		var gdf *discovery.ErrGroupDiscoveryFailed
		if !errors.As(err, &gdf) {
			return clientFailureStatus(500, err.Error()), nil
		}
	}
	_ = opts.NoHeaders
	rows, ferr := buildApiResourceRows(lists, opts)
	if ferr != nil {
		return clientFailureStatus(400, ferr.Error()), nil
	}
	return rows, nil
}

func buildApiResourceRows(lists []*metav1.APIResourceList, opts ApiResourcesOptions) ([]any, error) {
	if opts.SortBy != "" && opts.SortBy != "name" && opts.SortBy != "kind" {
		return nil, fmt.Errorf("sortBy must be %q, %q, or empty", "name", "kind")
	}
	if opts.Output != "" && opts.Output != "name" && opts.Output != "wide" {
		return nil, fmt.Errorf("output must be %q, %q, or empty", "name", "wide")
	}
	wide := opts.Output == "wide"
	rows := make([]apiResourceRow, 0, 256)
	for _, list := range lists {
		if list == nil {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for i := range list.APIResources {
			r := &list.APIResources[i]
			if strings.Contains(r.Name, "/") {
				continue
			}
			if !apiGroupFilterMatches(&gv, r, opts.APIGroup) {
				continue
			}
			if opts.Namespaced != nil && r.Namespaced != *opts.Namespaced {
				continue
			}
			if !resourceHasAllVerbs(r.Verbs, opts.Verbs) {
				continue
			}
			g, v, apiVer := groupVersionForResource(list.GroupVersion, r)
			rows = append(rows, apiResourceRow{
				name:       r.Name,
				kind:       r.Kind,
				namespaced: r.Namespaced,
				apiVersion: apiVer,
				group:      g,
				version:    v,
				shortNames: append([]string(nil), r.ShortNames...),
				verbs:      append([]string(nil), r.Verbs...),
				categories: append([]string(nil), r.Categories...),
				wide:       wide,
			})
		}
	}
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "name"
	}
	switch sortBy {
	case "name":
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].name != rows[j].name {
				return rows[i].name < rows[j].name
			}
			return rows[i].apiVersion < rows[j].apiVersion
		})
	case "kind":
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].kind != rows[j].kind {
				return rows[i].kind < rows[j].kind
			}
			if rows[i].name != rows[j].name {
				return rows[i].name < rows[j].name
			}
			return rows[i].apiVersion < rows[j].apiVersion
		})
	}
	out := make([]any, 0, len(rows))
	for _, row := range rows {
		m := map[string]any{
			"name":       row.name,
			"namespaced": row.namespaced,
			"kind":       row.kind,
			"apiVersion": row.apiVersion,
			"shortNames": anyStringSlice(row.shortNames),
		}
		if row.wide {
			m["group"] = row.group
			m["version"] = row.version
			m["verbs"] = anyStringSlice(row.verbs)
			m["categories"] = anyStringSlice(row.categories)
		}
		out = append(out, m)
	}
	return out, nil
}

type apiResourceRow struct {
	name       string
	kind       string
	namespaced bool
	apiVersion string
	group      string
	version    string
	shortNames []string
	verbs      []string
	categories []string
	wide       bool
}

func groupVersionForResource(listGV string, r *metav1.APIResource) (group, version, apiVersion string) {
	g, v := r.Group, r.Version
	if g == "" {
		parsed, err := schema.ParseGroupVersion(listGV)
		if err == nil {
			g, v = parsed.Group, parsed.Version
		}
	}
	if g == "" {
		return g, v, v
	}
	return g, v, schema.GroupVersion{Group: g, Version: v}.String()
}

func apiGroupFilterMatches(gv *schema.GroupVersion, r *metav1.APIResource, filter string) bool {
	if filter == "" {
		return true
	}
	g := r.Group
	if g == "" {
		g = gv.Group
	}
	if filter == "core" && g == "" {
		return true
	}
	return g == filter
}

func resourceHasAllVerbs(verbs, required []string) bool {
	if len(required) == 0 {
		return true
	}
	verbSet := make(map[string]struct{}, len(verbs))
	for _, v := range verbs {
		verbSet[v] = struct{}{}
	}
	for _, v := range required {
		if _, ok := verbSet[v]; !ok {
			return false
		}
	}
	return true
}

func anyStringSlice(s []string) any {
	if len(s) == 0 {
		return nil
	}
	out := make([]any, len(s))
	for i := range s {
		out[i] = s[i]
	}
	return out
}
