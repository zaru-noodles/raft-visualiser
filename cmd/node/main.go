package main

import (
	"github.com/zaru-noodles/raft-visualiser/config"
	"github.com/zaru-noodles/raft-visualiser/raft"
	"github.com/zaru-noodles/raft-visualiser/rpc"
)

func main() {
	cfg := config.Load()
	transport := rpc.MakeTransport(cfg.RPCPort, cfg.Peers, cfg.ID)
	node := raft.MakeNode(cfg.ID, transport)
	node.Run()
}