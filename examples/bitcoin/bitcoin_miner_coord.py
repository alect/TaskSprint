#!/usr/bin/python -u

import json

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/coordinator")
sys.path.append("examples/bitcoin")

from TaskSprintCoordinator import *
from bitcoin_miner_lib import *

class BitcoinMinerCoordinator(TaskSprintCoordinator):
    def init(self, seed):
        # Mining parameters
        self.mining_params = {}
        self.mining_params['coinbase_message'] = bin2hex("Hello World from TaskSprint!")
        self.mining_params['address'] = "15PKyTs3jJ3Nyf3i6R7D9tfGCY1ZbtqWdv"
        self.mining_params['mine_timeout'] = 8
        self.mining_params['debugnonce_start'] = False
        # Timeout for getblocktemplate task
        self.gbt_timeout = 10

        # Miner client list
        self.cid_miners = []

        # Miner task IDs
        self.tid_miners = []
        # getblocktemplate task ID
        self.tid_getblocktemplate = self.start_task(name = "bitcoind_getblocktemplate", base = [json.dumps({'duration': 0, 'mining_params': self.mining_params})], keys = [])

    def client_joined(self, cid, num_nodes):
        # Add the miner to our client list
        self.cid_miners.append(cid)
        print "Client joined %d" % cid

    def client_dead(self, cid):
        # Remove the miner from our client list
        if cid in self.cid_miners:
            self.cid_miners.remove(cid)
            print "Client dead %d" % cid

    def task_done(self, tid, values):
        # If the finished task is getblocktemplate
        if tid == self.tid_getblocktemplate:
            # Start new clients mining on new template
            for i in range(len(self.cid_miners) - len(self.tid_miners)):
                self.tid_miners.append(self.start_task(name = "mine", prekeys = ["block_template", "mining_params"], pretasks = [self.tid_getblocktemplate]*2, keys = ["result", "hps"]))

            # Fire up another getblocktemplate task
            self.tid_getblocktemplate = self.start_task(name = "bitcoind_getblocktemplate", base = [json.dumps({'duration': self.gbt_timeout, 'mining_params': self.mining_params})], keys = [])

        # If the finished task is a miner
        elif tid in self.tid_miners:

            # Remove it from the active miner task list
            self.tid_miners.remove(tid)

            print "Heard back from miner tid %d with hps: %d" % (tid, values['hps'])

            # If the miner found something
            if values['result'] != None:
                print "tid %d solved a block! Block hash:" % (tid, values['result'])
                # Start a submit task
                self.start_task(name = "bitcoind_submit", prekeys = ["mined_block"], pretask = [tid], keys = [])

            # Start up a new miner task to replace this one
            self.tid_miners.append(self.start_task(name = "mine", prekeys = ["block_template", "mining_params"], pretasks = [self.tid_getblocktemplate]*2, keys = ["result", "hps"]))

BitcoinMinerCoordinator().start()

