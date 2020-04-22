package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
)

const UDP_LISTNER = 10
const TCP_CONCURRENT = 15

type Redis struct {
	client *redis.Client
}

func RedisClient() Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return Redis{
		client,
	}
}

func (c Redis) getItem(key string) (string, error) {
	val, err := c.client.Get(key).Result()
	if err == redis.Nil {
		return "", errors.New("No item found")
	} else if err != nil {
		panic(err)
	}
	return val, nil
}

func (c Redis) setItem(key string, val interface{}) error {
	return c.client.Set(key, val, 0).Err()
}

func tcp_pool(tcp_conn *int32, l *net.TCPListener, r Redis, wg *sync.WaitGroup) {
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
			go ConnHandler(conn, tcp_conn, r)
		} else {
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}
}

func udp_pool(udp_conc *int32, udpConn *net.UDPConn, r Redis, wg *sync.WaitGroup, fail *uint32, succ *uint32) {
	defer wg.Done()
	for {
		if *udp_conc < UDP_LISTNER {
			*udp_conc += 1
			go display(udpConn, udp_conc, r, fail, succ)
		} else {
			time.Sleep(time.Duration(10) * time.Microsecond)
		}
	}
}

func main() {
	// Multicore Setup
	runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	wg.Add(2)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	udpAddr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
	}

	tcp_conn := int32(0)
	udp_conc := int32(0)
	fail := uint32(0)
	succ := uint32(0)

	r := RedisClient()

	go tcp_pool(&tcp_conn, listener, r, &wg)
	go udp_pool(&udp_conc, udpConn, r, &wg, &fail, &succ)

	// reporting per sec
	go func() {
		for {
			time.Sleep(time.Duration(5) * time.Second)
			log.Printf("")
			fmt.Printf("alive TCP connections - %v\n", tcp_conn)
			fmt.Printf("waiting UDP listener - %v\n", udp_conc)
			fmt.Printf("failed UDP req - %v\n", fail)
			fmt.Printf("successed UDP req - %v\n", succ)
			succ = 0
			fail = 0
		}
	}()

	wg.Wait()
}

const GENCACHE_READ = 1
const GENCACHE_READ_REPLY = 2
const GENCACHE_WRITE = 3
const GENCACHE_WRITE_REPLY = 4
const GENCACHE_DELETE = 5
const GENCACHE_DELETE_REPLY = 6
const GENCACHE_HOT_REPORT = 7    // don't reach to endpoint
const GENCACHE_HOST_RESPONSE = 8 // don't reach to endpoint

func ConnHandler(conn net.Conn, tcp_conn *int32, r Redis) {

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
				err = r.setItem(string(recvBuf[1:80]), string(recvBuf[85:]))
				if err != nil {
					log.Println(err)
					return
				}

				payload := make([]byte, 10)
				payload[0] = GENCACHE_WRITE_REPLY
				copy(payload[1:], []byte("success"))
				_, err = conn.Write(payload)
				if err != nil {
					log.Println(err)
					return
				}
			}
			if recvBuf[0] == GENCACHE_DELETE {
				// log.Printf("(tcp)DELETE, key: %v, seq: %v, value: %v\n", string(recvBuf[1:80]), string(recvBuf[81:85]), string(recvBuf[85:]))
				// TODO: erase item

				payload := make([]byte, 85)
				payload[0] = GENCACHE_DELETE_REPLY
				copy(payload[1:], []byte("success"))
				_, err = conn.Write(payload)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func display(conn *net.UDPConn, udp_conc *int32, r Redis, fail *uint32, succ *uint32) {
	defer func() { *udp_conc -= 1 }()
	buf := make([]byte, 4096)
	_, addr, err := conn.ReadFromUDP(buf) // n, addr, err
	if err != nil {
		fmt.Println("Error Reading")
		return
	}
	val, err := r.getItem(string(buf[1:80]))
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
	*succ += 1
}
