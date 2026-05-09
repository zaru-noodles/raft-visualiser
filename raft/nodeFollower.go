package raft

import (
	"log"
	"math/rand/v2"
	"time"
)

func (n *Node) follower() stateFn {
	log.Printf("Transitioned to FOLLOWER (Term %v)", n.currentTerm)
	n.state = Follower

	timeout := n.randomElectionTimeout()

	for {
		select {
		// transit to candidate if heartbeat is missing 
		case <- timeout:
			return n.candidate

		// handle any incoming RPCs
		case msg := <- n.transport.Recv():
			n.sleepForReplyDelay()
			switch payload := msg.Payload.(type) {
			case AppendEntries:
				// reply false if RPC came from old leader
				if payload.Term < n.currentTerm {
					msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
					continue
				}

				n.leaderID = int8(payload.LeaderID)
				n.setTermIfGreater(payload.Term)

				// reset timeout
				timeout = n.randomElectionTimeout()

				// reply false if log doesnt contain PrevLogIndex at the correct term
				if payload.PrevLogIndex != 0 && (int(payload.PrevLogIndex) > len(n.log) || n.log[payload.PrevLogIndex - 1].Term != payload.PrevLogTerm) {
					msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
					continue
				}

				// append to log if valid
				for _, e := range payload.Entries {
					if int(e.Index) > len(n.log) {
						log.Printf("Added {%v: %v} to log (Index: %v, Term: %v)", e.Key, e.Value, e.Index, e.Term)
						n.appendLogEntry(e)
					} else if n.log[e.Index-1].Term != e.Term {
						log.Printf("Modified and added {%v: %v} log (Index: %v, Term: %v)", e.Key, e.Value, e.Index, e.Term)
						n.deleteLogEntry(e.Index)
						n.appendLogEntry(e)
					} 
				}

				// update commit index
				for n.commitIndex < min(payload.LeaderCommit, uint64(len(n.log))) {
					n.commitNext()
				}
				msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: true}

			case RequestVote:
				n.setTermIfGreater(payload.Term)
				reply := n.handleRequestVote(payload)
				
				// if vote is granted, reset election timer
				if reply.Granted {
					log.Printf("Voted for node %v (Term %v)", payload.CandidateID, n.currentTerm)
					timeout = n.randomElectionTimeout()
				}

				msg.Reply <- reply
			}

		// handle incoming client requests
		case req := <- n.clientTransport.Recv():
			n.handleClientRequest(&req) 
		}

		n.sleepForLoopDelay()
	}
}

func (n *Node) randomElectionTimeout() <-chan time.Time {
	d := time.Duration((250 + rand.IntN(100)) * n.timeMultiplier) * time.Millisecond
	return time.After(d)
}