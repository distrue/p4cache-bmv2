from client_api import NetCacheClient

import time
import numpy as np


def main(n_servers, disable_cache, suppress, input_files):
    client = NetCacheClient(n_servers=n_servers, no_cache=disable_cache)
    total_start = time.time()

    for filepath in input_files:
        query = 0
        with open(filepath) as fp:
            line = fp.readline()
            cnt = line.split(' ')
            while line:
                query += 1
                if(cnt[0] == 'write'):
                    client.put(cnt[1].strip(), 'new', suppress=suppress)
                else:
                    client.read(cnt[1].strip(), suppress=suppress)
                
                line = fp.readline()  
                cnt = line.split(' ')          

        if disable_cache:
            x = 'nocache'
        else:
            x = 'netcache'

        input_file = filepath.split('/')[1].split('.')[0]

        out_file = 'results/{}_{}_{}_client.txt'.format(input_file, n_servers, x)
        out_fd = open(out_file, 'w')

        spend_time = time.time() - total_start
        client.request_metrics_report(output=out_fd, time=spend_time, query=query)


if __name__=="__main__":

    import argparse
    parser = argparse.ArgumentParser()

    parser.add_argument('--n-servers', help='number of servers', type=int, required=False, default=1)
    parser.add_argument('--disable-cache', help='disable in-network caching', action='store_true')
    parser.add_argument('--suppress', help='suppress output', action='store_true')
    parser.add_argument('--input', help='input files to execute queries', required=True, nargs="+")
    args = parser.parse_args()

    main(args.n_servers, args.disable_cache, args.suppress, args.input)
