#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

#define CONTROLLER_MIRROR_SESSION 100
#define HOT_KEY_THRESHOLD 3

#define PKT_INSTANCE_TYPE_NORMAL 0
#define PKT_INSTANCE_TYPE_INGRESS_CLONE 1
#define PKT_INSTANCE_TYPE_EGRESS_CLONE 2
#define PKT_INSTANCE_TYPE_COALESCED 3
#define PKT_INSTANCE_TYPE_INGRESS_RECIRC 4
#define PKT_INSTANCE_TYPE_REPLICATION 5
#define PKT_INSTANCE_TYPE_RESUBMIT 6

#define pkt_is_mirrored \
	((standard_metadata.instance_type != PKT_INSTANCE_TYPE_NORMAL) && \
	 (standard_metadata.instance_type != PKT_INSTANCE_TYPE_REPLICATION))

#define pkt_is_not_mirrored \
	 ((standard_metadata.instance_type == PKT_INSTANCE_TYPE_NORMAL) || \
	  (standard_metadata.instance_type == PKT_INSTANCE_TYPE_REPLICATION))


control MyEgress(inout headers hdr,
                 inout metadata meta,
                 inout standard_metadata_t standard_metadata) {

	#include "query_statistics.p4"

	// per-key counter to keep query frequency of each cached item
	counter((bit<32>) NETCACHE_ENTRIES * NETCACHE_VTABLE_NUM, CounterType.packets) query_freq_cnt;


    apply {

	}

}
