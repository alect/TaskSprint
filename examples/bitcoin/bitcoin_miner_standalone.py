from bitcoin_miner_lib import *

################################################################################
# Standalone Bitcoin Miner, Single-threaded
################################################################################

def standalone_miner(coinbase_message, address):
    while True:
        print "Mining new block template..."
        mined_block, hps = block_mine(rpc_getblocktemplate(), coinbase_message, 0, address, 15)
        print "Average Hashes Per Second:", hps

        if mined_block != None:
            print "Solved a block! Block hash:", mined_block['hash']
            submission = block_make_submit(mined_block)
            print "Submitting:", submission
            rpc_submitblock(submission)

standalone_miner(bin2hex("Hello from vsergeev"), "15PKyTs3jJ3Nyf3i6R7D9tfGCY1ZbtqWdv")

