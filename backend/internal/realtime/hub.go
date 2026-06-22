// Package realtime implements an in-process WebSocket hub. Clients subscribe to
// a company channel; publishers broadcast JSON messages to all subscribers of a
// company. The hub is safe for concurrent use and has no database dependency, so
// it can be constructed once and shared across handlers.
package realtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Message is a single outbound realtime frame.
type Message struct {
	Type      string    `json:"type"`       // e.g. "tick", "event", "decision", "funding"
	CompanyID string    `json:"company_id"`
	Payload   any       `json:"payload,omitempty"`
	SentAt    time.Time `json:"sent_at"`
}

// client is one subscribed WebSocket connection.
type client struct {
	companyID string
	conn      *websocket.Conn
	send      chan []byte
}

// Hub fans out published messages to all subscribers of a company.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	// joinLeave is buffered so Register/Unregister never block the caller.
	joinLeave chan joinLeave
	broadcast chan Message
	closed    chan struct{}
}

type joinLeave struct {
	c     *client
	leave bool
}

// NewHub creates a running Hub. It must be started once; Shutdown stops it.
func NewHub() *Hub {
	h := &Hub{
		clients:  make(map[*client]struct{}),
		joinLeave: make(chan joinLeave, 64),
		broadcast: make(chan Message, 256),
		closed:   make(chan struct{}),
	}
	go h.run()
	return h
}

// run is the single goroutine that mutates the client set and writes to
// connections, so no per-client locking is needed.
func (h *Hub) run() {
	for {
		select {
		case jl := <-h.joinLeave:
			if jl.leave {
				delete(h.clients, jl.c)
				close(jl.c.send)
			} else {
				h.clients[jl.c] = struct{}{}
			}
		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			for c := range h.clients {
				if c.companyID != msg.CompanyID {
					continue
				}
				// Non-blocking send: drop a slow client rather than block the hub.
				select {
				case c.send <- data:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
		case <-h.closed:
			return
		}
	}
}

// Publish broadcasts a message to all subscribers of msg.CompanyID. It never
// blocks: if the broadcast buffer is full the message is dropped (the next tick
// will re-publish fresh state).
func (h *Hub) Publish(msg Message) {
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	select {
	case h.broadcast <- msg:
	default:
	}
}

// Subscribe registers a WebSocket connection to a company channel and pumps
// outbound messages to it. It blocks until ctx is canceled or the connection
// closes.
func (h *Hub) Subscribe(ctx context.Context, companyID string, conn *websocket.Conn) {
	c := &client{
		companyID: companyID,
		conn:      conn,
		send:      make(chan []byte, 32),
	}
	h.joinLeave <- joinLeave{c: c, leave: false}
	defer func() {
		h.joinLeave <- joinLeave{c: c, leave: true}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-c.send:
			if !ok {
				return
			}
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				return
			}
		}
	}
}

// SubscriberCount returns the number of clients subscribed to a company. Useful
// for tests and observability.
func (h *Hub) SubscriberCount(companyID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for c := range h.clients {
		if c.companyID == companyID {
			n++
		}
	}
	return n
}

// Shutdown stops the hub run loop. Subsequent Publish calls are no-ops.
func (h *Hub) Shutdown() {
	select {
	case <-h.closed:
	default:
		close(h.closed)
	}
}
