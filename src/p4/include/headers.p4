#ifndef HEADERS_P4
#define HEADERS_P4


#define BLOOM_FILTER_ENTRIES 4096
#define BLOOM_IDX_WIDTH 12

#define CTRL_PORT 200

/* gencache size */
#define GENCACHE_ENTRIES 65536

/* maximum number of bits of netcache fields */
#define NETCACHE_VALUE_WIDTH_MAX 2048
#define NETCACHE_KEY_WIDTH 640

/* special reserved port for NetCache */
const bit<16> GENCACHE_PORT = 50000;
const bit<16> TYPE_IPV4 = 0x800;
const bit<8> TYPE_TCP = 0x06;
const bit<8> TYPE_UDP = 0x11;

/* current query supported types */
const bit<8> GENCACHE_READ = 0x01;
const bit<8> GENCACHE_READ_REPLY = 0x02;
const bit<8> GENCACHE_WRITE = 0x03;
const bit<8> GENCACHE_WRITE_REPLY = 0x04;
const bit<8> GENCACHE_DELETE = 0x05;
const bit<8> GENCACHE_DELETE_REPLY = 0x06;

/* netcache header field types */
typedef bit<NETCACHE_KEY_WIDTH> key_t;
typedef bit<NETCACHE_VALUE_WIDTH_MAX> value_t;

typedef bit<9>  egressSpec_t;
typedef bit<48> macAddr_t;
typedef bit<32> ip4Addr_t;

header ethernet_t {
    macAddr_t dstAddr;
    macAddr_t srcAddr;
    bit<16>   etherType;
}

header ipv4_t {
    bit<4>    version;
    bit<4>    ihl;
    bit<6>    dscp;
    bit<2>    ecn;
    bit<16>   totalLen;
    bit<16>   identification;
    bit<3>    flags;
    bit<13>   fragOffset;
    bit<8>    ttl;
    bit<8>    protocol;
    bit<16>   hdrChecksum;
    ip4Addr_t srcAddr;
    ip4Addr_t dstAddr;
}

header tcp_t{
    bit<16> srcPort;
    bit<16> dstPort;
    bit<32> seqNo;
    bit<32> ackNo;
    bit<4>  dataOffset;
    bit<4>  res;
    bit<1>  cwr;
    bit<1>  ece;
    bit<1>  urg;
    bit<1>  ack;
    bit<1>  psh;
    bit<1>  rst;
    bit<1>  syn;
    bit<1>  fin;
    bit<16> window;
    bit<16> checksum;
    bit<16> urgentPtr;
}

header tcp_options_t {
	varbit<320> options;
}

header Tcp_option_end_h {
    bit<8> kind;
}
header Tcp_option_nop_h {
    bit<8> kind;
}
header Tcp_option_ss_h {
    bit<32> maxSegmentSize;
}
header Tcp_option_s_h {
    bit<8>  kind;
    bit<8>  len;
    bit<8>  shift;
}
header Tcp_option_sack_p_h {
    bit<8>         kind;
    bit<8>         length;
}
header Tcp_option_sack_h {
    bit<8>         kind;
    bit<8>         length;
    varbit<256>    sack;
}

header Tcp_option_timestamp_h {
    bit<80> timestamp;
}

header_union Tcp_option_h {
    Tcp_option_end_h  end;
    Tcp_option_nop_h  nop;
    Tcp_option_ss_h   ss;
    Tcp_option_s_h    s;
    Tcp_option_sack_p_h sack_p;
    Tcp_option_sack_h sack;
    Tcp_option_timestamp_h ts;
}

// Defines a stack of 10 tcp options
typedef Tcp_option_h[10] Tcp_option_stack;

header Tcp_option_padding_h {
    varbit<256> padding;
}

header udp_t {
	bit<16> srcPort;
	bit<16> dstPort;
	bit<16> len;
	bit<16> checksum;
}

header gencache_t {
	bit<8> op;
	key_t key;
	bit<32> seq;
	value_t value;
}

struct fwd_metadata_t {
    bit<32> l2ptr;
    bit<24> out_bd;
}

struct metadata {
	bit<1> hot_query;

    fwd_metadata_t fwd_metadata;

    bool cache_valid;

	bit<16> tcpLength;
}

struct headers {
    ethernet_t   ethernet;
    ipv4_t       ipv4;
    tcp_t        tcp;
	tcp_options_t tcp_options;
    //Tcp_option_stack tcp_options_vec;
    //Tcp_option_padding_h tcp_options_padding;
	udp_t		 udp;
	gencache_t   gencache;
}

error {
    TcpDataOffsetTooSmall,
    TcpOptionTooLongForHeader,
    TcpBadSackOptionLength
}

struct Tcp_option_sack_top
{
    bit<8> kind;
    bit<8> length;
}

#endif   // HEADERS_P4
