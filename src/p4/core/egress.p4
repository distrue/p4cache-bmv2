#include <core.p4>
#include <v1model.p4>

#include "../include/headers.p4"

control MyEgress(inout headers hdr,
                 inout metadata meta,
                 inout standard_metadata_t standard_metadata) {

    //Digest<key_t>() query_digest;

    apply {
        //query_digest.pack(hdr.gencache.key);
        digest<key_t>(1, {hdr.gencache.key});
	}

}
