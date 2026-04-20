package application

import (
	"sync"

	"github.com/kleffio/platform/internal/core/notifications/domain"
)

// Hub manages open SSE connections keyed by user ID.
// Each connected browser tab gets its own channel; when a notification is
// created the Hub fans it out to all open tabs for that user.
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]chan *domain.Notification
}

// NewHub returns an initialised Hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string][]chan *domain.Notification)}
}

// Subscribe registers a new channel for userID and returns it.
// The caller must call Unsubscribe when the SSE connection closes.
func (h *Hub) Subscribe(userID string) chan *domain.Notification {
	ch := make(chan *domain.Notification, 16)
	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], ch)
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from the hub and closes it.
func (h *Hub) Unsubscribe(userID string, ch chan *domain.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels := h.clients[userID]
	for i, c := range channels {
		if c == ch {
			h.clients[userID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			return
		}
	}
}

// Push delivers a notification to all open SSE connections for the user.
// Non-blocking: a slow consumer misses the event (it will catch up via polling).
func (h *Hub) Push(userID string, n *domain.Notification) {
	h.mu.RLock()
	channels := make([]chan *domain.Notification, len(h.clients[userID]))
	copy(channels, h.clients[userID])
	h.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- n:
		default:
		}
	}
}
