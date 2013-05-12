#!/usr/bin/python
from TaskSprintNode import *

import pickle
import sys

class MRNode(TaskSprintNode):
    def Partition(self, item, R):
        return hash(item[0]) % R
    
    def doMap(self, fname, keyvalue, mIndex, R):
        f = open(fname, 'r')
        value = f.read()
        f.close()
        keyvaluelist = self.Map(keyvalue, value)
        results = {}
        for r in xrange(R):
            results['reduce-%s' % r] = [item for item in keyvaluelist if self.Partition(item, R) == r]
        fileResults = {}
        for result in results:
            resultname = 'map-%s-%s' % (mIndex, result)
            f = open(resultname, 'w+')
            pickle.dump(results[result], f)
            f.close()
            fileResults[result] = resultname
        return fileResults

    def doReduce(self, *mapResults):
        keys = {}
        out = []
        for mapResultName in mapResults:
            f = open(mapResultName, 'r')
            mapResult = pickle.load(f)
            f.close()
            for item in mapResult:
                if keys.has_key(item[0]):
                    keys[item[0]].append(item)
                else:
                    keys[item[0]] = [item]
        for k in sorted(keys.keys()):
            out.append(self.Reduce(k, keys[k]))
        outname = 'reduce-%s' % mapResults[0]
        f = open(outname, 'w+')
        pickle.dump(out, f)
        f.close()
        return { 'output' : outname }

    # Override this function for different mapreduce programs
    def doMerge(self, *reduceResults):
        results = []
        for resultName in reduceResults:
            f = open(resultName, 'r')
            result = pickle.load(f)
            results.append(result)
        finalResults = self.Merge(results)
        outputname = 'final-output'
        f = open(outputname, 'w+')
        pickle.dump(finalResults, f)
        f.close()
        return { 'output' : outputname }
        
    # Override this function for different mapreduce programs
    def Map(self, keyvalue, value):
        pass

    # Override this function for different mapreduce programs
    def Reduce(self, key, keyvalues):
        pass

    # Override this funciton for different mapreduce programs
    def Merge(self, results):
        pass
