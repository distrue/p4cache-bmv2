#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

control MyIngress(inout headers hdr,
                  inout metadata meta,
                  inout standard_metadata_t standard_metadata) {

	register<bit<16>>(KEY_ENTRIES) lookup1;
	register<bit<16>>(KEY_ENTRIES) lookup2;
	register<bit<32>>(KEY_ENTRIES) last_commit;
	register<bit<32>>(DIRTYSET_ENTRIES) dirtyset1;
	register<bit<32>>(DIRTYSET_ENTRIES) dirtyset2;
	register<bit<32>>(DIRTYSET_ENTRIES) dirtyset3;
	register<bit<32>>(DIRTYSET_ENTRIES) dirtyset4; 

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

	bit<16> lookup_now1;
	bit<16> lookup_now2;
	bit<32> seq_lcommit;
	bit<32> hash_result;
	bit<32> seq_dset1;
	bit<32> seq_dset2;
	bit<32> seq_dset3;
	bit<32> seq_dset4;
	bit<1> write_ctrl;
	bit<1> write_retms;
	bit<48> tmpEth;

	apply {
		ipv4_lpm.apply();
		if(hdr.gencache.op == GENCACHE_READ) {
			lookup1.read(lookup_now1, hdr.gencache.key);
			lookup2.read(lookup_now2, hdr.gencache.key);
			if(lookup_now1 != (bit<16>) 0) {
				standard_metadata.egress_spec = CTRL_PORT;
			}
			if(lookup_now2 != (bit<16>) 0) {
				standard_metadata.egress_spec = CTRL_PORT;
			}
		}

		// GENCACHE_READ_REPLY

		if(hdr.gencache.op == GENCACHE_WRITE || hdr.gencache.op == GENCACHE_DELETE) {
			write_ctrl = (bit<1>) 0;
			lookup1.read(lookup_now1, hdr.gencache.key);
			lookup2.read(lookup_now2, hdr.gencache.key);
			if(lookup_now1 != (bit<16>) 0) {
				write_ctrl = (bit<1>) 1;
			}
			if(lookup_now2 != (bit<16>) 0) {
				write_ctrl = (bit<1>) 1;
			}
			if(write_ctrl == (bit<1>) 1) {
				hash(meta.lcommit, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
				last_commit.read(seq_lcommit, meta.lcommit);

				if(seq_lcommit != 0 && seq_lcommit != hdr.gencache.seq) {
					hash(meta.dsetidx1, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
					hash(meta.dsetidx2, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
					hash(meta.dsetidx3, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
					hash(meta.dsetidx4, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
					
					dirtyset1.read(seq_dset1, meta.dsetidx1);
					dirtyset2.read(seq_dset2, meta.dsetidx2);
					dirtyset3.read(seq_dset3, meta.dsetidx3);
					dirtyset4.read(seq_dset4, meta.dsetidx4);
					
					write_retms = (bit<1>) 0;
					if(seq_dset1 == hdr.gencache.seq) {
						write_retms = (bit<1>) 1;
						dirtyset1.write(0, meta.dsetidx1);
					}
					if(seq_dset2 == hdr.gencache.seq) {
						write_retms = (bit<1>) 1;
						dirtyset2.write(0, meta.dsetidx2);
					}
					if(seq_dset3 == hdr.gencache.seq) {
						write_retms = (bit<1>) 1;
						dirtyset3.write(0, meta.dsetidx3);
					}
					if(seq_dset4 == hdr.gencache.seq) {
						write_retms = (bit<1>) 1;
						dirtyset4.write(0, meta.dsetidx4);
					}
					if(write_retms == (bit<1>) 1) {
						hdr.gencache.op = GENCACHE_WRITE_REPLY;
						tmpEth = hdr.ethernet.srcAddr;
						hdr.ethernet.srcAddr = hdr.ethernet.dstAddr;
						hdr.ethernet.dstAddr = tmpEth;
				        standard_metadata.egress_spec = standard_metadata.ingress_port;
					}
					else {
						standard_metadata.egress_spec = CTRL_PORT;
					}
				}
				else {
					standard_metadata.egress_spec = CTRL_PORT;
				}
			}
		}

		if(hdr.gencache.op == GENCACHE_WRITE_REPLY || hdr.gencache.op == GENCACHE_DELETE_REPLY || hdr.gencache.op == GENCACHE_ADDCACHE_EVICT) {
			if(hdr.gencache.op == GENCACHE_DELETE_REPLY || hdr.gencache.op == GENCACHE_ADDCACHE_EVICT) {
				lookup1.write( 0, hdr.gencache.key);
				lookup2.write( 0, hdr.gencache.key);

				hash(meta.lcommit, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
				last_commit.write((bit<32>)0, meta.lcommit);
			}

			if(hdr.gencache.op == GENCACHE_ADDCACHE_EVICT) {
		        mark_to_drop(standard_metadata);
			}

			hash(meta.dsetidx1, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx2, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx3, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx4, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			
			dirtyset1.read(seq_dset1, meta.dsetidx1);
			dirtyset2.read(seq_dset2, meta.dsetidx2);
			dirtyset3.read(seq_dset3, meta.dsetidx3);
			dirtyset4.read(seq_dset4, meta.dsetidx4);
			
			if(seq_dset1 == hdr.gencache.seq) {
				dirtyset1.write(0, meta.dsetidx1);
			}
			if(seq_dset2 == hdr.gencache.seq) {
				dirtyset2.write(0, meta.dsetidx2);
			}
			if(seq_dset3 == hdr.gencache.seq) {
				dirtyset3.write(0, meta.dsetidx3);
			}
			if(seq_dset4 == hdr.gencache.seq) {
				dirtyset4.write(0, meta.dsetidx4);
			}
		}

		if(hdr.gencache.op == GENCACHE_WRITE_CACHE || hdr.gencache.op == GENCACHE_DELETE_CACHE) {
			hash(meta.lcommit, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			last_commit.write(hdr.gencache.seq, meta.lcommit);

			hash(meta.dsetidx1, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx2, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx3, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx4, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			
			dirtyset1.write(hdr.gencache.seq, meta.dsetidx1);
			dirtyset2.write(hdr.gencache.seq, meta.dsetidx2);
			dirtyset3.write(hdr.gencache.seq, meta.dsetidx3);
			dirtyset4.write(hdr.gencache.seq, meta.dsetidx4);
		}

		// GENCACHE_ADDCACHE_REPORT
		// GENCACHE_ADDCACHE_REQ
		// --> already set to CTRL_PORT

		if(hdr.gencache.op == GENCACHE_ADDCACHE_FETCH) {
			lookup1.write( 1, hdr.gencache.key);
			lookup2.write( 1, hdr.gencache.key);				
			
			hash(meta.lcommit, HashAlgorithm.crc32_custom, (bit<1>) 0,
					{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			last_commit.write((bit<32>)0, meta.lcommit);

			hash(meta.dsetidx1, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx2, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx3, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			hash(meta.dsetidx4, HashAlgorithm.crc32_custom, (bit<1>) 0,
			{ hdr.gencache.key }, (bit<16>) DIRTYSET_ENTRIES);
			
			dirtyset1.write(hdr.gencache.seq, meta.dsetidx1);
			dirtyset2.write(hdr.gencache.seq, meta.dsetidx2);
			dirtyset3.write(hdr.gencache.seq, meta.dsetidx3);
			dirtyset4.write(hdr.gencache.seq, meta.dsetidx4);

			mark_to_drop(standard_metadata);
		}
	}

}
