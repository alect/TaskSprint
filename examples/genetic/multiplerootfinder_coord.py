#!/usr/bin/python -u

import json

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/coordinator")
sys.path.append("examples/genetic")

from TaskSprintCoordinator import *
from ga_multiplerootfinder_lib import *

import multiplerootfinder_config

class MultipleRootFinderCoordinator(TaskSprintCoordinator):
    def init(self, seed):
        # Initialize known roots of the function to empty
        self.params = multiplerootfinder_config.params
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
                print "Found Root!\t#%d\t%s\t%f" % (len(self.params['f_params']['solutions']), str(point_fixed2float(point, self.params['f_params'])), fitness)

        # Start another GA task if we still have runtime
        if self.num_finished < self.params['run_params']['runtime']:
            self.ga_tasks.append(self.start_task(name = "rootEvolve", base = [json.dumps(self.params)], keys = ["best", "generations"]))

        # Otherwise, stop all GA tasks
        else:
            for t in self.ga_tasks:
                self.kill_task(tid = t)
            self.ga_tasks = []

            print "\nFound %d roots" % len(self.params['f_params']['solutions'])

MultipleRootFinderCoordinator().start()

