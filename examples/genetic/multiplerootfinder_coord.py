#!/usr/bin/python -u

import json

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/coordinator")
sys.path.append("examples/genetic")

from TaskSprintCoordinator import *
from ga_multiplerootfinder_lib import *

class MultipleRootFinderCoordinator(TaskSprintCoordinator):
    # Function, Genetic Algorithm, and Runtime Parameters
    params = {
        'f_params': {
            # Weird function with 11 roots from x = -6.0 to 6.0
            'function':     "lambda x: math.sin(3*x[0])*math.cos(0.5*x[0])",
            # Input dimensions of function (int)
            'dimensions':   1,
            # Search solution space (float)
            'argmin':       -12.0,
            'argmax':       +12.0,
            # Resolution in bits (int)
            'resolution':   32,

            # Solution epsilon (float)
            'solution_epsilon':   0.0001,
            # Cluster epsilon (float)
            'cluster_epsilon':     0.20,
            # Crossover epsilon (float)
            'crossover_epsilon':   7.00,
        },

        'ga_params': {
            # Total number of evolutions (int)
            'evolutions':   100,
            # New population size (int)
            'new_popsize':      200,
            # Minimum population size (int)
            'min_popsize':      100,
            # Keep percent population size (float)
            'keep_percent_popsize':     0.25,
            # Probability crossover (float)
            'prob_crossover':   0.30,
            # Probability mutation (float)
            'prob_mutation':    0.001,
        },

        'run_params': {
            # Max number of GAs to execute total (int)
            'runtime':      50,
            # Max number of GAs at any given time (int)
            'concurrent':   10,
        },
    }

    def init(self, seed):
        # Initialize known roots of the function to empty
        self.params['f_params']['solutions'] = []

        # Start concurrent GA tasks
        self.ga_tasks = []
        for i in range(self.params['run_params']['concurrent']):
            self.ga_tasks.append(self.start_task(name = "rootEvolve", base = [json.dumps(self.params)], keys = ["best", "generations"]))

        # Initialize our GA finished counter
        self.num_finished = 0

    def client_joined(self, cid, num_nodes):
        print "Client joined %d" % cid

    def client_dead(self, cid):
        print "Client dead %d" % cid

    def task_done(self, tid, values):
        # Catch confused / crashed clients here
        if 'best' not in values: return
        if tid not in self.ga_tasks: return

        # Remove the finished task from our active task list
        self.ga_tasks.remove(tid)

        # Count the finished task
        self.num_finished += 1
        print "GA %d / %d finished! Generations: %d" % (self.num_finished, self.params['run_params']['runtime'], values['generations'])

        (point, fitness) = values['best']

        # If this solution meets our epsilon
        if abs(fitness) < self.params['f_params']['solution_epsilon']:
            # Check that this solution is not close to past solutions
            solution_found = True
            for s in self.params['f_params']['solutions']:
                if point_distance(s, point, self.params['f_params']) < self.params['f_params']['cluster_epsilon']:
                    solution_found = False

            if solution_found:
                # Amend our function parameters with the solution
                self.params['f_params']['solutions'].append(point)
                print "Found root #%d: x = %s / f(x) = %f" % (len(self.params['f_params']['solutions']), str(point_fixed2float(point, self.params['f_params'])), fitness)

        # Start another GA task if we still have runtime
        if self.num_finished < self.params['run_params']['runtime']:
            self.ga_tasks.append(self.start_task(name = "rootEvolve", base = [json.dumps(self.params)], keys = ["best", "generations"]))

        # Otherwise, stop all GA tasks
        else:
            self.finish(taskids = self.ga_tasks)

MultipleRootFinderCoordinator().start()

