#!/usr/bin/python
from TaskSprintCoordinator import *

import os
import sys
import string
import pickle

# M and R values
M = 8
R = 4

inputfile = 'kjv12.txt'

# Coordinator for the MapReduce example that relies on shared storage
class MRCoordinator(TaskSprintCoordinator):
    def init(self, seed):
        self.M = M
        self.R = R
        # To keep track of the files we'll later need to remove
        self.splitNames = []
        self.mapFileNames = []
        self.reduceFileNames = []
        self.mapTasks = []
        self.reduceTasks = []
        # First, split our input file
        size = os.stat(inputfile).st_size
        chunk = size / M
        chunk += 1
        f = open(inputfile, 'r')
        buffer = f.read()
        f.close()
        fname = '#split-%s-%s' % (inputfile, 0)
        self.splitNames.append(fname)
        offset = 0
        f = open(fname, 'w+')
        i = 0
        m = 1
        for c in buffer:
            f.write(c)
            i += 1
            if (c in string.whitespace) and (i > chunk * m):
                f.close()
                self.mapTasks.append(self.start_task(
                    name='doMap',
                    base=[fname, offset, m, R],
                    keys=['reduce-%s' % r for r in xrange(R)]
                ))
                m+=1
                fname = '#split-%s-%s' % (inputfile, m-1)
                self.splitNames.append(fname)
                f = open(fname, 'w+')
                offset = i
        f.close()

        # Now, schedule all of the reduce tasks
        for i in xrange(R):
            self.reduceTasks.append(self.start_task(
                name='doReduce',
                prekeys=['reduce-%s' % i for m in range(len(self.mapTasks))],
                base=[i],
                pretasks=self.mapTasks[:],
                keys=['output']
            ))
        self.final = self.start_task(
            name='doMerge',
            prekeys=['output' for i in xrange(len(self.reduceTasks))],
            pretasks=self.reduceTasks[:],
            keys=['output']
        )

    def client_joined(self, cid):
        pass

    def client_dead(self, cid):
        pass

    def task_done(self, tid, values):
        if tid in self.mapTasks:
            for key in values:
                self.mapFileNames.append(values[key])
        elif tid in self.reduceTasks:
            self.reduceFileNames.append(values['output'])
        elif tid == self.final:
            fake_vals = { 'result' : 22222 }
            self.finish(taskids = [tid], values = fake_vals)
            fname = values['output']
            try:
                f = open(fname, 'r')
                final_val = pickle.load(f)
                f.close()
                sys.stderr.write('Finished MapReduce. Result: %s\n' % final_val)
                os.unlink(fname)
            except:
                pass
            # Clean up
            for name in self.splitNames:
                try:
                    os.unlink(name)
                except:
                    pass
            for name in self.mapFileNames:
                try:
                    os.unlink(name)
                except:
                    pass
            for name in self.reduceFileNames:
                try:
                    os.unlink(name)
                except:
                    pass
            

MRCoordinator().start()
