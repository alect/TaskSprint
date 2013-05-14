import sys
sys.path.append("../")

from ga_multiplerootfinder_lib import *

################################################################################
# Standalone Multiple Root Finder, Single-threaded
################################################################################

# Function, Genetic Algorithm, and Runtime Parameters
params = {
        'f_params': {
            # Weird function with 11 roots from x = -6.0 to 6.0
            'function':     lambda x: math.sin(3*x[0])*math.cos(0.5*x[0]),
            # Input dimensions of function (int)
            'dimensions':   1,
            # Search solution space (float)
            'argmin':       -6.0,
            'argmax':       +6.0,
            # Resolution in bits (int)
            'resolution':   32,

            # Solution epsilon (float)
            'solution_epsilon':    0.0001,
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
            'runtime':      25,
        },
    }

solutions = []

# Initialize known roots of the function to empty
params['f_params']['solutions'] = []

for attempt in range(params['run_params']['runtime']):
    print "Root-find attempt %d..." % (attempt+1)

    population = []

    # Enter GA evolution loop
    for generation in range(params['ga_params']['evolutions']):
        population = evolve(population, params)

        # If the best fitness makes the solution epsilon
        if abs(population[-1][1]) < params['f_params']['solution_epsilon']:
            point, fitness = population[-1]

            # Check that this solution isn't close to other found solutions
            solution_found = True
            for s in solutions:
                if point_distance(s, point, params['f_params']) < params['f_params']['cluster_epsilon']:
                    solution_found = False

            if solution_found:
                params['f_params']['solutions'].append(point)
                print "\tFound root #%d on GA gen %d: x = %s / f(x) = %f" % (len(params['f_params']['solutions']), generation, str(point_fixed2float(point, params['f_params'])), fitness)
                break

