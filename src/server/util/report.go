package util

import (
	"fmt"
	"log"
	"time"
)

type Monitor struct {
	TCP_conn int32
	UDP_conc int32
	Fail     uint32
	Succ     uint32
}

func Report(m Monitor) {
	for {
		time.Sleep(time.Duration(5) * time.Second)
		log.Printf("")
		fmt.Printf("alive TCP connections - %v\n", m.TCP_conn)
		fmt.Printf("waiting UDP listener - %v\n", m.UDP_conc)
		fmt.Printf("failed UDP req - %v\n", m.Fail)
		fmt.Printf("successed UDP req - %v\n", m.Succ)
		m.Succ = 0
		m.Fail = 0
	}
}
