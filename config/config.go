package config

import (
    "os"
    "strings"
	"strconv"
	"log"
)

type Config struct {
    ID      uint8   
    Peers   []string // addresses of other nodes, e.g. ["node-1:9000", "node-2:9000"]
    RPCPort string   // port for inter-node RPCs
    WSPort  string   // port for WebSocket dashboard connection
    DataDir string   // directory for persistent state
}

func Load() Config {
	id, err := strconv.Atoi(getEnv("NODE_ID", "0"))
	if err != nil {
		id = 0
		log.Printf("Invalid NODE_ID, defaulting to 0: %v", err)
	}

    return Config{
        ID:      uint8(id),
        Peers:   strings.Split(getEnv("PEERS", ""), ","),
        RPCPort: getEnv("RPC_PORT", "9000"),
        WSPort:  getEnv("WS_PORT", "8080"),
        DataDir: getEnv("DATA_DIR", "/data"),
    }
}

func getEnv(key, fallback string) string {
    if val, present := os.LookupEnv(key); present {
        return val
    }
	log.Printf("Invalid %s, defaulting to %s", key, fallback)
    return fallback
}