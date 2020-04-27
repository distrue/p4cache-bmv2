#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

control MyIngress(inout headers hdr,
                  inout metadata meta,
                  inout standard_metadata_t standard_metadata) {

	register<bit<8>>(CUCKOO_SEQ_ENTRIES) dirtySet1;
	register<bit<8>>(CUCKOO_SEQ_ENTRIES) dirtySet2;
	register<bit<8>>(CUCKOO_SEQ_ENTRIES) dirtySet3;
	register<bit<8>>(CUCKOO_SEQ_ENTRIES) dirtySet4;

	action drop() {
        mark_to_drop(standard_metadata);
    }
    
    action ipv4_forward(macAddr_t dstAddr, egressSpec_t port) {
        standard_metadata.egress_spec = port;
        hdr.ethernet.srcAddr = hdr.ethernet.dstAddr;
        hdr.ethernet.dstAddr = dstAddr;
        hdr.ipv4.ttl = hdr.ipv4.ttl - 1;
    }
    
    table ipv4_lpm {
        key = {
            hdr.ipv4.dstAddr: lpm;
        }
        actions = {
            ipv4_forward;
            drop;
            NoAction;
        }
        size = 1024;
        default_action = drop();
    }

	/* store metadata for a given key to find its values and index it properly */
	action set_lookup_metadata(bit<32> last_commited) {
		meta.last_commited = last_commited;
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
		ipv4_lpm.apply();
		switch(lookup_table.apply().action_run) {

			set_lookup_metadata: {
				// read query
				if(hdr.gencache.op == GENCACHE_READ) {
					standard_metadata.egress_spec = CTRL_PORT;
				}

				// write query
				if(hdr.gencache.op == GENCACHE_WRITE) {
					hash(meta.fingerprint, HashAlgorithm.crc16_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<16>) 255);
					
					hash(meta.dirtySet_idx1, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					
					hash(meta.dirtySet_idx2, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					
					hash(meta.dirtySet_idx3, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					
					hash(meta.dirtySet_idx4, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);

					bit<8> val_1;
					dirtySet1.read(val_1, (bit<32>) meta.dirtySet_idx1);
					bit<8> val_2;
					dirtySet1.read(val_2, (bit<32>) meta.dirtySet_idx2);
					bit<8> val_3;
					dirtySet1.read(val_3, (bit<32>) meta.dirtySet_idx3);
					bit<8> val_4;
					dirtySet1.read(val_4, (bit<32>) meta.dirtySet_idx4);
					
					if(!(val_1 == 1 && val_2 == 1 && val_3 == 1 && val_4 == 1)) { // retransmission
						if(meta.last_commited != hdr.gencache.seq) { // overwritted 
							// fallback to user
							// TODO: ingress port, ipv4, ethernet swap
						} else { // not overwritted
							// send to controller
							standard_metadata.egress_spec = CTRL_PORT;
						}
					}
					else {
						// send to controller
						standard_metadata.egress_spec = CTRL_PORT;
					}
				}

				if(hdr.gencache.op == GENCACHE_WRITE_REPLY) {
					hash(meta.fingerprint, HashAlgorithm.crc16_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<16>) 255);
					
					hash(meta.dirtySet_idx1, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					bit<8> val_1;
					dirtySet1.read(val_1, (bit<32>)meta.dirtySet_idx1);
					if(val_1 == meta.fingerprint) {
						dirtySet1.write((bit<32>)meta.dirtySet_idx1, 0);
					}

					hash(meta.dirtySet_idx2, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					bit<8> val_2;
					dirtySet1.read(val_2, (bit<32>)meta.dirtySet_idx2);
					if(val_2 == meta.fingerprint) {
						dirtySet2.write((bit<32>)meta.dirtySet_idx2, 0);
					}

					hash(meta.dirtySet_idx3, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					bit<8> val_3;
					dirtySet1.read(val_3, (bit<32>)meta.dirtySet_idx3);
					if(val_3 == meta.fingerprint) {
						dirtySet3.write((bit<32>)meta.dirtySet_idx3, 0);
					}

					hash(meta.dirtySet_idx4, HashAlgorithm.crc32_custom, (bit<1>) 0, { hdr.gencache.seq }, (bit<32>) CUCKOO_SEQ_ENTRIES);
					bit<8> val_4;
					dirtySet1.read(val_4, (bit<32>)meta.dirtySet_idx4);
					if(val_4 == meta.fingerprint) {
						dirtySet4.write((bit<32>)meta.dirtySet_idx4, 0);
					}
				}
			}
		}
	}

}
