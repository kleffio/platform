// Package events provides a lightweight in-process event bus for publishing
// domain events between modules. For cross-process events, replace this with
// a message queue adapter (e.g. Redis Streams, NATS, or Kafka).
package events

import (
	"context"
	"sync"
)

// Event is the base interface for all domain events.
type Event interface {
	// EventName returns a dot-separated identifier, e.g. "deployment.created".
	EventName() string
}

// Handler is a function that processes an event.
type Handler func(ctx context.Context, event Event) error

// Bus dispatches events to registered handlers.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New creates a new event bus.
func New() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// Subscribe registers a handler for a specific event name.
func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish dispatches an event synchronously to all registered handlers.
// Returns the first error encountered, if any.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
