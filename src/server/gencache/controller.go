package gencache

import (
	"fmt"
	"log"
	"net"
	"time"
)

const GENCACHE_HOT_REPORT = 7

const UPDATE_INTERVAL = 1

func initUDP(url string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println(err)
	}
	return udpConn
}

func inttob(val int) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}

// send on UDP
func ControllerReport() {
	// IP reserved for Controller
	conn := initUDP("10.0.0.10")

	for {
		time.Sleep(time.Duration(UPDATE_INTERVAL) * time.Second)

		ans := TopN()
		for _, item := range ans {
			// fmt.Printf("%v - %v\n", item.Id, item.Val)

			payload := make([]byte, 85) // 1byte + 80byte + 4byte
			payload[0] = byte(GENCACHE_HOT_REPORT)
			copy(payload[1:], item.Id)
			copy(payload[81:], inttob(item.Val))
			conn.Write(payload)
			// send HOT REPORT PACKET
		}
	}
}
