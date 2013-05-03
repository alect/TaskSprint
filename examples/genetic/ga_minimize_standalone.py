import random
import math
import numpy
import pylab
import bisect
import sys

# Convert a fixed point representation to a float
def fixed2float(point, fprops):
    x = point[0]
    y = point[1]
    x = (fprops['varmax']-fprops['varmin'])*(float(x) / (2**fprops['res'])) + fprops['varmin']
    y = (fprops['varmax']-fprops['varmin'])*(float(y) / (2**fprops['res'])) + fprops['varmin']
    return (x,y)

# Evaluate the fitness of a point on function fprops['func']
def fitness(point, fprops):
    (x,y) = fixed2float(point, fprops)
    return fprops['func'](x,y)

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
    p = []
    f = []

    for i in range(n):
        x = random.randint(0, 2**fprops['res'])
        y = random.randint(0, 2**fprops['res'])
        point = (x,y)

        p.append(point)
        f.append(fitness(point, fprops))

    return zip(p,f)

#
# Evolve a population to a new generation,
# including selection, crossover, mutation.
#
# args  pf              (list) population, fitness tuples
#       n               (int) size of new population
#       pr_crossover    (float) crossover probabiltiy
#       pr_mutation     (float) mutation probability
#       fprops          (list) function properties
#
# ret   (list) (population, fitness) tuples
def ga_evolve(pf, n, pr_crossover, pr_mutation, fprops):
    # Sort population by increasing fitness
    pf = sorted(pf, key=lambda x: x[1])[::-1]

    def crossover(p1, p2):
        # If crossover occurs, average both points into one
        if (random.random() < pr_crossover):
            p3 = ( (p1[0] + p2[0]) / 2, (p1[1] + p2[1]) / 2)
            return [p3]

        # Return original parents
        return [p1, p2]

    def mutate(p):
        for i in range(fprops['res']):
            # If mutation occurs, flip a bit
            if (random.random() < pr_mutation):
                p = (p[0] ^ 2**i, p[1])
            # If mutation occurs, flip a bit
            if (random.random() < pr_mutation):
                p = (p[0], p[1] ^ 2**i)

        return p

    p = []
    f = []

    while len(p) < n:
        # Choose new member p1 by elitist selection
        i = len(pf)-1
        p1 = pf[i][0]
        del pf[i]

        # Choose new member p2 by elitist selection
        i = len(pf)-1
        p2 = pf[i][0]
        del pf[i]

        # Crossover p1 and p2
        new_members = crossover(p1, p2)
        # Mutate children
        new_members = [mutate(m) for m in new_members]

        # Add them to our pool
        p += new_members
        f += [fitness(m, fprops) for m in new_members]

    return zip(p, f)

################################################################################
################################################################################

fprops = {}
fprops['varmin'] = -5.0
fprops['varmax'] = 5.0
fprops['res'] = 20
fprops['func'] = lambda x,y: x**2 + y**2

population_size = 100
population_new = 0.50
pr_crossover = 0.30
pr_mutation = 0.02

################################################################################

# Generate inital population and fitness
pf = ga_generate(population_size, fprops)

# Enter GA loop
for i in range(1000):
    # Sort and print our top 5 fitnesses
    pf = sorted(pf, key=lambda x: x[1])
    print i, [x[1] for x in pf[0:5]]
    sys.stdout.flush()
    if pf[0][1] < 0.00005: break

    # Evolve
    pf = ga_evolve(pf, int(population_new*len(pf)), pr_crossover, pr_mutation, fprops)
    # Insert more random members to fill population
    pf += ga_generate(population_size - len(pf), fprops)

################################################################################

print "\nNumber of generations:", i
print "(x,y) = ", fixed2float(pf[0][0], fprops), "; f(x,y) =", fitness(pf[0][0], fprops)

