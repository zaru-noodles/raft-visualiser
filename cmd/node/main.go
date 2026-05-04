package main

import (
	"github.com/zaru-noodles/raft-visualiser/config"
	"github.com/zaru-noodles/raft-visualiser/http"
	"github.com/zaru-noodles/raft-visualiser/raft"
	"github.com/zaru-noodles/raft-visualiser/rpc"
)

func main() {
	cfg := config.Load()
	transport := rpc.MakeTransport(cfg.RPCPort, cfg.Peers, cfg.ID)
	clientTransport := http.MakeHTTPServer(cfg.WSPort)
	node := raft.MakeNode(cfg.ID, transport, clientTransport, cfg.DataDir)
	node.Run()
}