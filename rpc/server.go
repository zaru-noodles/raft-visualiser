package rpc

import (
	"log"
	"net/rpc"
	"net"
	
	"github.com/zaru-noodles/raft-visualiser/raft"
)

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