package raft

import (
	"log"
	"time"
)

func (n *Node) leader() stateFn {
	log.Printf("Transitioned to LEADER (Term %v)", n.currentTerm)
	n.state = Leader
	n.initLeaderStates()

	heartbeat := time.NewTicker(50 * time.Millisecond * time.Duration(n.timeMultiplier))
	appendReplies := make(chan AppendReplyWrapper, 32)
	defer heartbeat.Stop()

	for {
		if *n.Paused {
            time.Sleep(100 * time.Millisecond)
            continue
        }

		select {
		case <- heartbeat.C:
			n.sendHeartbeats(appendReplies)

		// handle RPCs
		case msg := <- n.transport.Recv():
			n.sleepForReplyDelay()
            switch payload := msg.Payload.(type) {
			// demote to follower if leader with higher term is found
            case AppendEntries:
                if payload.Term > n.currentTerm {
                    n.setTermIfGreater(payload.Term)
					n.leaderID = int8(payload.LeaderID)
                    msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}
                    return n.follower
                }
                msg.Reply <- AppendEntriesReply{Term: n.currentTerm, Success: false}

			// demote to follower if candidate with higher term is found
            case RequestVote:
                if payload.Term > n.currentTerm {
                    n.setTermIfGreater(payload.Term)
                    msg.Reply <- n.handleRequestVote(payload)
                    return n.follower
                }
                msg.Reply <- RequestVoteReply{Term: n.currentTerm, Granted: false}
			}

		// handle heartbeat replies
		case r := <- appendReplies:
			if r.Reply.Term > n.currentTerm {
				n.setTermIfGreater(r.Reply.Term)
				return n.follower
			}

			if r.Reply.Success {
				if r.LastEntryIndex != 0 {
					n.matchIndex[r.PeerID] = r.LastEntryIndex
					n.nextIndex[r.PeerID] = n.matchIndex[r.PeerID] + 1
					n.leaderAdvanceCommitIndex()				
				}

			// prevent underflow 
			} else if n.nextIndex[r.PeerID] > 1 {
				n.nextIndex[r.PeerID]--
			}

		// handle incoming client requests
		case req := <- n.clientTransport.Recv():
			n.handleClientRequest(&req) 
		}

		n.sleepForLoopDelay()
	}
}

func (n *Node) sendHeartbeats(appendReplies chan AppendReplyWrapper) {
	for i := uint8(0); i < 5; i++ {
		if i == n.id {
			continue
		}

		// send missing entries to peers
		go func(peerID uint8) {
            prevIndex := n.nextIndex[peerID] - 1
			prevTerm := uint64(0)
			if prevIndex != 0 {
				prevTerm = n.log[prevIndex-1].Term
			}

            var entries []Entry
			var lastEntryIndex uint64 = 0
            if n.nextIndex[peerID] <= uint64(len(n.log)) {
                entries = n.log[n.nextIndex[peerID]-1:]
				lastEntryIndex = entries[len(entries)-1].Index
            }

            req := AppendEntries{
                Term:         n.currentTerm,
                LeaderID:     n.id,
                PrevLogIndex: prevIndex,
                PrevLogTerm:  prevTerm,
                Entries:      entries,
                LeaderCommit: n.commitIndex,
            }

            reply, err := n.transport.SendAppendEntries(peerID, req)
            if err == nil {
                appendReplies <- AppendReplyWrapper{PeerID: peerID, Reply: reply, LastEntryIndex: lastEntryIndex}
            }
        }(i)
	}
}

// init leaderID as n.id
// init each element in nextIndex to last log index + 1
// init each element in matchIndex to 0
func (n *Node) initLeaderStates() {
	n.leaderID = int8(n.id)
	n.pendingCommits = make(map[uint64]chan bool)
	for i := 0; i < 5; i++ {
		if len(n.log) == 0 {
			n.nextIndex[i] = 1
		} else {
			n.nextIndex[i] = n.log[len(n.log)-1].Index + 1
		}

		n.matchIndex[i] = 0
	}
}

// increment commit index until there is no majority in matchIndex
func (n *Node) leaderAdvanceCommitIndex() {
    for idx := n.commitIndex + 1; idx <= uint64(len(n.log)); idx++ {
        if n.log[idx-1].Term != n.currentTerm {
            continue
        }

        count := 1
        for i := uint8(0); i < 5; i++ {
            if i != n.id && n.matchIndex[i] >= idx {
                count++
            }
        }

        if count >= 3 {
            n.commitNext()
			if ch, ok := n.pendingCommits[idx]; ok {
				ch <- true
			}
        }
    }
}