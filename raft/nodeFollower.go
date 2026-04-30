package raft

func (n *Node) follower() stateFn {
	return n.follower
}