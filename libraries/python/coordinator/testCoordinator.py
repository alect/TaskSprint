#!/usr/bin/python
from TaskSprintCoordinator import *

class TestCoordinator(TaskSprintCoordinator):
  def init(self, seed):
    alltasks = []
    num = 200
    for i in xrange(num):
      tid = self.start_task(
        name = "sum",
        base = [i * (num / 10) + k for k in range(num / 10)],
      )
      alltasks.append(tid)

    finaltasks = []
    toassign, size = num, num / 10
    for i in xrange(toassign / size):
      pretasks = alltasks[i * size : (i + 1) * size]
      print pretasks
      tid = self.start_task(
        name = "sum",
        prekeys = ["result" for i in range(size)],
        pretasks = pretasks
      )
      finaltasks.append(tid)

    self.final = self.start_task(
      name = "sum",
      prekeys = ["result" for i in range(len(finaltasks))],
      pretasks = finaltasks,
      keys = ["result"]
    )

  def client_joined(self, cid, num_nodes):
    print "Somebody joined! %d" %cid

  def client_dead(self, cid):
    print "Somebody died :( %d" %cid

  def task_done(self, tid, values):
    if tid == self.final:
      self.finish(taskids = [tid], values = values)
    print "Task %d is done. Result: %s" %(tid, values)

TestCoordinator().start()
