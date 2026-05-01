package raft

import (
	"log"
	"time"
)

func (n *Node) leader() stateFn {
	log.Printf("Transitioned to LEADER (Term %v)", n.currentTerm)
	n.initLeaderStates()

	heartbeat := time.NewTicker(50 * time.Millisecond)
	appendReplies := make(chan AppendEntriesReply, 32)
	defer heartbeat.Stop()

	for {
		select {
		case <- heartbeat.C:
			n.sendHeartbeats(appendReplies)

		case msg := <- n.transport.Recv():
            switch payload := msg.Payload.(type) {
            case AppendEntries:
                if payload.Term > n.currentTerm {
                    n.setTermIfGreater(payload.Term)
                    msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
                    return n.follower
                }
                msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}

            case RequestVote:
                if payload.Term > n.currentTerm {
                    n.setTermIfGreater(payload.Term)
                    msg.Reply <- n.handleRequestVote(payload)
                    return n.follower
                }
                msg.Reply <- RequestVoteReply{Term: n.currentTerm, Granted: false}
			}

		case <- appendReplies:
			// TODO
		}
	}
}

func (n *Node) sendHeartbeats(appendReplies chan AppendEntriesReply) {
	index, term := n.getLastLogData()
	req := AppendEntries{Term: n.currentTerm, LeaderID: n.id, PrevLogIndex: index, PrevLogTerm: term, 
		Entries: make([]Entry, 0), LeaderCommit: n.commitIndex}

	for i := uint8(0); i < 5; i++ {
		if i == n.id {
			continue
		}

		go func() {
			reply, err := n.transport.SendAppendEntries(i, req)
			if err == nil {
				appendReplies <- reply
			}
		}()
	}
}