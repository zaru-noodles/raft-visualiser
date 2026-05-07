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

        // single writer goroutine
		writeCh := make(chan []byte, 64)
        go func() {
            for msg := range writeCh {
                if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
                    return
                }
            }
        }()

		// send RPC events to dashboard
		go func() {
			for event := range s.Node.GetEventHistory() {
				event["kind"] = "rpc"
				data, _ := json.Marshal(event)
				writeCh <- data
			}
		}()

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		// send a summary of the node's attributes to the dashboard
		for range ticker.C {
			snapshot := s.Node.GetNodeSummary()
			snapshot["kind"] = "state"
			data, _ := json.Marshal(snapshot)
			writeCh <- data
		}
	}
}
