package raft

type Entry struct {
	Term uint64
	Index uint64
	Op string
	Key int
	Value int
}