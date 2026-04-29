package rpc

import (
	"log"
	
	"github.com/zaru-noodles/raft-visualiser/raft"
)

type RaftService struct {
	inbox chan raft.Message
}

// adds a RequestVote message to the inbox channel
func (s *RaftService) RequestVote(args raft.RequestVote, reply *raft.RequestVoteReply) error {
	log.Print("Recieved RequestVote RPC!")

    msg := raft.Message{
        Type:    raft.MsgRequestVote,
        Payload: args,
        Reply:   make(chan any, 1),
    }

    s.inbox <- msg
    *reply = (<- msg.Reply).(raft.RequestVoteReply)
    return nil
}

// adds a AppendEntries message to the inbox channel
func (s *RaftService) AppendEntries(args raft.AppendEntries, reply *raft.AppendEntriesReply) error {
	log.Print("Recieved AppendEntries RPC!")

    msg := raft.Message{
        Type:    raft.MsgAppendEntries,
        Payload: args,
        Reply:   make(chan any, 1),
    }

    s.inbox <- msg
    *reply = (<- msg.Reply).(raft.AppendEntriesReply)
    return nil
}