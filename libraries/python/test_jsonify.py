#!/usr/bin/python
import json, jsonpickle, jsonify, unittest
import datetime

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

  def __eq__(self, other):
    return (isinstance(other, self.__class__)
        and self.__dict__ == other.__dict__)

  def __ne__(self, other):
    return not self.__eq__(other)

class TestJsonify(unittest.TestCase):
  def encode_test(self, obj):
    string = jsonify.encode(obj)
    self.assertEqual(string, json.dumps(obj))

  def test_simple_list_int_encode(self):
    self.encode_test([1,2,3])
    self.encode_test([1,2,3,5,-12])
    self.encode_test([])
    self.encode_test((1,2,3))
    self.encode_test((2,3))
    self.encode_test((2,3,-10,12,0))

  def test_simple_list_mixed_encode(self):
    self.encode_test([1,"hello",3])
    self.encode_test([1,"hello",3,5,"hi"])
    self.encode_test(["hello", "hi", "bye!"])
    self.encode_test([True, False])
    self.encode_test([True, 1, "hi"])
    self.encode_test((True, False))
    self.encode_test((True, 1, "hi"))
    self.encode_test((1,"hello",3))
    self.encode_test((2,"hi", "hi", "hi"))
    self.encode_test(("hi", "bye"))

  def test_nested_list_encode(self):
    self.encode_test([[1,2,3], [1], []])
    self.encode_test([[1,2,3], (1), ()])
    self.encode_test(((1,2,3), (1), ()))
    self.encode_test(((1,2,3), [1,2], ()))
    self.encode_test([[True,"hi",3], ["hello"], [[1,"false",3]]])
    self.encode_test([[False,2,"hi"], (True), ()])
    self.encode_test((('Hello',2,False), (1), ()))
    self.encode_test(((1,True,3), ["yo",2], (())))

  def test_simple_dict_encode(self):
    d = {"hello" : 1, "test" : 2}
    self.encode_test(d)
    d = {2 : True, "test" : False, 1 : 1, True: 1, False : True}
    self.encode_test(d)

  def test_nested_dict_encode(self):
    d = {2 : (1, [1,2], True), "test" : {1 : "one", 2 : [1,2]}, 2 : [1,2]}
    self.encode_test(d)

  def test_simple_object_encode(self):
    obj = Person("Sergio", 21)
    string = json.dumps(jsonpickle.encode(obj))
    self.assertEqual(string, jsonify.encode(obj))

  def test_simple_nested_object_encode(self):
    obj = Person("Sergio", 21)
    fullobj = [1, 2, obj]
    picklestring = jsonpickle.encode(obj)
    jsonobj = [1, 2, picklestring]
    string = json.dumps(jsonobj)
    self.assertEqual(string, jsonify.encode(fullobj))

    obj = Person("Sergio", 21)
    fullobj = [obj, [1,2,obj], obj, (obj, (obj))]
    os = jsonpickle.encode(obj)
    jsonos = [os, [1,2,os], os, (os, (os))]
    string = json.dumps(jsonos)
    self.assertEqual(string, jsonify.encode(fullobj))

  def test_simple_decode(self):
    obj = [1, 2, 3]
    string = json.dumps(obj)
    self.assertEqual(obj, jsonify.decode(string))

    obj = [True, 2, 3]
    string = json.dumps(obj)
    self.assertEqual(obj, jsonify.decode(string))

  def ed_test(self, obj):
    string = jsonify.encode(obj)
    decodeobj = jsonify.decode(string)
    self.assertEqual(obj, decodeobj)

  def test_encode_then_decode(self):
    self.ed_test([1,2,3])
    self.ed_test([[1,2,3],["hi", 2],3])
    self.ed_test([[1,2,3], [[Person("Sergio", 21)], 2],3])
    self.ed_test([{"1":2, "me":Person("me", 12)}, 
        ([(Person("Sergio", 21))]), [2,False]])

if __name__ == '__main__':
  unittest.main()
