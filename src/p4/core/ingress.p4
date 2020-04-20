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
	action set_lookup_metadata() {
		standard_metadata.egress_spec = CTRL_PORT;
	}

	/* define cache lookup table */
	table lookup_table {
		key = {
			hdr.gencache.key : exact;
		}

		actions = {
			set_lookup_metadata;
			NoAction;
		}

		size = GENCACHE_ENTRIES;
		default_action = NoAction;
	}

	apply {
		l2_forward.apply();
	}

}
