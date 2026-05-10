package api

import (
	"log"
	"net/http"
)

func (s *HTTPServer) handlePause() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Paused!")
		*s.Node.Paused = true
		w.WriteHeader(200)
	}
}

func (s *HTTPServer) handleResume() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Resumed!")
		*s.Node.Paused = false
		w.WriteHeader(200)
	}
}