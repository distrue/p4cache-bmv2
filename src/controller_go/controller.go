package main

import (
	"context"
	"fmt"
	"log"

	p4runtime "github.com/distrue/p4goruntime/p4/v1"
	grpc "google.golang.org/grpc"
	// p4config "github.com/distrue/p4gontroller/p4/config/v1";
	// "google.golang.org/grpc";
)

const (
	address = "localhost:9090"
)

/*
import grpc
from p4.v1 import p4runtime_pb2
from p4.tmp import p4config_pb2
*/

const MSG_LOG_MAX_LEN = 1024

var connections []*SwitchConnection

// def ShutdownAllSwitchConnections():
// for c in connections:
//    c.shutdown()

type SwitchConnection struct {
	name      string
	address   string
	device_id uint64
	p4info    string // temporary
	channel   *grpc.ClientConn
	// if proto_dump_file is not None:
	// interceptor = GrpcRequestLogger(proto_dump_file)
	// self.channel = grpc.intercept_channel(self.channel, interceptor)
	client_stub p4runtime.P4RuntimeClient
	// requests_stream - temporary disable
	stream_msg_resp p4runtime.P4Runtime_StreamChannelClient
	// self.stream_msg_resp = self.client_stub.StreamChannel(iter(self.requests_stream))
	ctx context.Context
}

func NewSwitchConnection(name string, address string, device_id uint64) *SwitchConnection {
	if name == "" {
		name = "client"
	}
	if address == "" {
		address = "127.0.0.1:9090"
	}
	device_id = 0
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	fmt.Println("grpc connected")

	c := p4runtime.NewP4RuntimeClient(conn)
	ctx := context.Background() // context.WithTimeout(, time.Second)
	fmt.Println("client generated")

	steamChannel, err := c.StreamChannel(ctx)
	if err != nil {
		log.Fatalf("could not Read: %v", err)
	}

	switchConnection := &SwitchConnection{
		name:            name,
		address:         address,
		device_id:       device_id,
		channel:         conn,
		client_stub:     c,
		stream_msg_resp: steamChannel,
		ctx:             ctx,
	}
	connections = append(connections, switchConnection)
	return switchConnection
}

func main() {
	swt := NewSwitchConnection("client", "127.0.0.1:9090", 0)
	cfgreq := &p4runtime.GetForwardingPipelineConfigRequest{
		DeviceId:     swt.device_id,
		ResponseType: p4runtime.GetForwardingPipelineConfigRequest_ALL,
	}
	pipeline, err := swt.client_stub.GetForwardingPipelineConfig(swt.ctx, cfgreq)
	fmt.Println("get forwarding pipeline config")
	if err != nil {
		panic(err)
	}
	info := pipeline.Config.GetP4Info()
	fmt.Println("get p4info")
	tables := info.GetTables()
	fmt.Println("get tables")
	for _, it := range tables {
		fmt.Printf(it.String())
	}
	fmt.Printf("p4controller")
}
