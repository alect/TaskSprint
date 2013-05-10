import random
import math
import numpy
import pylab
import bisect
import sys



################################################################################
# Point Utility Functions
################################################################################

# Convert a fixed width representation of a point to a float representation of
# a point
def point_fixed2float(point, f_params):
    fl_point = []
    for p in point:
        fl_point.append((f_params['argmax']-f_params['argmin'])*(float(p) / (2**f_params['resolution'])) + f_params['argmin'])
    return fl_point

# Evaluate the Euclidean distance between two points
def point_distance(p1, p2, f_params):
    distance = 0
    for i in range(f_params['dimensions']):
        distance += (p1[i] - p2[i])**2
    distance = int(math.sqrt(distance))
    return distance

# Evaluate the fitness of a point on f_params function
def point_fitness(point, f_params):
    point = point_fixed2float(point, f_params)
    return -1.0*abs(f_params['function'](point))

# Generate a new point
def point_generate(f_params):
    point = []
    for j in range(f_params['dimensions']):
        point.append( random.randint(0, 2**f_params['resolution']) )

    return point

# Crossover two points
def point_crossover(p1, p2, f_params, prob_crossover):
    # If crossover occurs and the points are within crossover distance
    if random.random() < prob_crossover and \
        point_distance(p1, p2, f_params) < f_params['crossover_distance']:
        # Average both points into one
        p3 = []
        for j in range(len(p1)):
            p3.append( (p1[j] + p2[j]) / 2 )
        return [p3]

    # Return original parents
    return [p1, p2]

# Mutate a point
def point_mutate(p, f_params, prob_mutation):
    # For each dimension of the point
    for i in range(f_params['dimensions']):
        # For each of the bits
        for j in range(f_params['resolution']):
            # If mutation occurs, flip a bit
            if (random.random() < prob_mutation):
                p[i] = p[i] ^ 2**j

    return p

################################################################################

#
# Evolve a population to a new generation,
# including initialization, selection, crossover, mutation.
#
# args  population      (list) (population, fitness) tuples
#       params          (dict) function and ga parameters
#
# ret   (list) (population, fitness) tuples
#
def evolve(population, params):
    f_params = params['f_params']
    ga_params = params['ga_params']

    # Initialize population with random members
    if len(population) == 0:
        for i in range(ga_params['new_popsize']):
            p = point_generate(f_params)
            f = point_fitness(p, f_params)
            population.append( (p, f) )

    # Fill population to minimum size with random members
    for i in range(ga_params['min_popsize'] - len(population)):
        p = point_generate(f_params)
        f = point_fitness(p, f_params)
        population.append( (p, f) )

    # Sort population by fitness
    population = sorted(population, key=lambda x: x[1])

    target_popsize = ga_params['keep_percent_popsize']*len(population)
    population_evolved = []

    while len(population_evolved) < target_popsize and len(population) > 2:
        # Choose fit member p1 by elitist selection
        p1 = population.pop()[0]
        # Choose fit member p2 by elitist selection
        p2 = population.pop()[0]

        # Crossover p1 and p2
        p_new = point_crossover(p1, p2, f_params, ga_params['prob_crossover'])
        # Mutate children
        p_new = [point_mutate(p, f_params, ga_params['prob_mutation']) for p in p_new]

        # Calculate their fitness and add them to our evolved population
        for p in p_new:
            population_evolved.append( (p, point_fitness(p, f_params)) )

    # Sort evolved population by fitness
    population_evolved = sorted(population_evolved, key=lambda x: x[1])

    return population_evolved

################################################################################

params = {
    'f_params': {
        # Polynomial with roots at -2, 2.5, 3, 3.3, 10
        'function':     lambda x: x[0]**5 - 16.8*x[0]**4 + 76.05*x[0]**3 - 53.95*x[0]**2 - 315.0*x[0] + 495.0,
        'dimensions':   1,
        # Search solution space x = [-25.0, 25.0]
        'argmin':       -25.0,
        'argmax':       25.0,
        # Resolution in bits
        'resolution':   20,

        # Solution epsilon (float)
        'epsilon':      0.05,

        # Crossover distance
        'crossover_distance':     2500,
        # Cluster epsilon (integer)
        'cluster_epsilon':        500,
    },

    'ga_params': {
        # New population size
        'new_popsize':      300,
        # Minimum population size
        'min_popsize':      100,
        # Keep percent population size
        'keep_percent_popsize':     0.70,
        # Probability crossover
        'prob_crossover':   0.30,
        # Probability mutation
        'prob_mutation':    0.001,
    }
}

solutions = []

# Attempt 10 different GAs
for attempt in range(50):
    print "Root-find attempt %d..." % attempt

    population = []

    # Enter GA evolution loop
    for generation in range(250):
        population = evolve(population, params)

        #print generation, [x[1] for x in population[-5:]]
        #sys.stdout.flush()

        # If the best fitness makes the solution epsilon
        if abs(population[-1][1]) < params['f_params']['epsilon']:
            # If our solution set is empty, add it
            if len(solutions) == 0:
                solutions.append( (population[-1][0], population[-1][1], generation) )
                print "\tFound a solution!"
                break

            # Check that this solution isn't close to other found solutions
            else:
                solution_found = True
                for s in solutions:
                    if point_distance(s[0], population[-1][0], params['f_params']) < params['f_params']['cluster_epsilon']:
                        solution_found = False
                if solution_found:
                    solutions.append( (population[-1][0], population[-1][1], generation) )
                    print "\tFound a solution!"
                    break

print "\nFound %d solutions\n" % len(solutions)

for s in solutions:
    print "Point:\t\t%s\nFitness:\t%f\nGeneration:\t%d\n" % (str(point_fixed2float(s[0], params['f_params'])), s[1], s[2])

