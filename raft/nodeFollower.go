package raft

import (
	"log"
	"math/rand/v2"
	"time"
)

func (n *Node) follower() stateFn {
	timeout := randomElectionTimeout()

	for {
		select {
		// transit to candidate if heartbeat is missing 
		case <- timeout:
			log.Printf("Transitioned to CANDIDATE")
			return n.candidate

		// handle any incoming RPCs
		case msg := <- n.Transport.Recv():
			switch payload := msg.Payload.(type) {
			case AppendEntries:
				// reset timeout
				timeout = randomElectionTimeout()
				n.setTermIfGreater(payload.Term)

				// TODO: append entries

			case RequestVote:
				n.setTermIfGreater(payload.Term)
				reply := handleRequestVote(n, payload)

				// if vote is granted, reset election timer
				if reply.Granted {
					timeout = randomElectionTimeout()
				}

				msg.Reply <- reply
			} 
		}
	}
}

func randomElectionTimeout() <-chan time.Time {
	d := time.Duration(150 + rand.IntN(150)) * time.Millisecond
	return time.After(d)
}

func handleRequestVote(n *Node, req RequestVote) RequestVoteReply {
	rejectReply := RequestVoteReply{Term: n.currentTerm, Granted: false}
	// if already voted for someone, reject the request
	if n.votedFor == -1 {
		return rejectReply
	}

	n.votedFor = int8(req.CandidateID)
    return RequestVoteReply{Term: n.currentTerm, Granted: true}
}