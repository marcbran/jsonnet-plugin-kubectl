local k = import './main.libsonnet';
local p = import 'pkg/main.libsonnet';

p.ex({
  get: p.ex([{
    name: 'list pods in a namespace',
    inputs: [{ context: 'prod', namespace: 'default' }, 'pods', null],
  }, {
    name: 'get one pod by name',
    inputs: [{ context: 'prod', namespace: 'default' }, 'pods', 'my-pod'],
  }, {
    name: 'kubeconfig defaults',
    inputs: [{}, 'nodes', null],
  }]),
})
