import json, jsonpickle

class Jsonify:
  def __init__(self, process):
    self.process = process

  def visit(self, node):
    meth = None
    for cls in node.__class__.__mro__:
      meth_name = 'visit_' + cls.__name__
      meth = getattr(self, meth_name, None)
      if meth: break

    if not meth:
      meth_name = "generic_" + self.process + '_visit'
      meth = getattr(self, meth_name, None)

    return meth(node)

  def generic_encode_visit(self, node):
    if hasattr(node, '__module__') and node.__module__ != "__builtin__":
      return jsonpickle.encode(node)
    return node

  def generic_decode_visit(self, node):
    return node

  def visit_unicode(self, node):
    if self.process == "decode" and "py/object" in node:
      return jsonpickle.decode(node)
    return str(node) # because fuck Unicode

  def visit_list(self, node):
    out = [None] * len(node)
    for i in xrange(len(node)):
      out[i] = self.visit(node[i])
    return out

  def visit_dict(self, node):
    out = {}
    for key in node:
      out[self.visit(key)] = self.visit(node[key])
    return out

  def visit_tuple(self, node):
    lst = [None] * len(node)
    for i in xrange(len(node)):
      lst[i] = self.visit(node[i])
    return lst

def encode(obj):
  encoded = Jsonify("encode").visit(obj)
  return json.dumps(encoded)

def decode(string):
  obj = json.loads(string)
  return Jsonify("decode").visit(obj)
