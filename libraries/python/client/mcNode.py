#!/usr/bin/python
from TaskSprintNode import *
import random

class MonteCarloNode(TaskSprintNode):
  @timeout(5)
  def process(self):
    while True:
      x, y = random.random(), random.random()
      if x**2 + y**2 < 1: self.inside += 1
      self.total += 1

  def montecarlo(self):
    self.inside, self.total = 0, 0
    try:
      self.process()
    except TimeoutError:
      pass
    return {"result" : {"inside" : self.inside, "total" : self.total}}

  def merge(self, *results):
    inside, total = 0, 0
    for result in results:
      inside += result["inside"]
      total += result["total"]
    return {"result" : (inside / float(total)) * 4}

MonteCarloNode().start()
