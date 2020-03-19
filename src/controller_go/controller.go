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
	ctx context.Context
}

func NewSwitchConnection(name string, address string, device_id uint64) *SwitchConnection {
	if name == "" {
		name = "client"
	}
	if address == "" {
		address = "localhost:9090"
	}
	if device_id == 0 {
		device_id = 1
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	fmt.Println("grpc connected", conn.GetState())

	c := p4runtime.NewP4RuntimeClient(conn)
	ctx := context.Background()
	fmt.Println("client generated")

	switchConnection := &SwitchConnection{
		name:        name,
		address:     address,
		device_id:   device_id,
		channel:     conn,
		client_stub: c,
		ctx:         ctx,
	}
	connections = append(connections, switchConnection)

	// grpc work test
	cfgreq := &p4runtime.GetForwardingPipelineConfigRequest{
		DeviceId:     device_id,
		ResponseType: p4runtime.GetForwardingPipelineConfigRequest_ALL,
	}
	out := new(p4runtime.SetForwardingPipelineConfigResponse)
	conn.Invoke(ctx, "/p4.v1.P4Runtime/SetForwardingPipelineConfig", cfgreq, out)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	return switchConnection
}

func main() {
	swt := NewSwitchConnection("client", "localhost:9090", 1)
	cfgreq := &p4runtime.GetForwardingPipelineConfigRequest{
		DeviceId:     swt.device_id,
		ResponseType: p4runtime.GetForwardingPipelineConfigRequest_ALL,
	}
	fmt.Println(cfgreq.String())
	pipeline, err := swt.client_stub.GetForwardingPipelineConfig(swt.ctx, cfgreq)
	if err != nil {
		panic(err)
	}
	fmt.Println("get forwarding pipeline config")
	info := pipeline.Config.GetP4Info()
	fmt.Println("get p4info")
	tables := info.GetTables()
	fmt.Println("get tables")
	for _, it := range tables {
		fmt.Printf(it.String())
	}
	fmt.Printf("p4controller")
}
