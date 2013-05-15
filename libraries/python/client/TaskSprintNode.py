from functools import wraps
import sys, os, socket, inspect, errno, signal

sys.path.insert(1, os.path.join(sys.path[0], '..'))
import jsonify

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
      (clientsocket, address), data, max_buf = self.socket.accept(), "", 1024
      while True:
        buf = clientsocket.recv(max_buf)
        data += buf
        if len(buf) < max_buf: break
      self.recv(clientsocket, data)
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
      return client.sendall(jsonify.encode(error))
    client.sendall(jsonify.encode(getattr(self, task)(*args)))

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

class TimeoutError(Exception):
    pass

def timeout(seconds=10, error_message=os.strerror(errno.ETIME)):
    def decorator(func):
        def _handle_timeout(signum, frame):
            raise TimeoutError(error_message)

        def wrapper(*args, **kwargs):
            signal.signal(signal.SIGALRM, _handle_timeout)
            signal.alarm(seconds)
            try:
                result = func(*args, **kwargs)
            finally:
                signal.alarm(0)
            return result

        return wraps(func)(wrapper)

    return decorator
