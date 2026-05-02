package raft

import (
	"log"
	"math/rand/v2"
	"slices"
	"time"
)

func (n *Node) follower() stateFn {
	log.Printf("Transitioned to FOLLOWER (Term %v)", n.currentTerm)
	n.state = Follower

	timeout := randomElectionTimeout()

	for {
		select {
		// transit to candidate if heartbeat is missing 
		case <- timeout:
			return n.candidate

		// handle any incoming RPCs
		case msg := <- n.transport.Recv():
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
				timeout = randomElectionTimeout()

				// reply false if log doesnt contain PrevLogIndex at the correct term
				if payload.PrevLogIndex != 0 && (int(payload.PrevLogIndex) > len(n.log) || n.log[payload.PrevLogIndex - 1].Term != payload.PrevLogTerm) {
					msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
					continue
				}

				// append to log if valid
				for _, e := range payload.Entries {
					log.Printf("Added (%v: %v) to log", e.Key, e.Value)
					if int(e.Index) > len(n.log) {
						n.log = append(n.log, e)
					} else if n.log[e.Index-1].Term != e.Term {
						n.log = slices.Delete(n.log, int(e.Index)-1, len(n.log))
						n.log = append(n.log, e)
					} 
				}

				// update commit index
				if payload.LeaderCommit > n.commitIndex {
					n.commitIndex = min(payload.LeaderCommit, uint64(len(n.log)))
					log.Printf("Commited up to index %v", n.commitIndex)
					// TODO commit to state machine
				}
				msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: true}


			case RequestVote:
				n.setTermIfGreater(payload.Term)
				reply := n.handleRequestVote(payload)
				
				// if vote is granted, reset election timer
				if reply.Granted {
					log.Printf("Voted for node %v (Term %v)", payload.CandidateID, n.currentTerm)
					timeout = randomElectionTimeout()
				}

				msg.Reply <- reply
			}

		// handle incoming client requests
		case req := <- n.clientTransport.Recv():
			n.handleClientRequest(&req) 
		}
	}
}

func randomElectionTimeout() <-chan time.Time {
	d := time.Duration(250 + rand.IntN(100)) * time.Millisecond
	return time.After(d)
}