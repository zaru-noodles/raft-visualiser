package raft

func (n *Node) leader() stateFn {
	return n.leader
}