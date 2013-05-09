#!/usr/bin/python
from test_library import Person # for testing only
from TaskSprintNode import *

class TestNode(TaskSprintNode):
  def task1(self):
    return {"result" : True}

  def task2(self, name):
    return {"name" : name}

  def task3(self, first, last):
    return {
      "first" : first, 
      "last" : last,
      "name" : first + " " + last
    }
  
  def multiply(self, x, y):
    return {"result": x * y}

  def person(self, name, age):
    return {"person": Person(name, age)}

  def sum(self, *nums):
    return {"result": sum(nums)}

  def detail_person(self, person):
    result = {}
    result["name"] = person.name
    result["age"] = person.age
    result["first_name"] = person.first_name()
    result["last_name"] = person.last_name()
    result["year_born"] = person.year_born()
    return result

TestNode().start()
