#!/bin/sh
echo This is a client
sleep 7
./client/bin/client -servers localhost:5002 -program examples/bitcoin/bitcoin_miner_node.py -socket localhost:0 -network tcp
