#!/usr/bin/python
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


TestNode().start()
