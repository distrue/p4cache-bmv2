from random import shuffle, random

import numpy as np
import argparse

DATA_DIR='data/'
REFRESH_INTERVAL = 10000 # change hot item per one interval pasts
HOT_CHANGE = 200
ZIPF = 0.99

def main(n_servers, n_queries, dynamics):
    keys = []
    sample = []

    alpha = 1.0 / (1.0 - ZIPF)

    # adds all generated keys to the set of keys to sample from
    for i in range(1, 1+int(n_servers)):
        with open(DATA_DIR + 'server' + str(i) + '.txt') as f:
            content = f.readlines()
        content = [x.strip().split('=')[0] for x in content]
        keys.extend(content)

    # shuffle keys
    shuffle(keys)

    # draw random query items
    while len(sample) < int(n_queries):
        # smaller value is hotter value
        query_index = np.random.zipf(alpha, 1)[0]
        if dynamics == 1 and ch != len(sample) / REFRESH_INTERVAL:
            shuffle(keys)
        ch = len(sample) / REFRESH_INTERVAL
        if query_index <= len(keys):
            if dynamics == 0:
                tgt = query_index - 1
                if(tgt < ch * HOT_CHANGE):
                    tgt = len(keys) - 1 - tgt
            if dynamics == 2:
                tgt = query_index - 1 + ch * HOT_CHANGE
                tgt %= len(keys)
            else:
                tgt = query_index - 1
            sample.append(keys[tgt])


    sample_file = '{}zipf_dynamic_sample_{}_{}.txt'.format(DATA_DIR, n_queries, str(dynamics))

    with open(sample_file, 'w') as f:
    	for query_item in sample:
            f.write("read {}\n".format(query_item))



def check_valid_dynamic(value):
    ivalue = int(value)
    if ivalue >= 3 or ivalue <= -1:
        raise argparse.ArgumentTypeError("value should be 0, 1, or 2")
    return ivalue

if __name__=="__main__":

    parser = argparse.ArgumentParser()

    parser.add_argument('--n-servers', help='number of servers', type=int, required=True)
    parser.add_argument('--n-queries', help='number of queries to generate', type=int, required=True)
    parser.add_argument('--dynamics', help='dynamic type of the workload (0: hot-in, 1: random, 2: hot-out)', type=check_valid_dynamic,
            required=False, default=0)
    args = parser.parse_args()

    main(args.n_servers, args.n_queries, args.dynamics)
