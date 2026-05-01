package http

import (
	"encoding/json"
	"net/http"

	"github.com/zaru-noodles/raft-visualiser/raft"
)

type HTTPServer struct {
	Inbox chan raft.ClientRequest
	port  string
}

func MakeHTTPServer(port string) *HTTPServer {
	return &HTTPServer{Inbox: make(chan raft.ClientRequest, 4), port: port}
}

func (s *HTTPServer) StartServer() {
	// listen for POST requests on /submit
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
        var cmd raft.ClientRequest
        if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
            http.Error(w, "bad request", 400)
            return
        }
        s.Inbox <- cmd
        w.WriteHeader(200)
    })

    http.ListenAndServe(":" + s.port, nil)
}

func (s *HTTPServer) Recv() <-chan raft.ClientRequest {
	return s.Inbox
}