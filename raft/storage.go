package raft

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PersistentStorage struct {
	n       *Node
	path    string
}

type PersistentState struct {
	CurrentTerm uint64  // lastest term a server has seen
	VotedFor    int8    // id of candidate that recieved vote in current term, -1 if none
	Log         []Entry // log entries
}

func makeStorage(n *Node, dataDir string) *PersistentStorage {
	return &PersistentStorage{n: n, path: filepath.Join(dataDir, "state.json")}
}

func (p *PersistentStorage) Save() error {
	state := PersistentState{
		CurrentTerm: p.n.currentTerm,
		VotedFor:    p.n.votedFor,
		Log:         p.n.log,
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(p.path, data, 0644)
}

func (p *PersistentStorage) Load() error {
    data, err := os.ReadFile(p.path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // first boot, no state to load
        }
        return err
    }

    var state PersistentState
    if err := json.Unmarshal(data, &state); err != nil {
        return err
    }

    p.n.currentTerm = state.CurrentTerm
    p.n.votedFor = state.VotedFor
    p.n.log = state.Log
    return nil
}