import sys, os, socket, inspect

sys.path.insert(1, os.path.join(sys.path[0], '..'))
import jsonify

class TaskSprintCoordinator:
  def __init__(self):
    if len(sys.argv) != 3:
      raise LookupError("TaskSprintNode: Wrong paramters.")
    self.socket_name = sys.argv[1]
    self.coord_socket_name = sys.argv[2]
    self.socket = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

  def start(self):
    self.clear_socket()
    self.socket.bind(self.socket_name)
    self.socket.listen(1)
    while True:
      (clientsocket, address) = self.socket.accept()
      self.recv(clientsocket, clientsocket.recv(4096))
      clientsocket.close()

  def start_task(self, **params):
    self.connect_to_coord()
    self.coord_socket.send("start_task:" + jsonify.encode(params))
    result = self.coord_socket.recv(256)
    self.coord_socket.close()
    return jsonify.decode(result)["tid"]

  def finish(self, **params):
    self.connect_to_coord()
    self.coord_socket.send("finish:" + jsonify.encode(params))
    self.coord_socket.close()

  def connect_to_coord(self):
    self.coord_socket = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    self.coord_socket.connect(self.coord_socket_name)

  def kill(self):
    self.clear_socket()
    sys.exit(0)

  def clear_socket(self):
    try:
      os.remove(self.socket_name)
    except:
      pass

  def recv(self, client, data):
    error, task, args = self.verify_data(data)
    if error != None:
      print "ERROR:", error
    getattr(self, task)(*args)

  def verify_data(self, data):
    error, taskfunc = None, None
    try:
      task, args = jsonify.decode(data)
      if type(args) != list:
        raise TypeError("Wrong argument type.")
      taskfunc = getattr(self, task)
      fargs, va, kw, defaults = inspect.getargspec(taskfunc)
      if len(fargs) > len(args) + 1:
        error = {"error" : "invalid_param_count"}
      elif len(fargs) != len(args) + 1 and not va:
        error = {"error" : "invalid_param_count"}
    except AttributeError:
      error = {"error" : "no_such_task"}
    except:
      error = {"error" : "invalid_data_format"}

    return error, task, args
