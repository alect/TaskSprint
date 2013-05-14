#!/usr/bin/python -u

import time
import json

import sys
sys.path.append("libraries/python")
sys.path.append("libraries/python/client")
sys.path.append("examples/bitcoin")

from TaskSprintNode import *
from bitcoin_miner_lib import *

class BitcoinMinerNode(TaskSprintNode):
    def mine(self, *params):
        params = json.loads(params[0])
        mined_block, hps = block_mine(params['block_template'], params['coinbase_message'], params['extranonce_start'], params['address'], timeout=params['mine_timeout'], debugnonce_start=params['debugnonce_start'])
        return {'mined_block': json.dumps(mined_block), 'hps': hps}

    def bitcoind_getblocktemplate(self):
        block_template = rpc_getblocktemplate()
        print "sending", len(json.dumps(block_template))
        return {'block_template': json.dumps(block_template)}

    def bitcoind_wait(self, *params):
        duration = params[0]
        print "I am waiting for", duration
        time.sleep(duration)
        return {}

    def bitcoind_submit(self, *params):
        mined_block = json.loads(params[0])
        submission = block_make_submit(mined_block)
        print "Client Submitting Mined Block:", submission
        rpc_submitblock(submission)
        return {}

BitcoinMinerNode().start()

