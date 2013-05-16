#!/bin/sh
echo This is Coordinator 1 piped to Plotter
sleep 10
./coordinator/bin/main -servers localhost:5001,localhost:5002 -me 0 -dc examples/genetic/multiplerootfinder_coord.py | python examples/genetic/multiplerootfinder_plotter.py
