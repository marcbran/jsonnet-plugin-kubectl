{
  get(options={}, resource, name=null): std.native('invoke:kubectl')('get', [options, resource, name]),
  neat: {
    get(options={}, resource, name=null): $.get(options, resource, name) {
      metadata+: {
        managedFields:: [],
      },
    },
  },
}
