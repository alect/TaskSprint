import sys
sys.path.append("../")

import unittest
from ga_multiplerootfinder_lib import *

################################################################################
# Unit Tests
################################################################################

class TestConversions(unittest.TestCase):
    def assertEqualsFloat(self, a, b, e):
        self.assertTrue( abs(a-b) < e )

    def test_fixed2float(self):
        f_params = {}
        f_params['argmax'] = 5.0
        f_params['argmin'] = -5.0
        f_params['resolution'] = 10

        e = 1.0 / (2**(f_params['resolution']/2))

        self.assertEqualsFloat(fixed2float(0, f_params), -5.0, e)
        self.assertEqualsFloat(fixed2float(2**9 - 1, f_params), 0.0, e)
        self.assertEqualsFloat(fixed2float(2**10 - 1, f_params), 5.0, e)

    def test_float2fixed(self):
        f_params = {}
        f_params['argmax'] = 5.0
        f_params['argmin'] = -5.0
        f_params['resolution'] = 10

        self.assertEquals(float2fixed(-5.0, f_params), 0)
        self.assertEquals(float2fixed(0.0, f_params), 2**9 - 1)
        self.assertEquals(float2fixed(5.0, f_params), 2**10 - 1)

    def test_epsilon_float2fixed(self):
        f_params = {}
        f_params['argmax'] = 5.0
        f_params['argmin'] = -5.0
        f_params['resolution'] = 10

        self.assertEquals(epsilon_float2fixed(0.0, f_params), 0)
        self.assertEquals(epsilon_float2fixed(10.0, f_params), 2**10)
        self.assertEquals(epsilon_float2fixed(5.0, f_params), 2**9)


class TestPoint(unittest.TestCase):
    def assertEqualsFloat(self, a, b, e):
        self.assertTrue( abs(a-b) < e )

    def test_point_distance(self):
        f_params = {}
        f_params['argmax'] = 5.0
        f_params['argmin'] = -5.0
        f_params['resolution'] = 32
        f_params['dimensions'] = 3

        p1 = [float2fixed(1.0, f_params), float2fixed(2.0, f_params), float2fixed(3.0, f_params)]
        p2 = [float2fixed(4.0, f_params), float2fixed(5.0, f_params), float2fixed(6.0, f_params)]
        # Distance = sqrt( (4-1)^2 + (5-2)^2 + (6-3)^2 ) = 5.196

        self.assertEqualsFloat(point_distance(p1, p2, f_params), 5.1961524, 1e-5)

    def test_point_fitness(self):
        f_params = {}
        f_params['argmax'] = 5.0
        f_params['argmin'] = -5.0
        f_params['resolution'] = 10
        f_params['function'] = lambda x: x[0]**2 + x[1]**2 + x[2]**2
        f_params['solutions'] = []

        e = 1.0 / (2**(f_params['resolution']/2))

        # p = [-5.0, -5.0, -5.0]
        self.assertEqualsFloat(point_fitness([0, 0, 0], f_params), -75.0, e)
        # p = [0.0, 0.0, 0.0]
        self.assertEqualsFloat(point_fitness([511, 511, 511], f_params), -0.0, e)
        # p = [+5.0, +5.0, +5.0]
        self.assertEqualsFloat(point_fitness([1023, 1023, 1023], f_params), -75.0, e)

if __name__ == "__main__":
    unittest.main()

