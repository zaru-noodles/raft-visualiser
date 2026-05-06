package raft

import (
	"fmt"
	"log"
	"net/http"
)

func (n *Node) handleClientRequest(req *ClientRequest) {
	writer := req.Response

	switch n.state {
	case Follower, Candidate:
		// do not handle the request, redirect to leader if exists
		log.Printf("Recieved client request, redirecting to node %v", n.leaderID)
		if n.leaderID != -1 {
			writer.Header().Set("Location", fmt.Sprintf("http://node-%d:8080/submit", n.leaderID))
			http.Error(writer, "not leader", 307)
		} else {
			http.Error(writer, "no leader present", 503)
		}
		req.Done <- true

	case Leader:
		// add request to log
		log.Printf("Recieved client request, processing...")
		i, _ := n.getLastLogData()
		i++
		entry := n.makeNewEntry(i, req.Key, req.Value)
		n.log = append(n.log, entry)
		n.pendingCommits[i] = make(chan bool, 1)
		log.Printf("Added (%v: %v) to log", req.Key, req.Value)
		
		// await for the request to be commited
		go func() {
			if <- n.pendingCommits[i] {
				writer.WriteHeader(200)
			} else {
				http.Error(writer, "failed to commit", 500)
			}
			delete(n.pendingCommits, i)
			req.Done <- true
		}()
	}
}