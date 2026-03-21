local lib = {
  flatten(pkg)::
    if std.isFunction(pkg) then self.flatten(pkg()) else pkg,

  scan(obj)::
    local aux(old, key) =
      if std.startsWith(key, '#') then
        true
      else if std.isObject(obj[key]) then
        old || $.scan(obj[key])
      else old;
    std.foldl(aux, std.objectFieldsAll(obj), false),

  load(pkg)::
    local obj = self.flatten(pkg);
    local aux(old, key) =
      if !std.isObject(obj[key]) then
        old
      else if std.objectHasAll(obj, '#' + key) && obj['#' + key] == 'ignore' then
        old
      else if std.startsWith(key, '#') then
        old { [key]: obj[key] }
      else if self.scan(obj[key]) then
        old { [key]: $.load(obj[key]) }
      else old;

    std.foldl(aux, if !std.isObject(obj) then [] else std.objectFieldsAll(obj), {}),
};


lib.load(std.extVar('main'))
