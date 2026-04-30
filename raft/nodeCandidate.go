package raft

func (n *Node) candidate() stateFn {
	return n.candidate
}