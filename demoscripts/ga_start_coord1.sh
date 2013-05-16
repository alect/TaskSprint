#!/bin/sh
echo This is Coordinator 1
sleep 10
./coordinator/bin/main -servers localhost:5001,localhost:5002 -me 0 -dc examples/genetic/multiplerootfinder_coord.py
