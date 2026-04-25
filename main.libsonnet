{
  config: {
    currentContext(options={}): std.native('invoke:kubectl')('configCurrentContext', [options]),
    getContexts(options={}): std.native('invoke:kubectl')('configGetContexts', [options]),
  },
  get(options={}, resource, name=null): std.native('invoke:kubectl')('get', [options, resource, name]),
  apiResources(options={}): std.native('invoke:kubectl')('apiResources', [options]),
  neat: {
    get(options={}, resource, name=null): $.get(options, resource, name) {
      metadata+: {
        managedFields:: [],
      },
    },
  },
}
