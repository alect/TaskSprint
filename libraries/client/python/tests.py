#!/usr/bin/python
import unittest, threading, json, subprocess, socket, time

def start_sub_proc():
  subprocess.call(["./testNode.py", "socket"]) 

def runTask(name, data = []):
  s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
  s.connect("socket")
  s.send(json.dumps([name, data]))
  data = json.dumps("done")
  if name != "kill":
    data = s.recv(4096)
  s.close()
  return json.loads(data)

class TestPythonLibrary(unittest.TestCase):
  @classmethod
  def setUpClass(self):
    thread = threading.Thread(target = start_sub_proc)
    thread.start()
    time.sleep(0.5) # wait for thrad to start

  @classmethod
  def tearDownClass(self):
    runTask("kill")
    time.sleep(0.5)

  def test_task1(self):
    result = runTask("task1")
    self.assertTrue("result" in result)
    self.assertDictEqual(result, {"result" : True})

  def test_task2(self):
    result = runTask("task2", ["sergio"])
    self.assertDictEqual(result, {"name" : "sergio"})
    result = runTask("task2", ["mike"])
    self.assertDictEqual(result, {"name" : "mike"})
    result = runTask("task2", [""])
    self.assertDictEqual(result, {"name" : ""})

  def test_task3(self):
    result = runTask("task3", ["sergio", "benitez"])
    self.assertDictEqual(result, {
      "first" : "sergio",
      "last" : "benitez",
      "name" : "sergio benitez"
      })
    result = runTask("task3", ["mike", "mican"])
    self.assertDictEqual(result, {
      "first" : "mike",
      "last" : "mican",
      "name" : "mike mican"
      })

  def test_multiply(self):
    result = runTask("multiply", [1, 10])
    self.assertDictEqual(result, {"result" : 10})
    result = runTask("multiply", [20, 10])
    self.assertDictEqual(result, {"result" : 200})
    result = runTask("multiply", [0, 0])
    self.assertDictEqual(result, {"result" : 0})

  def test_invalid_task(self):
    result = runTask("divide", [10, 2])
    self.assertTrue("error" in result)
    self.assertTrue("no" in result["error"])
    self.assertTrue("task" in result["error"])

  def test_invalid_params(self):
    result = runTask("multiply", [10])
    self.assertTrue("error" in result)
    self.assertTrue("invalid" in result["error"])
    self.assertTrue("param" in result["error"])
    self.assertTrue("count" in result["error"])
    result = runTask("multiply", [10, 100, 120])
    self.assertTrue("error" in result)
    self.assertTrue("invalid" in result["error"])
    self.assertTrue("param" in result["error"])
    self.assertTrue("count" in result["error"])
    result = runTask("multiply", [])
    self.assertTrue("error" in result)
    self.assertTrue("invalid" in result["error"])
    self.assertTrue("param" in result["error"])
    self.assertTrue("count" in result["error"])

  def test_invalid_data(self):
    result = runTask("multiply", "nothing")
    self.assertTrue("error" in result)
    self.assertTrue("data" in result["error"])
    self.assertTrue("format" in result["error"])
    result = runTask("multiply", 1)
    self.assertTrue("error" in result)
    self.assertTrue("data" in result["error"])
    self.assertTrue("format" in result["error"])

if __name__ == '__main__':
  unittest.main()
