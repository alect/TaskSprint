import random
import math
import bisect
import sys
import types

################################################################################
# Dynamic Fixed Width <-> Float Value Conversion Functions
################################################################################

import numpy

# Dynamic float cast depending on the resolution specified in f_params
def float_cast(f_params, *args):
    # Use Python float for resolutions <= 32, numpy float64 for resolutions > 32
    if f_params['resolution'] <= 32: FloatCast = float
    else: FloatCast = numpy.float64

    if len(args) == 1: return FloatCast(args[0])
    return [FloatCast(x) for x in args]

# Fixed Width Value to Float on f_params
def fixed2float_dynamic(x, f_params):
    amax, amin, xf, r = float_cast(f_params, f_params['argmax'], f_params['argmin'], x, 2**f_params['resolution'] - 1)
    return (amax - amin)*(xf/r) + amin

# Float to Fixed Width Value on f_params
def float2fixed_dynamic(x, f_params):
    amax, amin, xf, r = float_cast(f_params, f_params['argmax'], f_params['argmin'], x, 2**f_params['resolution'] - 1)
    return int( ((xf - amin) / (amax - amin)) * r )

# Float Epsilon to Fixed Width Epsilon on f_params
def epsilon_float2fixed_dynamic(x, f_params):
    amax, amin, xf, r = float_cast(f_params, f_params['argmax'], f_params['argmin'], x, 2**f_params['resolution'])
    return int( (x / (amax-amin)) * r )

################################################################################
# Faster Fixed Width <-> Float Value Conversion Functions
################################################################################

# Fixed Width Value to Float on f_params
def fixed2float_pyfloat(x, f_params):
    amax, amin, xf, r = float(f_params['argmax']), float(f_params['argmin']), float(x), float(2**f_params['resolution'] - 1)
    return (amax - amin)*(xf/r) + amin

# Float to Fixed Width Value on f_params
def float2fixed_pyfloat(x, f_params):
    amax, amin, xf, r = float(f_params['argmax']), float(f_params['argmin']), float(x), float(2**f_params['resolution'] - 1)
    return int( ((xf - amin) / (amax - amin)) * r )

# Float Epsilon to Fixed Width Epsilon on f_params
def epsilon_float2fixed_pyfloat(x, f_params):
    amax, amin, xf, r = float(f_params['argmax']), float(f_params['argmin']), float(x), float(2**f_params['resolution'])
    return int( (x / (amax-amin)) * r )

################################################################################
# Choose our conversion Functions (Dynamic vs All Pyfloat)
################################################################################

#fixed2float = fixed2float_pyfloat
#float2fixed = float2fixed_pyfloat
#epsilon_float2fixed = epsilon_float2fixed_pyfloat

fixed2float = fixed2float_dynamic
float2fixed = float2fixed_dynamic
epsilon_float2fixed = epsilon_float2fixed_dynamic

################################################################################
# Point Float Utility Functions
################################################################################

# Convert Fixed Width multi-dimensional point to Float multi-dimensional Point
# on f_params
#
# Returns (float_point) Float multi-dimensional point
def point_fixed2float(point, f_params):
    fl_point = []
    for p in point:
        fl_point.append(fixed2float(p, f_params))
    return fl_point

# Evaluate Euclidean distance between two points
#
# Returns (float) distance
def point_distance(p1, p2, f_params):
    p1 = point_fixed2float(p1, f_params)
    p2 = point_fixed2float(p2, f_params)

    distance = type(p1[0])(0.0)
    for i in range(f_params['dimensions']):
        distance += (p1[i] - p2[i])**2
    distance = type(p1[0])(numpy.sqrt(distance))

    return distance

# Evaluate fitness of a point on f_params function
#
# Returns (float) fitness
def point_fitness(point, f_params):
    # Calculate fitness as -abs(f(x))
    func = f_params['function']
    if  isinstance(f_params['function'], types.StringType) or \
        isinstance(f_params['function'], types.UnicodeType):
        func = eval(func)
    fitness = -1.0*abs(func(point_fixed2float(point, f_params)))

    # Set fitness large if close to known roots
    for r in f_params['solutions']:
        if point_distance(point, r, f_params) < f_params['cluster_epsilon']:
            fitness = -1.0*(2**f_params['resolution'])
            break

    return fitness

################################################################################
# Point Fixed Width Utility Functions
################################################################################

# Generate a new point
#
# Returns (point) Fixed Width multi-dimensional point
def point_generate(f_params):
    point = []
    for j in range(f_params['dimensions']):
        point.append( random.randint(0, 2**f_params['resolution']) )

    return point

# Crossover two points
#
# Returns (list) Fixed Width multi-dimesional points
def point_crossover(p1, p2, f_params, prob_crossover):
    # If crossover occurs and the points are within crossover distance
    if random.random() < prob_crossover and \
        point_distance(p1, p2, f_params) < f_params['crossover_epsilon']:
        # Average both points into one
        p3 = []
        for j in range(len(p1)):
            p3.append( (p1[j] + p2[j]) / 2 )
        return [p3]

    # Return original parents
    return [p1, p2]

# Mutate a point
#
# Returns (point) Fixed Width multi-dimensional point
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
# Core GA Evolution Function
################################################################################

# Evolve a population to a new generation, including
#   initialization, selection, crossover, mutation.
#
# Arguments:
#       population      (list) (point, fitness) tuples
#       params          (dict) function and ga parameters
#
# Returns (list) (point, fitness) tuples
def evolve(population, params):
    f_params = params['f_params']
    ga_params = params['ga_params']

    # If population is empty, initialize population with new random members
    if len(population) == 0:
        for i in range(ga_params['new_popsize']):
            p = point_generate(f_params)
            f = point_fitness(p, f_params)
            population.append( (p, f) )

    # Fill population to minimum size with new random members
    for i in range(ga_params['min_popsize'] - len(population)):
        p = point_generate(f_params)
        f = point_fitness(p, f_params)
        population.append( (p, f) )

    # Sort population by fitness
    population = sorted(population, key=lambda x: x[1])

    # Compute the target evolved population size
    target_popsize = int(ga_params['keep_percent_popsize']*len(population))

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

