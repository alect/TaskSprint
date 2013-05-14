import threading
import Queue
import time
import random
import types
import math
import sys

import pylab
import numpy

import multiplerootfinder_config

# Shared queue for passing a point between plotting and reading threads
pointQueue = Queue.Queue()

def plotter():
    params = multiplerootfinder_config.params

    # Look up the function lambda
    func = params['f_params']['function']
    if  isinstance(func, types.StringType) or \
        isinstance(func, types.UnicodeType):
        func = eval(func)

    # Assemble f(x) versus x points from argmin to argmax
    xl = numpy.arange(params['f_params']['argmin'], params['f_params']['argmax'], 0.01)
    yl = [func([x]) for x in xl]

    # Plot the function in black
    pylab.plot(xl, yl, color='black')
    pylab.xlabel('x')
    pylab.ylabel('y')
    pylab.title('Multiple Root Finder')
    pylab.grid(True)
    # Adjust the y limits to be a little beyond min and max of function
    d = max(yl) - min(yl)
    pylab.ylim(min(yl) - 0.10*d, max(yl) + 0.10*d)
    # Draw graph and toolbar
    pylab.draw()
    pylab.draw()

    while True:
        # Grab a new point from the queue
        p = pointQueue.get()
        # Exit cleanly if user hit Ctrl-C
        if p == "q": break
        # Plot p, f(p) as a circle in red
        pylab.plot([p], [func(p)], marker='o', color='red')
        # Draw graph and toolbar
        pylab.draw()
        pylab.draw()

# Pylab interactive mode on
pylab.ion()
# Start plotter thread
threading.Thread(target=plotter).start()

try:
    while True:
        # Read line and echo it back
        line = sys.stdin.readline().strip()
        print line

        # Parse found roots into points and pass to shared point queue
        if "Found Root!" in line:
            linesplit = line.split("\t")
            p = eval(linesplit[2])
            pointQueue.put(p)

except KeyboardInterrupt:
    # Tell other thread to exit cleanly in case of Ctrl-C here
    pointQueue.put("q")

