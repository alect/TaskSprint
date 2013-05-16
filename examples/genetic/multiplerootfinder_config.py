# Function, Genetic Algorithm, and Runtime Parameters
params = {
    'f_params': {
        # Weird function with 11 roots from x = -12.0 to 12.0
        'function':     "lambda x: math.sin(3*x[0])*math.cos(0.5*x[0])",
        # Input dimensions of function (int)
        'dimensions':   1,
        # Search solution space (float)
        'argmin':       -20.0,
        'argmax':       +20.0,
        # Resolution in bits (int)
        'resolution':   32,

        # Solution epsilon (float)
        'solution_epsilon':   0.001,
        # Cluster epsilon (float)
        'cluster_epsilon':     0.20,
        # Crossover epsilon (float)
        'crossover_epsilon':   7.00,
    },

    'ga_params': {
        # Total number of evolutions (int)
        'evolutions':       150,
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
        'runtime':      110,
        # Max number of GAs at any given time (int)
        'concurrent':   16,
    },
}
