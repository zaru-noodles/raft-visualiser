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
	tmp := &HTTPServer{Inbox: make(chan raft.ClientRequest, 4), port: port}
	go tmp.StartServer()
	return tmp
}

func (s *HTTPServer) StartServer() {
	// listen for POST requests on /submit
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
        var req raft.ClientRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "bad request", 400)
            return
        }

		req.Response = w
		req.Done = make(chan bool, 0)
        s.Inbox <- req
		<- req.Done
    })

    http.ListenAndServe(":" + s.port, nil)
}

func (s *HTTPServer) Recv() <-chan raft.ClientRequest {
	return s.Inbox
}