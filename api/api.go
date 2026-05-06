package api

import (
	"net/http"

	"github.com/zaru-noodles/raft-visualiser/raft"
)

type HTTPServer struct {
	Inbox chan raft.ClientRequest
	port  string
	Node *raft.Node
}

func MakeHTTPServer(port string) *HTTPServer {
	tmp := &HTTPServer{Inbox: make(chan raft.ClientRequest, 4), port: port}
	go tmp.StartServer()
	return tmp
}

func (s *HTTPServer) StartServer() {
	// listen for POST requests on /submit
	http.HandleFunc("/submit", corsMiddleware(s.handleClientRequest()))
	http.HandleFunc("/ws", s.handleWebsocket())
    http.ListenAndServe(":" + s.port, nil)
}

func (s *HTTPServer) Recv() <-chan raft.ClientRequest {
	return s.Inbox
}

// allows requests from dashboard
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(204)
            return
        }

        next(w, r)
    }
}
