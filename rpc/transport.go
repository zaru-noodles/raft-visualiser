package rpc

import (
	"fmt"
	"log"

	"github.com/zaru-noodles/raft-visualiser/raft"
)

type TransportRPC struct {
	id           uint8
	peers        map[uint8]*PeerClient
	inbox        chan raft.Message
	eventHistory chan map[string]any // to be used by websockets to send RPC data to dashboard
}

func MakeTransport(RPCPort string, addrs []string, id uint8) TransportRPC {
	inbox := make(chan raft.Message, 16)
	eventHistory := make(chan map[string]any, 256)

	go startServer(inbox, RPCPort, eventHistory, id)

	// store addresses
	peers := make(map[uint8]*PeerClient)
	for i, addr := range addrs {
		// skip own id
		i := uint8(i)
		if uint8(i) >= id {
			peers[i+1] = MakePeerClient(addr)
		} else {
			peers[i] = MakePeerClient(addr)
		}
	}

	return TransportRPC{peers: peers, inbox: inbox, id: id, eventHistory: eventHistory}
}

func (t TransportRPC) SendRequestVote(peer uint8, req raft.RequestVote) (raft.RequestVoteReply, error) {
	var reply raft.RequestVoteReply

	// check if client exists
	client := t.peers[peer].GetClient()
	if client == nil {
		return reply, fmt.Errorf("no connection to peer %d", peer)
	}

	t.eventHistory <- map[string]any{"type": "request_vote", "from": t.id, "to": peer}
	err := client.Call("RaftService.RequestVote", req, &reply)
	if err != nil {
		log.Printf("Lost connection to node %v", peer)
		t.peers[peer].Reset()
	}

	return reply, err
}

func (t TransportRPC) SendAppendEntries(peer uint8, req raft.AppendEntries) (raft.AppendEntriesReply, error) {
	var reply raft.AppendEntriesReply

	// check if client exists
	client := t.peers[peer].GetClient()
	if client == nil {
		return reply, fmt.Errorf("no connection to peer %d", peer)
	}

	t.eventHistory <- map[string]any{"type": "append_entries", "from": t.id, "to": peer}
	err := client.Call("RaftService.AppendEntries", req, &reply)
	if err != nil {
		log.Printf("Lost connection to node %v", peer)
		t.peers[peer].Reset()
	}
	return reply, err
}

func (t TransportRPC) Recv() <-chan raft.Message {
	return t.inbox
}

func (t TransportRPC) GetEventHistory() <-chan map[string]any {
	return t.eventHistory
}
