#!/usr/bin/python -u

import time
import json
import random

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/client")
sys.path.append("examples/bitcoin")

from TaskSprintNode import *
from bitcoin_miner_lib import *

class BitcoinMinerNode(TaskSprintNode):
    def mine(self, *params):
        block_template = json.loads(params[0])
        mining_params = json.loads(params[1])

        print "Starting mining with params:", mining_params

        mined_block, hps = block_mine(block_template, mining_params['coinbase_message'], random.getrandbits(32), mining_params['address'], timeout=mining_params['mine_timeout'], debugnonce_start=mining_params['debugnonce_start'])

        print "Done mining!"

        if mined_block == None: result = None
        else: result = mined_block['hash']

        return {'result': result, 'mined_block': json.dumps(mined_block), 'hps': hps}

    def bitcoind_getblocktemplate(self, *params):
        params = json.loads(params[0])

        time.sleep(params['duration'])
        print "Getting new block template..."

        block_template = rpc_getblocktemplate()
        return {'block_template': json.dumps(block_template), 'mining_params': json.dumps(params['mining_params'])}

    def bitcoind_submit(self, *params):
        mined_block = json.loads(params[0])
        submission = block_make_submit(mined_block)
        print "Client Submitting Mined Block:", submission
        rpc_submitblock(submission)
        return {}

BitcoinMinerNode().start()

