package realtime

import (
	"encoding/json"
	"sync"
)

type Message []byte

type Hub struct {
	mu      sync.RWMutex
	clients map[chan Message]struct{}
	publish chan Message
}

func NewHub() *Hub {
	return &Hub{clients: map[chan Message]struct{}{}, publish: make(chan Message, 128)}
}
func (h *Hub) Run() {
	for msg := range h.publish {
		h.mu.RLock()
		for c := range h.clients {
			select {
			case c <- msg:
			default:
			}
		}
		h.mu.RUnlock()
	}
}
func (h *Hub) Broadcast(v any) {
	b, _ := json.Marshal(v)
	select {
	case h.publish <- b:
	default:
	}
}
func (h *Hub) Subscribe() (<-chan Message, func()) {
	c := make(chan Message, 32)
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	return c, func() {
		h.mu.Lock()
		if _, ok := h.clients[c]; ok {
			delete(h.clients, c)
			close(c)
		}
		h.mu.Unlock()
	}
}
