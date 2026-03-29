{
  get(options={}, resource, name=null): std.native('invoke:kubectl')('get', [options, resource, name]),
}
