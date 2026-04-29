package raft

type Transport interface {
	SendRequestVote(peer uint8, req RequestVote) (RequestVoteReply, error)
    SendAppendEntries(peer uint8, req AppendEntries) (AppendEntriesReply, error)
    Recv() <-chan Message
}

type Message struct {
    Type    MessageType
    From    string
    Payload any
	Reply   chan any
}

type MessageType int

const (
    MsgAppendEntries  MessageType = iota
    MsgAppendEntriesReply
    MsgRequestVote
    MsgVoteReply
)

type AppendEntries struct {
    Term         uint64
    LeaderID     uint8
    PrevLogIndex uint64
    PrevLogTerm  uint64
    Entries      []Entry
    LeaderCommit uint64
}

type AppendEntriesReply struct {
    Term    uint64
    Success bool
}

type RequestVote struct {
    Term         uint64
    CandidateID  uint8
    LastLogIndex uint64
    LastLogTerm  uint64
}

type RequestVoteReply struct {
    Term    uint64
    Granted bool
}