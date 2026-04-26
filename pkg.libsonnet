local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jsonnet-plugin-kubectl',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin/kubectl',
  path: 'plugin/kubectl',
  target: 'kubectl',
}, |||
  Read-only access to Kubernetes resources via client-go, similar to `kubectl get`.
  Requires a valid kubeconfig and cluster reachability when the plugin runs.
|||, {
  get: p.desc(|||
    Fetches one resource by name or lists resources. Pass `name: null` (default) to list.

    `options` may include `context` and `namespace`; omitted values use kubeconfig defaults.

    On success returns the API object or list JSON. On failure returns a Kubernetes `Status` object (`kind: "Status"`).
  |||),
})
