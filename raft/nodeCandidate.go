package raft

import "log"

func (n *Node) candidate() stateFn {
	log.Printf("Transitioned to CANDIDATE (Term %v)", n.currentTerm + 1)
	n.state = Candidate
	n.currentTerm++
	n.leaderID = -1

	// send RequestVote RPCs to peers
	index, term := n.getLastLogData()
	req := RequestVote{Term: n.currentTerm, CandidateID: n.id, LastLogIndex: index, LastLogTerm: term}
	replies := make(chan RequestVoteReply, 4)
	for i := uint8(0); i < 5; i++ {
		if i == n.id {
			continue
		}

		go func() {
			reply, err := n.transport.SendRequestVote(i, req)
			if err == nil {
				replies <- reply
			}
		}()
	}

	// count votes, if majority reached before timeout, transit to leader, else retry election
	timeout := randomElectionTimeout()
	voteCount := 1
	n.votedFor = int8(n.id)
	for {
		select {
		case <-timeout:
			return n.candidate

		// handle replies to RequestVote
		case r := <-replies:
			// if reply term is greater, stop the election and transit to follower
			if r.Term > n.currentTerm {
				n.setTermIfGreater(r.Term)
				return n.follower
			}

			if r.Granted && r.Term == n.currentTerm {
				voteCount++
				if voteCount >= 3 {
					return n.leader
				}
			}

		// handle any incoming RPCs
		case msg := <- n.transport.Recv():
			switch payload := msg.Payload.(type) {
			// convert back to follower if received AppendEntries from new leader
			case AppendEntries:
				if payload.Term >= n.currentTerm {
					n.setTermIfGreater(payload.Term)
					n.leaderID = int8(payload.LeaderID)
					msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
					return n.follower
				}
				msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}

			// deny any incoming RequestVote unless it has a higher term
			case RequestVote:
				if payload.Term > n.currentTerm {
					n.setTermIfGreater(payload.Term)
					msg.Reply <- n.handleRequestVote(payload)
					return n.follower
				}
				msg.Reply <- n.handleRequestVote(payload)
			}

		// handle incoming client requests
		case req := <- n.clientTransport.Recv():
			n.handleClientRequest(&req) 
		}
	}
}
