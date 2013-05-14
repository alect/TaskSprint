#!/usr/bin/python -u
from TaskSprintCoordinator import *

class MonteCarloCoordinator(TaskSprintCoordinator):
  def init(self, seed):
    alltasks = [self.start_task(name = "montecarlo") for i in xrange(11)]

    self.final = self.start_task(
      name = "merge",
      prekeys = ["result" for i in range(len(alltasks))],
      pretasks = alltasks,
      keys = ["result"]
    )

  def client_joined(self, cid, num_nodes):
    print "Somebody joined! %d" %cid

  def client_dead(self, cid):
    print "Somebody died :( %d" %cid

  def task_done(self, tid, values):
    if tid == self.final:
      self.finish(taskids = [tid], values = values)

    message = "Task %d is done. Result: %s" %(tid, values)
    with open("result", "a") as f:
      f.write(message)

    print message

MonteCarloCoordinator().start()
