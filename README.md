# Raft Visualiser

A Go implementation of a Raft cluster with a browser dashboard for live visualization and control.

The project runs a multi-node Raft cluster where each node exposes RPC, dashboard control, and WebSocket state streams. The dashboard displays node connectivity, leader election, term changes, RPC traffic, and allows pausing nodes or partitioning links.

<p align="center">
    <img width="868" height="508" alt="Recording 2026-08-09 200837" src="https://github.com/user-attachments/assets/80240f53-0453-4f2a-b19a-aa1bb6e5bf3e" />
</p>

## Features

- 5-node Raft cluster by default
- Node roles: leader, follower, candidate 
- Live cluster dashboard over WebSocket
- Submit key/value commands to the current leader
- Pause/resume individual nodes
- Block/unblock peer connections to simulate partitions
- Persistent node state using container volumes

<p align="center">
   <img width="868" height="508" alt="Recording 2026-08-09 201423" src="https://github.com/user-attachments/assets/aae7b8ba-3581-4069-b8a7-b3a440ab1c0a" />
   <p align="center">The classic split-brain scenario</p>
</p>

## Repository structure

- `cmd/node/main.go` - node process entrypoint
- `config/config.go` - environment config loader
- `api/` - HTTP/WebSocket server for dashboard control and client requests
- `raft/` - Raft node implementation and internal state
- `rpc/` - inter-node transport layer
- `dashboard/` - static front-end assets for cluster visualization
- `Dockerfile` - build container image for a Raft node
- `docker-compose.yml` - launch 5 nodes + dashboard service

## Prerequisites

- Go 1.26+
- Docker and Docker Compose

## Running with Docker Compose

From the repository root:

```powershell
docker compose up --build
```

This starts:

- `dashboard` on `http://localhost:3000`
- `node-0` on `http://localhost:8080`
- `node-1` on `http://localhost:8081`
- `node-2` on `http://localhost:8082`
- `node-3` on `http://localhost:8083`
- `node-4` on `http://localhost:8084`

Each node also listens on port `9000` internally for Raft RPC traffic.

## Dashboard usage

Open the dashboard at `http://localhost:3000`.

Dashboard controls include:

- Pause/resume individual nodes
- Cut and restore connections between nodes
- Submit key/value command pairs to the current leader
- View event logs, leader elections, and connection state

## API Endpoints

Each node exposes dashboard control endpoints on its configured `WS_PORT`:

- `POST /submit` - submit a client command
- `POST /pause` - pause the node
- `POST /resume` - resume the node
- `POST /block/{peer}` - block traffic to a peer
- `POST /unblock/{peer}` - unblock traffic to a peer
- `GET  /ws` - WebSocket stream for state and RPC events

## Notes

- The dashboard connects to node WebSocket endpoints on `localhost:8080` through `localhost:8084`.
- Network partition simulation is handled by per-node block/unblock endpoints.
- The project is intended for exploration and demonstration of Raft behavior rather than production use.
- AI was used to design and style the dashboard.
