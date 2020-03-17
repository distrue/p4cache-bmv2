#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

control MyEgress(inout headers hdr,
                 inout metadata meta,
                 inout standard_metadata_t standard_metadata) {

	#include "query_statistics.p4"

	// per-key counter to keep query frequency of each cached item
	counter((bit<32>) NETCACHE_ENTRIES * NETCACHE_VTABLE_NUM, CounterType.packets) query_freq_cnt;


    apply {

	}

}
