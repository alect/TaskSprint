
from boto.s3.connection import S3Connection
from boto.s3.key import Key
import pickle

aws_access_key_id = 'AKIAIX5D67EBPQUFFHPQ'
aws_secret_access_key = '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'

def get(bucket, key):
    k = Key(bucket)
    k.key = key
    return pickle.loads(k.get_contents_as_string())

def put(bucket, key, value):
    k = Key(bucket)
    k.key = key
    k.set_contents_from_string(pickle.dumps(value))

if __name__ == '__main__':
    conn = S3Connection(aws_access_key_id, aws_secret_access_key)
    bucket = conn.get_bucket('alec-storage')
    # Try a put
    put(bucket, 'test_key', { 'value1' : 22, 'value2' : [2, 2, 2, 2, 2] })
    print str(get(bucket, 'test_key'))
    
     
    
