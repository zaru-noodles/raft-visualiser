package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// upgrades from HTTP to websockets
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *HTTPServer) handleWebsocket() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Node == nil {
			http.Error(w, "node not ready", 503)
			log.Printf("node not ready")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrader err")
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		// send a summary of the node's attributes every 100ms to the dashboard
		for range ticker.C {
			snapshot := s.Node.GetNodeSummary()
			data, _ := json.Marshal(snapshot)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}
