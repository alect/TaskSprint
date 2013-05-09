#!/usr/bin/python
from TaskSprintCoordinator import *

class TestCoordinator(TaskSprintCoordinator):
  def init(self, seed):
    print "Should initialize with a seed: %d" %seed
    tid = self.start_task(
        name = "sum",
        base = [1, 2, 3, 4],
        keys = ["result"]
    )
    print "Started task %d" %tid

  def client_joined(self, cid):
    print "Somebody joined! %d" %cid

  def client_dead(self, cid):
    print "Somebody died :( %d" %cid

  def task_done(self, tid, values):
    self.finish(tid = tid, values = values)
    print "Task %d is done. Result: %s" %(tid, values)

TestCoordinator().start()
