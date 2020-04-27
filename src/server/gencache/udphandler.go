package gencache

import (
	"fmt"
	"log"
	"net"

	"github.com/distrue/gencache/src/server/util"
)

const GENCACHE_READ = 1
const GENCACHE_READ_REPLY = 2

const GENCACHE_ADDCACHE_REQ = 9
const GENCACHE_ADDCACHE_FETCH = 10

func UdpHandler(conn *net.UDPConn, udp_conc *int32, r util.RedisClient, fail *uint32, succ *uint32) {
	defer func() { *udp_conc -= 1 }()
	buf := make([]byte, 4096)
	_, addr, err := conn.ReadFromUDP(buf) // n, addr, err
	if err != nil {
		fmt.Println("Error Reading")
		return
	}
	switch buf[0] {
	case GENCACHE_READ:
		val, err := r.GetItem(string(buf[1:80]))
		var payload []byte
		if err != nil {
			*fail += 1
			return
		}
		payload = make([]byte, 1024)
		payload[0] = GENCACHE_READ_REPLY
		copy(payload[1:], buf[1:81])
		copy(payload[81:], val)

		_, err = conn.WriteToUDP(payload, addr)
		if err != nil {
			log.Printf("!! %v\n", err)
			return
		}
		CountItem(string(buf[1:80]))
		*succ += 1

	case GENCACHE_ADDCACHE_REQ:
		val, err := r.GetItem(string(buf[1:80]))
		var payload []byte
		if err != nil {
			*fail += 1
			return
		}
		payload = make([]byte, 1024)
		payload[0] = GENCACHE_ADDCACHE_FETCH
		copy(payload[1:], buf[1:81])
		copy(payload[81:], val)

		_, err = conn.WriteToUDP(payload, addr)
		if err != nil {
			log.Printf("!! %v\n", err)
			return
		}
		CountItem(string(buf[1:80]))
		*succ += 1
	}
}
