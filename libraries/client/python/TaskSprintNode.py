import sys, os, socket, json, inspect, jsonpickle

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
    print "Listening for tasks..."
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

  def unpickle(self, arg):
    if type(arg) == unicode and "py/object" in arg:
      return jsonpickle.decode(arg)
    return arg

  def pickle(self, data_dict):
    for key in data_dict:
      d = data_dict[key]
      if hasattr(d, '__module__') and d.__module__ != "__builtin__":
        data_dict[key] = jsonpickle.encode(d)
    return data_dict

  def recv(self, client, data):
    error, task, args = self.verify_data(data)
    if error != None:
      return client.send(json.dumps(error))
    args = map(self.unpickle, args)
    client.send(json.dumps(self.pickle(getattr(self, task)(*args))))

  def verify_data(self, data):
    error, taskfunc = None, None
    try:
      task, args = json.loads(data)
      if type(args) != list:
        raise TypeError("Wrong argument type.")
      taskfunc = getattr(self, task)
      if len(inspect.getargspec(taskfunc)[0]) != len(args) + 1:
        error = {"error" : "invalid_param_count"}
    except AttributeError:
      error = {"error" : "no_such_task"}
    except:
      error = {"error" : "invalid_data_format"}

    return error, task, args
