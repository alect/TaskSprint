#!/usr/bin/python -u

import json

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/client")
sys.path.append("examples/genetic")

from TaskSprintNode import *
from ga_multiplerootfinder_lib import *

class MultipleRootFinderNode(TaskSprintNode):
    def rootEvolve(self, *params):
        params = json.loads(params[0])
        print "This client is evolving!"

        # Initialize population to empty
        population = []

        # Evolve for ga_params.evolutions times
        for i in range(params['ga_params']['evolutions']):
            population = evolve(population, params)

            # If we've met the solution epsilon, stop evolution early
            if abs(population[-1][1]) < params['f_params']['solution_epsilon']:
                break

        ## Return most fit member of our population
        return {'best' : population[-1], 'generations': i+1}

MultipleRootFinderNode().start()

