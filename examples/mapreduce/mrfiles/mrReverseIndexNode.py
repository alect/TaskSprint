#!/usr/bin/python
from mrNode import *

import string
import sys

class ReverseIndex(MRNode):

    def Map(self, keyvalue, value):
        results = []
        i = 0
        n = len(value)
        while i < n:
            while i < n and value[i] not in string.ascii_letters:
                i += 1
            start = i
            while i < n and value[i] in string.ascii_letters:
                i += 1
            w = value[start:i]
            if start < i:
                results.append([w.lower(), i+int(keyvalue)])
        return results

    def Reduce(self, key, keyvalues):
        return (key, sorted([value for [key, value] in keyvalues]))


    def Merge(self, results):
        out = []
        for result in results:
            out.extend(result)
        out = sorted(out, key=lambda pair: pair[0])
        return out[10:22]
        
os.chdir(os.path.dirname(os.path.abspath(__file__)))
ReverseIndex().start()
