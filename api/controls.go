package api

import (
	"log"
	"net/http"
	"strconv"
)

func (s *HTTPServer) handlePause() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Paused!")
		s.Node.SetPaused(true)
		w.WriteHeader(200)
	}
}

func (s *HTTPServer) handleResume() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Resumed!")
		s.Node.SetPaused(false)
		w.WriteHeader(200)
	}
}

func (s *HTTPServer) handleBlock() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		peer, err := strconv.Atoi(r.URL.Path[len("/block/"):])
		if err != nil {
			http.Error(w, "invalid peer id", 400)
			return
		}

		(*s.Node.BlockedPeers)[uint8(peer)] = true
		log.Printf("Blocked peer %d", peer)
		w.WriteHeader(200)
	}
}

func (s *HTTPServer) handleUnblock() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		peer, err := strconv.Atoi(r.URL.Path[len("/unblock/"):])
		if err != nil {
			http.Error(w, "invalid peer id", 400)
			return
		}

		(*s.Node.BlockedPeers)[uint8(peer)] = false
		log.Printf("Unblocked peer %d", peer)
		w.WriteHeader(200)
	}
}