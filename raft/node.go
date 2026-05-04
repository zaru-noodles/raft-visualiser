package raft

// stateFn is a function type that represents the state of a node in the Raft algorithm.

type stateFn func() stateFn

type Node struct {
	// PERSISTANT STATES
	id uint8          // unique identifier for each node from 0 to 4
	currentTerm uint64 // lastest term a server has seen
	votedFor int8    // id of candidate that recieved vote in current term, -1 if none
	log []Entry     // log entries
	
	// VOLATILE STATES
	commitIndex uint64 // index of lastest log entry known to be commited
	lastApplied uint64 // index of lastest log applied to the state machine
	
	// VOLATILE LEADER STATES
	nextIndex [4]uint64
	matchIndex [4]uint64
}

func NewNode() Node {
	node := Node{}
	node.votedFor = -1
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
			n.nextIndex[i] = n.log[len(n.log) - 1].Index + 1
		}
		
		n.matchIndex[i] = 0
	}
}