package gencache

import (
	"io"
	"log"
	"net"

	"github.com/distrue/gencache/src/server/util"
)

const GENCACHE_WRITE = 3
const GENCACHE_WRITE_REPLY = 4
const GENCACHE_DELETE = 5
const GENCACHE_DELETE_REPLY = 6

func TcpHandler(conn net.Conn, tcp_conn *int32, r util.RedisClient) {

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
				err = r.SetItem(string(recvBuf[1:80]), string(recvBuf[85:]))
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
