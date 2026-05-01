package raft

type stateFn func() stateFn

type Node struct {
	// PERSISTANT STATES
	id          uint8   // unique identifier for each node from 0 to 4
	currentTerm uint64  // lastest term a server has seen
	votedFor    int8    // id of candidate that recieved vote in current term, -1 if none
	log         []Entry // log entries

	// VOLATILE STATES
	commitIndex uint64 // index of lastest log entry known to be commited
	lastApplied uint64 // index of lastest log applied to the state machine

	// VOLATILE LEADER STATES
	nextIndex  [5]uint64
	matchIndex [5]uint64

	Transport Transport // supports RPC connection between peers
}

func MakeNode(id uint8, t Transport) Node {
	node := Node{id: id, Transport: t, votedFor: -1}
	node.initLeaderStates()
	return node
}

// init each element in nextIndex to last log index + 1
// init each element in matchIndex to 0
func (n Node) initLeaderStates() {
	for i := 0; i < 5; i++ {
		if len(n.log) == 0 {
			n.nextIndex[i] = 1
		} else {
			n.nextIndex[i] = n.log[len(n.log)-1].Index + 1
		}

		n.matchIndex[i] = 0
	}
}

func (n *Node) Run() {
	// start as follower
    state := n.follower

	// when a node transitions state, 
	// state() will return a stateFn of the next state, which will be called in the next iteration
    for state != nil {
        state = state()
    }
}

func (n *Node) setTermIfGreater(newTerm uint64) {
	if newTerm <= n.currentTerm {
		return
	}

	n.currentTerm = newTerm
	n.votedFor = -1
}

func (n *Node) handleRequestVote(req RequestVote) RequestVoteReply {
	rejectReply := RequestVoteReply{Term: n.currentTerm, Granted: false}

	// reject if candidate's term is behind
    if req.Term < n.currentTerm {
        return rejectReply
    }
	
    // reject if already voted for someone else this term
    if n.votedFor != -1 {
        return rejectReply
    }

	// reject if candidate's log is less up-to-date
    lastIndex, lastTerm := n.getLastLogData()
    if req.LastLogTerm < lastTerm {
        return rejectReply
    }
    if req.LastLogTerm == lastTerm && req.LastLogIndex < lastIndex {
        return rejectReply
    }


	n.votedFor = int8(req.CandidateID)
    return RequestVoteReply{Term: n.currentTerm, Granted: true}
}

// returns lastest log's index and term
func (n *Node) getLastLogData() (uint64, uint64) {
	if len(n.log) == 0 {
		return 0, 0
	} 
	l := n.log[len(n.log)-1]
	return l.Index, l.Term
}