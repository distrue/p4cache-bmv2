package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Sniffer struct {
	parser         *gopacket.DecodingLayerParser
	handle         *pcap.Handle
	datasource     *gopacket.PacketSource
	packets        chan gopacket.Packet
	tcp_connection []TCP_CONN

	ip4     layers.IPv4
	eth     layers.Ethernet
	tcp     layers.TCP
	udp     layers.UDP
	payload gopacket.Payload

	TCP_CONCURRENT_CONST uint32
	TCP_WORKER_CONST     uint32

	TCP_N_CONCURRENT uint32
	TCP_N_WORKER     uint32

	TCP_WORKER func(func(gopacket.Payload), gopacket.Payload)
	UDP_WORKER func(func(gopacket.Payload), gopacket.Payload)
}

type TCP_CONN struct {
	ipv4 layers.IPv4
	tcp  layers.TCP
	wait bool
}

func (s *Sniffer) workpool() {
	decodedLayers := []gopacket.LayerType{}

	/*
		parser_tcp := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&s.eth,
			&s.ip4,
			&s.tcp,
			// &s.udp
			&s.payload,
		)*/

	parser_udp := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&s.eth,
		&s.ip4,
		// &s.tcp,
		&s.udp,
		&s.payload,
	)

	// go func() {
	for {
		select {
		case packet := <-s.packets:
			//We are decoding layers, and switching on the layer type
			err := parser_udp.DecodeLayers(packet.Data(), &decodedLayers)
			for _, typ := range decodedLayers {
				switch typ {
				case layers.LayerTypeIPv4:
					fmt.Printf("Source ip = %s - Destination ip = %s \n", s.ip4.SrcIP.String(), s.ip4.DstIP.String())

				case layers.LayerTypeUDP:
					fmt.Println("Capture udp traffic")
					ans := s.replyPacket("udp", gopacket.Payload([]byte("answer")))
					err := s.handle.WritePacketData(ans.Bytes())
					if err != nil {
						fmt.Println(err)
					}
				}
			}

			if len(decodedLayers) == 0 {
				fmt.Println("Packet truncated")
			}

			//If the DecodeLayers is unable to decode the next layer type
			if err != nil {
				//fmt.Printf("Layer not found : %s", err)
			}
		}
	}
	// }()
	/*
		for {
			select {
			case packet := <-s.packets:
				//We are decoding layers, and switching on the layer type
				err := parser_tcp.DecodeLayers(packet.Data(), &decodedLayers)
				for _, typ := range decodedLayers {
					switch typ {
					case layers.LayerTypeIPv4:
						fmt.Printf("Source ip = %s - Destination ip = %s \n", s.ip4.SrcIP.String(), s.ip4.DstIP.String())

					case layers.LayerTypeTCP:
						//Here, we can access tcp packet properties
						fmt.Println("Capture tcp traffic")

						fmt.Printf("%v %v\n", s.ip4.DstIP, net.IP{10, 0, 0, 1})
						if reflect.DeepEqual(s.ip4.DstIP, net.IP{10, 0, 0, 1}) {
							conn, exist := s.get_conn(s.tcp_connection, s.ip4.SrcIP)
							if exist == false {
								s.add_conn(s.tcp_connection, s.ip4, s.tcp)
							} else {
								s.seq_sync(conn, s.tcp.Ack)
							}
						}

						if err != nil {
							fmt.Print(err)
						}
					}

				}

				if len(decodedLayers) == 0 {
					fmt.Println("Packet truncated")
				}

				//If the DecodeLayers is unable to decode the next layer type
				if err != nil {
					//fmt.Printf("Layer not found : %s", err)
				}
			}
		}
	*/
}

func NewSniffer(intf string) Sniffer {
	//Create parser
	sniffer := Sniffer{
		TCP_CONCURRENT_CONST: 10,
		TCP_WORKER_CONST:     30,
		TCP_N_CONCURRENT:     0,
		TCP_N_WORKER:         0,
	}

	handle, err := pcap.OpenLive(intf, 65536, true, pcap.BlockForever)
	if err != nil {
		panic("Error opening pcap: " + err.Error())
	}

	sniffer.datasource = gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	sniffer.packets = sniffer.datasource.Packets()
	sniffer.handle = handle

	return sniffer
}

func (s *Sniffer) get_conn(pool []TCP_CONN, src net.IP) (TCP_CONN, bool) {
	for _, item := range pool {
		if reflect.DeepEqual(item.ipv4.SrcIP, src) {
			return item, true
		}
	}
	return TCP_CONN{}, false
}

func (s *Sniffer) add_conn(pool []TCP_CONN, ipv4 layers.IPv4, tcp layers.TCP) {
	for {
		if s.TCP_N_CONCURRENT < s.TCP_CONCURRENT_CONST {
			newone := TCP_CONN{
				ipv4: ipv4,
				tcp:  tcp,
				wait: true,
			}
			pool = append(pool, newone)

			tcp.Ack = tcp.Seq + 1
			tcp.Seq = rand.Uint32()
			s.response("tcp", newone)(gopacket.Payload([]byte("")))
			break
		} else {
			time.Sleep(time.Duration(1) * time.Millisecond)
		}
	}
}

func (s *Sniffer) replyPacket(layer4 string, payload gopacket.Payload) gopacket.SerializeBuffer {
	fmt.Printf("answser %v \n", payload)

	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	tmp := s.eth.SrcMAC
	s.eth.SrcMAC = s.eth.DstMAC
	s.eth.DstMAC = tmp

	tmp2 := s.ip4.SrcIP
	s.ip4.SrcIP = s.ip4.DstIP
	s.ip4.DstIP = tmp2

	if layer4 == "tcp" {
		tmp3 := s.tcp.SrcPort
		s.tcp.SrcPort = s.tcp.DstPort
		s.tcp.DstPort = tmp3

		gopacket.SerializeLayers(buffer, options,
			&s.eth,
			&s.ip4,
			&s.tcp,
			payload,
		)
	}

	if layer4 == "udp" {
		tmp3 := s.udp.SrcPort
		s.udp.SrcPort = s.udp.DstPort
		s.udp.DstPort = tmp3

		gopacket.SerializeLayers(buffer, options,
			&s.eth,
			&s.ip4,
			&s.udp,
			payload,
		)
	}
	return buffer
}

func (s *Sniffer) response(layer4 string, conn TCP_CONN) func(gopacket.Payload) {
	obj := func(payload gopacket.Payload) {
		buffer := s.replyPacket(layer4, payload)
		fmt.Printf("%v\n", buffer.Bytes())
		err := s.handle.WritePacketData(buffer.Bytes())
		if err != nil {
			log.Fatalf("send error")
		}
		conn.tcp.Ack += 1
	}
	return obj
}

func (s *Sniffer) seq_sync(conn TCP_CONN, ack uint32) {
	for {
		if conn.tcp.Seq+1 == ack {
			conn.tcp.Seq += 1
			if conn.wait == true {
				conn.wait = false
			} else {
				s.TCP_WORKER(s.response("tcp", conn), s.payload)
			}
		} else {
			time.Sleep(time.Duration(1) * time.Millisecond)
		}
	}
}

func main() {
	sniffer := NewSniffer("server1-eth0")
	sniffer.TCP_WORKER = func(res func(payload gopacket.Payload), data gopacket.Payload) {
		fmt.Print(data)
		res(gopacket.Payload([]byte("answer")))
	}
	sniffer.workpool()
}
