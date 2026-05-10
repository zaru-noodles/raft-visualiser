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
	transport := rpc.MakeTransport(cfg.RPCPort, cfg.Peers, cfg.ID, &paused)
	clientTransport := api.MakeHTTPServer(cfg.WSPort)
	node := raft.MakeNode(cfg.ID, transport, clientTransport, cfg.DataDir, &paused)
	clientTransport.Node = node
	node.Run()
}