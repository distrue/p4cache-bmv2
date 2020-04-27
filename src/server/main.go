package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/distrue/gencache/src/server/gencache"
	"github.com/distrue/gencache/src/server/util"
)

const UDP_LISTNER = 10
const TCP_CONCURRENT = 15

func tcp_pool(tcp_conn *int32, l *net.TCPListener, r util.RedisClient, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if *tcp_conn <= TCP_CONCURRENT {
			conn, err := l.Accept()
			log.Println("new TCP connection")
			*tcp_conn += 1
			if err != nil {
				log.Println(err)
				continue
			}
			go gencache.TcpHandler(conn, tcp_conn, r)
		} else {
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}
}

func udp_pool(udp_conc *int32, udpConn *net.UDPConn, r util.RedisClient, wg *sync.WaitGroup, fail *uint32, succ *uint32) {
	defer wg.Done()
	for {
		if *udp_conc < UDP_LISTNER {
			*udp_conc += 1
			go gencache.UdpHandler(udpConn, udp_conc, r, fail, succ)
		} else {
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}
}

func initTCP(url string) *net.TCPListener {
	tcpAddr, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		log.Fatal(err)
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	return tcpListener
}

func initUDP(url string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
	}
	return udpConn
}

func main() {
	// Multicore Setup
	runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	wg.Add(2)

	url := ":8000"

	tcpListener := initTCP(url)
	defer tcpListener.Close()
	udpConn := initUDP(url)

	r := util.NewRedisClient()
	m := util.Monitor{}

	go tcp_pool(&m.TCP_conn, tcpListener, r, &wg)
	go udp_pool(&m.UDP_conc, udpConn, r, &wg, &m.Fail, &m.Succ)
	go gencache.ControllerReport()

	// cli report
	go util.Report(m)

	wg.Wait()
}
