package rpc

import (
	"net/rpc"
	"net"
	"log"
	"time"

	"github.com/zaru-noodles/raft-visualiser/raft"
)

type TransportRPC struct {
	peers map[uint8]*rpc.Client
	inbox chan raft.Message
}

func MakeTransport(RPCPort string, addrs []string, id uint8) TransportRPC {
	inbox := make(chan raft.Message, 16)

	go startServer(inbox, RPCPort)

	// connects to peers' RPC server
	peers := make(map[uint8]*rpc.Client)
	for i, addr := range addrs {
		// skip own id
		i := uint8(i)
		if uint8(i) >= id {
			peers[i+1] = dialWithRetry(addr)
		} else {
		    peers[i] = dialWithRetry(addr)
		}
	}
	log.Print("Connected!")

	return TransportRPC{ peers: peers, inbox: inbox } 
}

// attempts to dial address indefinitely
func dialWithRetry(address string) *rpc.Client {
    for {
        client, err := rpc.Dial("tcp", address)
        if err == nil {
            return client
        }
        log.Printf("Retrying connection to %s...", address)
        time.Sleep(time.Second)
    }
}

// opens port to allow RPC
func startServer(inbox chan raft.Message, port string) {
    service := &RaftService{inbox: inbox}
    rpc.Register(service)

    listener, err := net.Listen("tcp", ":" + port)
    if err != nil {
        log.Fatal("UNABLE TO START SERVER:", err)
    }

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go rpc.ServeConn(conn)
    }
}

func (t TransportRPC) SendRequestVote(peer uint8, req raft.RequestVote) (raft.RequestVoteReply, error) {
	var reply raft.RequestVoteReply
	log.Print("Sending RequestVote RPC...")
    err := t.peers[peer].Call("RaftService.RequestVote", req, &reply)
	log.Print("Recieved RequestVoteReply!")
    return reply, err
}

func (t TransportRPC) SendAppendEntries(peer uint8, req raft.AppendEntries) (raft.AppendEntriesReply, error) {
	var reply raft.AppendEntriesReply
	log.Print("Sending AppendEntries RPC...")
    err := t.peers[peer].Call("RaftService.AppendEntries", req, &reply)
	log.Print("Recieved AppendEntriesReply!")
    return reply, err
}

func (t TransportRPC) Recv() <-chan raft.Message {
	return t.inbox
}