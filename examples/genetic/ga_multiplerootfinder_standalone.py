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
            'epsilon':             0.0001,
            # Cluster epsilon (float)
            'cluster_epsilon':     0.20,
            # Crossover epsilon (float)
            'crossover_epsilon':   7.00,
        },

        'ga_params': {
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
            # Number of Fresh GAs to Evolve (int)
            'runtime':      20,
            # Evolutions per GA
            'evolutions':   100,
        },
    }

solutions = []

params['f_params']['roots_found'] = []

for attempt in range(params['run_params']['runtime']):
    print "Root-find attempt %d..." % attempt

    population = []

    # Enter GA evolution loop
    for generation in range(params['run_params']['evolutions']):
        population = evolve(population, params)

        # If the best fitness makes the solution epsilon
        if abs(population[-1][1]) < params['f_params']['epsilon']:
            # Check that this solution isn't close to other found solutions
            solution_found = True
            for s in solutions:
                if point_distance(s[0], population[-1][0], params['f_params']) < params['f_params']['cluster_epsilon']:
                    solution_found = False

            if solution_found:
                solutions.append( (population[-1][0], population[-1][1], generation) )
                params['f_params']['roots_found'].append(population[-1][0])
                print "\tFound a new solution!"
                break

print "\nFound %d solutions\n" % len(solutions)

for s in solutions:
    print "Point:\t\t%s\nFitness:\t%f\nGeneration:\t%d\n" % (str(point_fixed2float(s[0], params['f_params'])), s[1], s[2])

