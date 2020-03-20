package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	p4config "github.com/distrue/p4goruntime/p4/config/v1"
	p4runtime "github.com/distrue/p4goruntime/p4/v1"
	grpc "google.golang.org/grpc"
)

const (
	address = "localhost:9090"
)

const MSG_LOG_MAX_LEN = 1024

func GetSwitchConnection(name string, address string, device_id uint64) p4runtime.P4RuntimeClient {
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

	return p4runtime.NewP4RuntimeClient(conn)
}

func OpenStreamListener(stream p4runtime.P4Runtime_StreamChannelClient) sync.WaitGroup {

	var waitg sync.WaitGroup
	waitg.Add(1)

	go func() {
		for {
			inData, err := stream.Recv()
			if err == io.EOF {
				waitg.Done()
				return
			}
			if err != nil {
				fmt.Printf("[STREAM-ERROR] (%T) : %+v\n", err, err)
			}
			if inData != nil {
				fmt.Printf("[STREAM-INCOMING] (%T) : %+v\n", inData, inData)

			}
			// Act on the received message
		}
	}()

	return waitg
}

func SetMastership(stream p4runtime.P4Runtime_StreamChannelClient) {
	req := p4runtime.StreamMessageRequest{
		Update: &p4runtime.StreamMessageRequest_Arbitration{
			Arbitration: &p4runtime.MasterArbitrationUpdate{
				DeviceId: 1,
				Role: &p4runtime.Role{
					Id: 2,
				},
				ElectionId: &p4runtime.Uint128{High: 10000, Low: 9999},
			},
		},
	}

	err := stream.Send(&req)
	if err != nil {
		fmt.Println("ERROR SENDING STREAM REQUEST:")
		fmt.Println(err)
	}
}

func GetPipelineConfigs(client p4runtime.P4RuntimeClient) (*p4config.P4Info, error) {
	getReq := &p4runtime.GetForwardingPipelineConfigRequest{
		DeviceId:     1,
		ResponseType: p4runtime.GetForwardingPipelineConfigRequest_ALL,
	}
	reply, err := client.GetForwardingPipelineConfig(context.Background(), getReq)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	return reply.GetConfig().P4Info, nil
}

func main() {
	client := GetSwitchConnection("client", "localhost:9090", 1)
	stream, sErr := client.StreamChannel(context.Background())
	if sErr != nil {
		fmt.Println(sErr)
		log.Fatalf("cannot open stream channel with the server")
	}

	listener := OpenStreamListener(stream)

	SetMastership(stream)
	newConfig, err := GetPipelineConfigs(client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", newConfig)

	fmt.Printf("p4controller")
	listener.Wait()
}
