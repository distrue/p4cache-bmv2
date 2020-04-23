#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

control MyIngress(inout headers hdr,
                  inout metadata meta,
                  inout standard_metadata_t standard_metadata) {


	action drop() {
		mark_to_drop(standard_metadata);
	}

	action set_egress_port(egressSpec_t port) {
		standard_metadata.egress_spec = port;
	}


	/* Simple l2 forwarding logic */
	table l2_forward {

		key = {
			hdr.ethernet.dstAddr: exact;
		}

		actions = {
			set_egress_port;
			drop;
		}

		size = 1024;
		default_action = drop();

	}

	/* store metadata for a given key to find its values and index it properly */
	action set_lookup_metadata(last_commited seq_t) {
		// read query
		if(hdr.gencache.op == GENCACHE_READ) {
			standard_metadata.egress_spec = CTRL_PORT;
		}

		// write query
		if(hdr.gencache.op == GENCACHE_WRITE) {
			if(last_commited != hdr.gencache.seq) {
			}
		}
	}

	action set_digest() {
		// digest can only exists on ingress
		digest<bit<32>>(1, hdr.gencache.seq);
	}

	/* define cache lookup table */
	table lookup_table {
		key = {
			hdr.gencache.key : exact;
		}

		actions = {
			set_lookup_metadata;
			set_digest;
		}

		size = GENCACHE_ENTRIES;
		default_action = set_digest;
	}

	apply {
		l2_forward.apply();
	}

}
