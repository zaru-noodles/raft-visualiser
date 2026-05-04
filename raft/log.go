package raft

type Entry struct {
	Term uint64
	Index uint64
	Key int
	Value string
}

func (n *Node) makeNewEntry(i uint64, k int, v string) Entry {
	return Entry{
		Term: n.currentTerm,
		Index: i,
		Key: k,
		Value: v,
	}
}