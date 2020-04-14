package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

const UDP_LISTNER = 10
const TCP_CONCURRENT = 15

func main() {
	// Multicore Setup
	runtime.GOMAXPROCS(4)

	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	tcp_conn := int32(0)
	go func() {
		for {
			if tcp_conn <= TCP_CONCURRENT {
				conn, err := l.Accept()
				log.Println("new TCP connection")
				tcp_conn += 1
				if err != nil {
					log.Println(err)
					continue
				}
				go ConnHandler(conn, &tcp_conn)
			} else {
				time.Sleep(time.Duration(10) * time.Microsecond)
			}
		}
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
	}

	udp_conc := int32(0)

	go func() {
		for {
			time.Sleep(time.Duration(1) * time.Second)
			log.Printf("alive TCP connections - %v\n", tcp_conn)
			log.Printf("waiting UDP listener - %v\n", udp_conc)
		}
	}()

	for {
		if udp_conc < UDP_LISTNER {
			udp_conc += 1
			go display(udpConn, &udp_conc)
		} else {
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}
}

const GENCACHE_READ = 1
const GENCACHE_READ_REPLY = 2
const GENCACHE_WRITE = 3
const GENCACHE_WRITE_REPLY = 4
const GENCACHE_DELETE = 5
const GENCACHE_DELETE_REPLY = 6

func ConnHandler(conn net.Conn, tcp_conn *int32) {

	defer func() {
		*tcp_conn -= 1
		conn.Close()
	}()

	recvBuf := make([]byte, 4096)
	for {
		n, err := conn.Read(recvBuf)
		if nil != err {
			if io.EOF == err {
				// log.Println(err)
				return
			}
			log.Println(err)
			return
		}
		if 0 < n {
			if recvBuf[0] == GENCACHE_WRITE {
				// log.Printf("(tcp)WRITE, key: %v, seq: %v, value: %v\n", string(recvBuf[1:80]), string(recvBuf[81:85]), string(recvBuf[85:]))
				payload := make([]byte, 85)
				payload[0] = GENCACHE_WRITE_REPLY
				copy(payload[1:], recvBuf[1:80])
				copy(payload[81:], recvBuf[81:85])
				_, err = conn.Write(payload[:85])
				if err != nil {
					log.Println(err)
					return
				}
			}
			if recvBuf[0] == GENCACHE_DELETE {
				// log.Printf("(tcp)DELETE, key: %v, seq: %v, value: %v\n", string(recvBuf[1:80]), string(recvBuf[81:85]), string(recvBuf[85:]))
				payload := make([]byte, 85)
				payload[0] = GENCACHE_DELETE_REPLY
				copy(payload[1:], recvBuf[1:80])
				copy(payload[81:], recvBuf[81:85])
				_, err = conn.Write(payload[:85])
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func display(conn *net.UDPConn, udp_conc *int32) {
	defer func() { *udp_conc -= 1 }()
	buf := make([]byte, 4096)
	_, addr, err := conn.ReadFromUDP(buf) // n, addr, err
	if err != nil {
		fmt.Println("Error Reading")
		return
	}
	// log.Printf("(udp)READ, key: %v\n", string(buf[1:]))
	payload := make([]byte, 85)
	payload[0] = GENCACHE_READ_REPLY
	copy(payload[1:], buf[1:80])
	copy(payload[81:], buf[81:])
	_, err = conn.WriteToUDP(payload[:85], addr)
	if err != nil {
		log.Printf("!! %v\n", err)
		return
	}
}
