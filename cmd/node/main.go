package main

import (
	"github.com/zaru-noodles/raft-visualiser/config"
	"github.com/zaru-noodles/raft-visualiser/api"
	"github.com/zaru-noodles/raft-visualiser/raft"
	"github.com/zaru-noodles/raft-visualiser/rpc"
)

func main() {
	cfg := config.Load()
	paused := false
	blockedPeers := make([]bool, 5)

	transport := rpc.MakeTransport(cfg.RPCPort, cfg.Peers, cfg.ID, &paused, &blockedPeers)
	clientTransport := api.MakeHTTPServer(cfg.WSPort)
	node := raft.MakeNode(cfg.ID, transport, clientTransport, cfg.DataDir, &paused, &blockedPeers)
	clientTransport.Node = node
	node.Run()
}