package rpc

import (
	"log"
	"net/rpc"
	"net"
	
	"github.com/zaru-noodles/raft-visualiser/raft"
)

type RaftService struct {
	id           uint8
	inbox        chan raft.Message
	eventHistory chan map[string]any  // to be used by websockets to send RPC data to dashboard
}

// opens port to allow RPCs
func startServer(inbox chan raft.Message, port string, eventHistory chan map[string]any, id uint8) {
    service := &RaftService{inbox: inbox, eventHistory: eventHistory, id: id}
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

// adds a RequestVote message to the inbox channel
func (s *RaftService) RequestVote(args raft.RequestVote, reply *raft.RequestVoteReply) error {
    msg := raft.Message{
        Payload: args,
        Reply:   make(chan any, 1),
    }

    s.inbox <- msg
    *reply = (<- msg.Reply).(raft.RequestVoteReply)

	if reply.Granted {
		s.eventHistory <- map[string]any {"type": "reply_success", "from": s.id, "to": args.CandidateID}
	} else {
		s.eventHistory <- map[string]any {"type": "reply_fail", "from": s.id, "to": args.CandidateID}
	}

    return nil
}

// adds a AppendEntries message to the inbox channel
func (s *RaftService) AppendEntries(args raft.AppendEntries, reply *raft.AppendEntriesReply) error {
    msg := raft.Message{
        Payload: args,
        Reply:   make(chan any, 1),
    }

    s.inbox <- msg
    *reply = (<- msg.Reply).(raft.AppendEntriesReply)

	if reply.Success {
		s.eventHistory <- map[string]any {"type": "reply_success", "from": s.id, "to": args.LeaderID}
	} else {
		s.eventHistory <- map[string]any {"type": "reply_success", "from": s.id, "to": args.LeaderID}
	}
    return nil
}