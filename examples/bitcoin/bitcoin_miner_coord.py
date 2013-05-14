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
        self.curblock_template = None
        self.coinbase_message = bin2hex("Hello World from TaskSprint!")
        self.address = "15PKyTs3jJ3Nyf3i6R7D9tfGCY1ZbtqWdv"
        self.mine_timeout = 60
        self.gbt_timeout = 15

        # Miner client list
        self.cid_miners = []

        # Miner task IDs
        self.tid_miners = []
        # getblocktemplate task ID
        self.tid_bitcoind_getblocktemplate = self.start_task(name = "bitcoind_getblocktemplate", keys = ['block_template'])
        # wait task ID
        self.tid_bitcoind_wait = None

    def client_joined(self, cid, num_nodes):
        # Add the miner to our client list
        self.cid_miners.append(cid)
        print "Client joined %d" % cid

    def client_dead(self, cid):
        # Remove the miner from our client list
        if cid in self.cid_miners:
            self.cid_miners.remove(cid)
            print "Client dead %d" % cid

    def start_miners(self, num):
        # Prepare another parameter set for a new miner
        params = {
            'block_template': self.curblock_template,
            'coinbase_message': self.coinbase_message,
            'extranonce_start': random.getrandbits(32),
            'address': self.address,
            'mine_timeout': self.mine_timeout,
            'debugnonce_start': False
        }

        for i in range(num):
            self.tid_miners.append( self.start_task(name = "mine", base = [json.dumps(params)], keys = ['mined_block', 'hps']) )

    def task_done(self, tid, values):

        # If the finished task is getblocktemplate
        if tid == self.tid_bitcoind_getblocktemplate:
            # Copy out the new template
            self.curblock_template = json.loads(values['block_template'])

            # Stop tasks mining on old template
            for t in self.tid_miners:
                self.kill_task(tid = t)
            self.tid_miners = []

            # Start tasks mining on new template
            self.start_miners(len(self.cid_miners))

            # Fire up a wait task
            self.tid_getblocktemplate = None
            self.tid_duration = self.start_task(name = "timeout", base = [self.gbt_timeout])

        # If the finished task is wait
        elif tid == self.tid_bitcoind_wait:

            # Fire up a getblocktemplate task
            self.tid_bitcoind_wait = None
            self.tid_bitcoind_getblocktemplate = self.start_task(name = "bitcoind_getblocktemplate", keys = ['block_template'])

        # If the finished task is a miner
        elif tid in self.tid_miners:

            # Remove it from the active miner task list
            self.tid_miners.remove(tid)

            print "Heard back from tid %d with hps: %d" % (tid, keys['hps'])

            # Decode the mined block
            mined_block = json.loads(keys['mined_block'])
            if mined_block != None:
                print "tid %d solved a block! Block hash:" % tid, mined_block['hash']

                # Submit the block
                self.start_task(name = "bitcoind_submit", base = [keys['mined_block']])

            # Start up a new miner task to replace this one
            self.start_miners(1)

BitcoinMinerCoordinator().start()

