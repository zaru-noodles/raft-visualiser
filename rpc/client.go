package rpc

import (
	"net/rpc"
	"sync"
)

type PeerClient struct {
	client *rpc.Client
	addr string
	dialing bool
	ready chan any
	mu sync.Mutex
}

func MakePeerClient(addr string) *PeerClient{
	return &PeerClient{client:nil, addr: addr, dialing: false}
}

// if client is nil, only one call of GetClient is to dial the server at a time
func (p *PeerClient) GetClient() *rpc.Client{
	p.mu.Lock()

	// if client exists, return it
	if p.client != nil {
		tmp := p.client
		p.mu.Unlock()
		return tmp
	}

	// if another call to GetClient() is already dialing, await its result
	if p.dialing {
		ready := p.ready
		p.mu.Unlock()
		
		// return result of dialing
		<- ready
		p.mu.Lock()
		tmp := p.client
		p.mu.Unlock()
		return tmp
	}

	// if no other GetClient() is dialing, attempt to dial the server
	p.dialing = true
	p.ready = make(chan any)
	p.mu.Unlock()
	
	client, err := rpc.Dial("tcp", p.addr)
	
	p.mu.Lock()
	if err == nil {
		p.client = client
	}
	p.dialing = false
	close(p.ready)
	p.mu.Unlock()

	return p.client
}

// resets the PeerClient to nil
func (p *PeerClient) Reset() {
    p.mu.Lock()
    if p.client != nil {
        p.client.Close()
    }
    p.client = nil
    p.mu.Unlock()
}