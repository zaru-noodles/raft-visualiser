package main

import (
	"time"

	"github.com/zaru-noodles/raft-visualiser/config"
	"github.com/zaru-noodles/raft-visualiser/raft"
	"github.com/zaru-noodles/raft-visualiser/rpc"
)

func main() {
	cfg := config.Load()
	transport := rpc.MakeTransport(cfg.RPCPort, cfg.Peers, cfg.ID)
	node := raft.MakeNode(cfg.ID, transport)
	
	go clear(node)

	for {
		node.Transport.SendAppendEntries((cfg.ID + 1) % 5, raft.AppendEntries{})
		for i := 0; i < int(cfg.ID) + 1; i++ {
		    time.Sleep(time.Second)
		}
	}
}

func clear(node raft.Node) {
	for {
		for msg := range node.Transport.Recv() {
			switch msg.Payload.(type) {
			case raft.AppendEntries:
				msg.Reply <- raft.AppendEntriesReply{Term: 1, Success: true}
			case raft.RequestVote:
				msg.Reply <- raft.RequestVoteReply{Term: 1, Granted: false}
			}
		}
    }
}
