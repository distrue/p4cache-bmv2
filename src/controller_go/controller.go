package main

import (
	"context"
	"fmt"
	"log"
	"time"

	p4runtime "github.com/distrue/p4goruntime/p4/v1"
	grpc "google.golang.org/grpc"
	// p4config "github.com/distrue/p4gontroller/p4/config/v1";
	// "google.golang.org/grpc";
)

const (
	address = "localhost:50000"
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
		address = "127.0.0.1:50000"
	}
	device_id = 0
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := p4runtime.NewP4RuntimeClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	steamChannel, err := c.StreamChannel(ctx)
	if err != nil {
		log.Fatalf("could not Read: %v", err)
	}

	/*
	  read := p4runtime.ReadRequest{};
	  readCli, err := c.Read(ctx, &read);
	  if err != nil {
	    log.Fatalf("could not Read: %v", err)
	  }
	  msg, err := readCli.Recv()
	  log.Printf("Read: %v", msg)
	*/

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

/*
@abstractmethod
    def buildDeviceConfig(self, **kwargs):
        return p4config_pb2.P4DeviceConfig()
*/

func (s *SwitchConnection) shutdown() {
	s.stream_msg_resp.CloseSend()
}

func (s *SwitchConnection) MasterArbitrationUpdate() {
	request := p4runtime.StreamMessageRequest{}
	arbitration := request.GetArbitration()
	fmt.Print(arbitration)
	// python
	// request.arbitration.device_id = self.device_id
	// request.arbitration.election_id.high = 0
	// request.arbitration.election_id.low = 1
}

func (s *SwitchConnection) SetForwardingPipelineConfig(p4info string) {
	// python
	// request = p4runtime_pb2.SetForwardingPipelineConfigRequest()
	// request.election_id.low = 1
	// request.device_id = self.device_id
	// config = request.config

	// config.p4info.CopyFrom(p4info)
	// config.p4_device_config = device_config.SerializeToString()
	// request.action = p4runtime_pb2.SetForwardingPipelineConfigRequest.VERIFY_AND_COMMIT
}

func (s *SwitchConnection) WriteTableEntry() {
	request := p4runtime.WriteRequest{
		DeviceId: s.device_id,
		// request.election_id.low = 1
	}
	update := new(p4runtime.Update)
	// TODO: change to appropriate table id
	tableEntrySrc := p4runtime.TableEntry{TableId: 1}
	// if table_entry.is_default_action:
	//   update.type = p4runtime_pb2.Update.MODIFY
	// else:
	//   update.type = p4runtime_pb2.Update.INSERT
	update.Type = p4runtime.Update_MODIFY
	//  update.Type = p4runtime.Update_INSERT
	tableEntry := p4runtime.Entity_TableEntry{TableEntry: &tableEntrySrc}
	entry := p4runtime.Entity{Entity: &tableEntry}
	update.Entity = &entry
	request.Updates = append(request.Updates, update)
	s.client_stub.Write(s.ctx, &request)
}

func (s *SwitchConnection) ReadTableEntries() {
	request := p4runtime.ReadRequest{
		DeviceId: s.device_id,
	}
	// entity = request.entities.add()
	entity := new(p4runtime.Entity)
	tableEntry := entity.GetTableEntry()
	// if table_id is not None:
	//   table_entry.table_id = table_id
	// else:
	//   table_entry.table_id = 0
	tableEntry.TableId = 0
	s.client_stub.Read(s.ctx, &request)
}

func (s *SwitchConnection) ReadCounters() {
	/*
	   def ReadCounters(self, counter_id=None, index=None, dry_run=False):
	     request = p4runtime_pb2.ReadRequest()
	     request.device_id = self.device_id
	     entity = request.entities.add()
	     counter_entry = entity.counter_entry
	     if counter_id is not None:
	         counter_entry.counter_id = counter_id
	     else:
	         counter_entry.counter_id = 0
	     if index is not None:
	         counter_entry.index.index = index
	     if dry_run:
	         print "P4Runtime Read:", request
	     else:
	         for response in self.client_stub.Read(request):
	             yield response
	*/
}

func (s *SwitchConnection) WriteMulticastGroupEntry() {
	/*
	   def WriteMulticastGroupEntry(self, mc_entry, dry_run=False):
	     request = p4runtime_pb2.WriteRequest()
	     request.device_id = self.device_id
	     request.election_id.low = 1
	     update = request.updates.add()
	     update.type = p4runtime_pb2.Update.INSERT
	     update.entity.packet_replication_engine_entry.CopyFrom(mc_entry)
	     if dry_run:
	         print "P4Runtime Write:", request
	     else:
	         self.client_stub.Write(request)
	*/
}

/*
class GrpcRequestLogger(grpc.UnaryUnaryClientInterceptor,
                        grpc.UnaryStreamClientInterceptor):
    """Implementation of a gRPC interceptor that logs request to a file"""

    def __init__(self, log_file):
        self.log_file = log_file
        with open(self.log_file, 'w') as f:
            # Clear content if it exists.
            f.write("")

    def log_message(self, method_name, body):
        with open(self.log_file, 'a') as f:
            ts = datetime.utcnow().strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]
            msg = str(body)
            f.write("\n[%s] %s\n---\n" % (ts, method_name))
            if len(msg) < MSG_LOG_MAX_LEN:
                f.write(str(body))
            else:
                f.write("Message too long (%d bytes)! Skipping log...\n" % len(msg))
            f.write('---\n')

    def intercept_unary_unary(self, continuation, client_call_details, request):
        self.log_message(client_call_details.method, request)
        return continuation(client_call_details, request)

    def intercept_unary_stream(self, continuation, client_call_details, request):
        self.log_message(client_call_details.method, request)
        return continuation(client_call_details, request)
*/

/*
class IterableQueue(Queue):
    _sentinel = object()

    def __iter__(self):
        return iter(self.get, self._sentinel)

    def close(self):
        self.put(self._sentinel)
*/

func main() {
	swt := NewSwitchConnection("client", "127.0.0.1:50000", 0)
	swt.WriteTableEntry()
	fmt.Printf("p4controller")
}
