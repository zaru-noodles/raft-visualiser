package api

import (
	"encoding/json"
	"net/http"

	"github.com/zaru-noodles/raft-visualiser/raft"
)

func (s* HTTPServer) handleClientRequest() func(w http.ResponseWriter, r *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		var req raft.ClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}

		// send request to inbox and await confirmation before returning
		req.Response = w
		req.Done = make(chan bool, 1)
		s.Inbox <- req
		<- req.Done
	}
}