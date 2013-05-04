import random
import math
import numpy
import pylab
import bisect
import sys

# Convert a fixed width representation of a point to a float representation of
# a point
def fixed2float(point, fprops):
    fl_point = []
    for p in point:
        fl_point.append((fprops['argmax']-fprops['argmin'])*(float(p) / (2**fprops['resolution'])) + fprops['argmin'])
    return fl_point

# Evaluate the fitness of a point on function fprops['function']
def fitness(point, fprops):
    point = fixed2float(point, fprops)
    return -1.0*abs(fprops['function'](point))

################################################################################
################################################################################

#
# Randomly generate a new population of size n.
#
# args  n               (int) size of population
#       fprops          (list) function properties
#
# ret   (list) (population, fitness) tuples
def ga_generate(n, fprops):
    population = []

    for i in range(n):
        point = []
        for j in range(fprops['dimensions']):
            point.append( random.randint(0, 2**fprops['resolution']) )

        population.append( (point, fitness(point, fprops)) )

    return population

#
# Evolve a population to a new generation,
# including selection, crossover, mutation.
#
# args  population      (list) population, fitness tuples
#       n               (int) size of new population
#       pr_crossover    (float) crossover probabiltiy
#       pr_mutation     (float) mutation probability
#       fprops          (list) function properties
#
# ret   (list) (population, fitness) tuples
def ga_evolve(population, n, pr_crossover, pr_mutation, fprops):
    # Sort population by increasing fitness
    population = sorted(population, key=lambda x: x[1])

    def crossover(p1, p2):
        # If crossover occurs, average both points into one
        if (random.random() < pr_crossover):
            p3 = []
            for j in range(len(p1)):
                p3.append( (p1[j] + p2[j]) / 2 )
            return [p3]

        # Return original parents
        return [p1, p2]

    def mutate(p):
        for i in range(fprops['dimensions']):
            for j in range(fprops['resolution']):
                # If mutation occurs, flip a bit
                if (random.random() < pr_mutation):
                    p[i] = p[i] ^ 2**j
        return p

    population_evolved = []

    while len(population_evolved) < n or len(population) < 2:
        # Choose fit member p1 by elitist selection
        p1 = population.pop()[0]
        # Choose fit member p2 by elitist selection
        p2 = population.pop()[0]

        # Crossover p1 and p2
        p_new = crossover(p1, p2)
        # Mutate children
        p_new = [mutate(p) for p in p_new]

        # Add them to our evolved population
        population_evolved += [ (p, fitness(p, fprops)) for p in p_new ]

    return population_evolved

################################################################################
################################################################################

F = {}
F['argmin'] = -5.0
F['argmax'] = 5.0
F['dimensions'] = 3
F['resolution'] = 20
F['epsilon'] = 0.0005
F['function'] = lambda p: p[0]**2 + p[1]**2 + p[2]**2

POPULATION_SIZE     = 100
POPULATION_SIZE_NEW = 50
PR_CROSSOVER        = 0.30
PR_MUTATION         = 0.001

################################################################################

# Generate inital population and fitness
population = ga_generate(POPULATION_SIZE, F)

# Enter GA loop
for generation in range(1000):
    # Debugging: Sort and print our top 5 fitnesses
    population = sorted(population, key=lambda x: x[1])
    print generation, [x[1] for x in population[-5:]]
    sys.stdout.flush()

    # Break if the best fitness makes the epsilon
    if abs(population[-1][1]) < F['epsilon']:
        break

    # Evolve
    population = ga_evolve(population, POPULATION_SIZE_NEW, PR_CROSSOVER, PR_MUTATION, F)
    # Insert more random members to fill population to POPULATION_SIZE
    population += ga_generate(POPULATION_SIZE - len(population), F)

################################################################################

print "\nNumber of generations:", generation
print "p = ", fixed2float(population[-1][0], F), "; fitness(p) =", population[-1][1]

