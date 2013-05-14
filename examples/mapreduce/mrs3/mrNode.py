#!/usr/bin/python
from TaskSprintNode import *

import pickle
import sys
import private

from boto.s3.connection import S3Connection
from boto.s3.key import Key


class MRNode(TaskSprintNode):

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

    def __init__(self):
        TaskSprintNode.__init__(self)
        self.s3_conn = S3Connection(private.aws_access_key_id, private.aws_secret_access_key)
        self.s3_bucket = self.s3_conn.get_bucket(private.bucket_name)
    
    def Partition(self, item, R):
        return hash(item[0]) % R
    
    def doMap(self, fname, keyvalue, mIndex, R):
        value = self.s3_get_string(fname)
        keyvaluelist = self.Map(keyvalue, value)
        results = {}
        for r in xrange(R):
            results['reduce-%s' % r] = [item for item in keyvaluelist if self.Partition(item, R) == r]
        fileResults = {}
        for result in results:
            resultname = 'map-%s-%s' % (mIndex, result)
            self.s3_set_obj(resultname, results[result])
            fileResults[result] = resultname
        return fileResults

    def doReduce(self, *mapResults):
        keys = {}
        out = []
        for mapResultName in mapResults:
            mapResult = self.s3_get_obj(mapResultName)
            for item in mapResult:
                if keys.has_key(item[0]):
                    keys[item[0]].append(item)
                else:
                    keys[item[0]] = [item]
        for k in sorted(keys.keys()):
            out.append(self.Reduce(k, keys[k]))
        outname = 'reduce-%s' % mapResults[0]
        self.s3_set_obj(outname, out)
        return { 'output' : outname }

    # Override this function for different mapreduce programs
    def doMerge(self, *reduceResults):
        results = []
        for resultName in reduceResults:
            result = self.s3_get_obj(resultName)
            results.append(result)
        finalResults = self.Merge(results)
        outputname = 'final-output'
        self.s3_set_obj(outputname, finalResults)
        return { 'output' : outputname }
        
    # Override this function for different mapreduce programs
    def Map(self, keyvalue, value):
        pass

    # Override this function for different mapreduce programs
    def Reduce(self, key, keyvalues):
        pass

    # Override this function for different mapreduce programs
    def Merge(self, results):
        pass
