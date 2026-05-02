package raft

import "log"

type stateFn func() stateFn

type NodeState int

const (
    Follower NodeState = iota
    Candidate
    Leader
)

type Node struct {
	// PERSISTANT STATES
	id          uint8   // unique identifier for each node from 0 to 4
	currentTerm uint64  // lastest term a server has seen
	votedFor    int8    // id of candidate that recieved vote in current term, -1 if none
	log         []Entry // log entries

	// VOLATILE STATES
	state        NodeState       
	fsm          map[int]string  // KV store
	commitIndex  uint64          // index of lastest log entry known to be commited
	lastApplied  uint64          // index of lastest log applied to the state machine
	leaderID     int8            // ID of current leader, -1 if none

	// VOLATILE LEADER STATES
	nextIndex        [5]uint64
	matchIndex       [5]uint64
	pendingCommits   map[uint64]chan bool

	transport Transport               // supports RPC connection between peers
	clientTransport ClientTransport   // support client requests
}

func MakeNode(id uint8, t Transport, ct ClientTransport) Node {
	node := Node{
		id: id, 
		transport: t, 
		votedFor: -1, 
		leaderID: -1,
		fsm: map[int]string{},
		clientTransport: ct,
	}
	return node
}

func (n *Node) Run() {
	// start as follower
	n.state = Follower
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
	n.leaderID = -1
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

// commits entry with Index == commitIndex+1 to the KV storage
// increments commitIndex
func (n *Node) commitNext() {
	entry := n.log[n.commitIndex]
	log.Printf("Commited entry: {%v: %v} (Index: %v, Term: %v)", entry.Key, entry.Value, entry.Index, entry.Term)
	n.fsm[entry.Key] = entry.Value
	n.commitIndex++
}