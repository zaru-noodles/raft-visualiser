package raft

import (
	"log"
	"math/rand/v2"
	"time"
)

func (n *Node) follower() stateFn {
	log.Printf("Transitioned to FOLLOWER (Term %v)", n.currentTerm)
	timeout := randomElectionTimeout()

	for {
		select {
		// transit to candidate if heartbeat is missing 
		case <- timeout:
			return n.candidate

		// handle any incoming RPCs
		case msg := <- n.Transport.Recv():
			switch payload := msg.Payload.(type) {
			case AppendEntries:
				// reset timeout
				timeout = randomElectionTimeout()
				n.setTermIfGreater(payload.Term)

				// TODO: append entries
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
		}
	}
}

func randomElectionTimeout() <-chan time.Time {
	d := time.Duration(250 + rand.IntN(100)) * time.Millisecond
	return time.After(d)
}