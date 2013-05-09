#!/usr/bin/python
import sys, os, unittest, threading, subprocess, socket, time, datetime

sys.path.insert(1, os.path.join(sys.path[0], '..'))
import jsonify

class Person(object):
  def __init__(self, name, age):
    self.name = name
    self.age = age

  def year_born(self):
    return datetime.date.today().year - self.age - 1

  def first_name(self):
    return self.name.split(" ")[0]

  def last_name(self):
    return self.name.split(" ")[1]

def start_sub_proc():
  subprocess.call(["./testNode.py", "socket"]) 

def runTask(name, data = []):
  s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
  s.connect("socket")
  s.send(jsonify.encode([name, data]))
  data = jsonify.encode("done")
  if name != "kill":
    data = s.recv(4096)
  s.close()
  return jsonify.decode(data)

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

  def test_returned_object(self):
    result = runTask("person", ["Sergio Benitez", 21])
    self.assertTrue("person" in result)
    person = result["person"]
    self.assertEqual(person.name, "Sergio Benitez")
    self.assertEqual(person.age, 21)
    self.assertEqual(person.first_name(), "Sergio")
    self.assertEqual(person.last_name(), "Benitez")
    self.assertEqual(person.year_born(), 1991)

  def test_new_object(self):
    result = runTask("detail_person", [Person("Sergio Benitez", 21)])
    self.assertDictEqual(result, {
      "name" : "Sergio Benitez",
      "age" : 21,
      "first_name" : "Sergio",
      "last_name" : "Benitez",
      "year_born" : 1991,
    })

  def test_variable_nums(self):
    result = runTask("sum", [1, 2, 3, 4, 5])
    self.assertDictEqual(result, {"result" : 15})
    result = runTask("sum", [3, 4, 5])
    self.assertDictEqual(result, {"result" : 12})

if __name__ == '__main__':
  unittest.main()
