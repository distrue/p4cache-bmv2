package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
	//	"runtime"
)

const GENCACHE_READ = 1
const GENCACHE_READ_REPLY = 2
const GENCACHE_WRITE = 3
const GENCACHE_WRITE_REPLY = 4
const GENCACHE_DELETE = 5
const GENCACHE_DELETE_REPLY = 6
const GENCACHE_HOT_REPORT = 7    // don't reach to endpoint
const GENCACHE_HOST_RESPONSE = 8 // don't reach to endpoint

func cache_read(udpconn *net.UDPConn) {
	read_payload := make([]byte, 85) // 4bit(<1byte) + 640bit(=80byte) + 32bit(=4byte) = 85byte
	read_payload[0] = byte(GENCACHE_READ)
	copy(read_payload[1:], []byte("max length 640bit key"))
	copy(read_payload[81:], []byte("NONE"))
	udpconn.Write([]byte(read_payload))
}

func cache_write(conn net.Conn) { // conn net.Conn
	write_payload := make([]byte, 1120) // 85byte + 1024byte(item, temporary)
	write_payload[0] = byte(GENCACHE_WRITE)
	copy(write_payload[1:], []byte("max length 640bit key"))
	copy(write_payload[81:], []byte("hash"))
	copy(write_payload[85:], []byte("max length 1024byte value"))
	conn.Write([]byte(write_payload))

	recv := make([]byte, 4096)
	_, err := conn.Read(recv) // n, err
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println("Server send(tcp) : " + string(recv[:n]))
}

func cache_delete(conn net.Conn) { // conn net.Conn
	delete_payload := make([]byte, 85) // 4bit(<1byte) + 640bit(=80byte) + 32bit(=4byte) = 85byte
	delete_payload[0] = byte(GENCACHE_DELETE)
	copy(delete_payload[1:], []byte("max length 640bit key"))
	copy(delete_payload[81:], []byte("hash"))
	conn.Write([]byte(delete_payload))

	recv := make([]byte, 4096)
	_, err := conn.Read(recv) // n, err
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println("Server send(tcp) : " + string(recv[:n]))
}

func main() {
	// Multicore Setup
	runtime.GOMAXPROCS(4)

	conn, ch := net.Dial("tcp", "10.0.0.1:8000")
	if ch != nil {
		log.Println(ch)
	}
	defer conn.Close()

	s, err := net.ResolveUDPAddr("udp4", ":8000")
	if err != nil {
		log.Println(err)
	}
	udpconn, err := net.DialUDP("udp4", nil, s)

	if err != nil {
		log.Println(err)
	}
	defer udpconn.Close()

	count := 0
	tcpcount := 0

	go func() {
		data := make([]byte, 4096)
		for {
			_, _, err := udpconn.ReadFromUDP(data) // n, addr, err
			if err != nil {
				fmt.Println(err)
				return
			}
			count += 1
			// fmt.Printf("Server send(udp) : %v\n", string(data[0:n]))

			// disable client read lock
			// time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(1) * time.Second)
			fmt.Printf("UDP count - %v\n", count)
			count = 0
			fmt.Printf("TCP count - %v\n", tcpcount)
			tcpcount = 0
		}
	}()

	go func() {
		for {
			cache_read(udpconn)
		}
	}()

	// TCP send
	for {
		cache_write(conn)
		tcpcount += 1
		cache_delete(conn)
		tcpcount += 1
		// time.Sleep(time.Duration(1) * time.Millisecond)
	}
}
