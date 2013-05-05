import sys, os, socket, inspect, jsonify

class TaskSprintNode:
  def __init__(self):
    if len(sys.argv) != 2:
      raise LookupError("TaskSprintNode: No socket name found.")
    self.socket_name = sys.argv[1]
    self.socket = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

  def start(self):
    self.clear_socket()
    self.socket.bind(self.socket_name)
    self.socket.listen(1)
    while True:
      (clientsocket, address) = self.socket.accept()
      self.recv(clientsocket, clientsocket.recv(4096))
      clientsocket.close()

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
      return client.send(jsonify.encode(error))
    client.send(jsonify.encode(getattr(self, task)(*args)))

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
