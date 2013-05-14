#!/usr/bin/python
from TaskSprintCoordinator import *

import os
import sys
import string
import pickle
import private

from boto.s3.connection import S3Connection
from boto.s3.key import Key


# M and R values
M = 8
R = 4

inputfile = 'kjv12.txt'


# Coordinator for the MapReduce example that relies on shared storage
class MRCoordinator(TaskSprintCoordinator):
    def s3_set_obj(self, key, value):
        k = Key(self.s3_bucket)
        k.key = key
        k.set_contents_from_string(pickle.dumps(value))

    def s3_set_file(self, key, filename):
        k = Key(self.s3_bucket)
        k.key = key
        k.set_contents_from_filename(filename)

    def s3_get_string(self, key):
        k = Key(self.s3_bucket)
        k.key = key
        return k.get_contents_as_string()
    
    def s3_get_obj(self, key):
        k = Key(self.s3_bucket)
        k.key = key
        return pickle.loads(k.get_contents_as_string())

    def s3_get_file(self, key, filename):
        k = Key(self.s3_bucket)
        k.key = key
        k.get_contents_to_filename(filename)
    
    def init(self, seed):
        self.s3_conn = S3Connection(private.aws_access_key_id, private.aws_secret_access_key)
        self.s3_bucket = self.s3_conn.get_bucket(private.bucket_name)
        
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
                # Send this file to s3
                self.s3_set_file(fname, fname)
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
                prekeys=['reduce-%s' % i for m in xrange(len(self.mapTasks))],
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

    def client_joined(self, cid, num_nodes):
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
                final_val = self.s3_get_obj(fname)
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
            
os.chdir(os.path.dirname(os.path.abspath(__file__)))
MRCoordinator().start()
