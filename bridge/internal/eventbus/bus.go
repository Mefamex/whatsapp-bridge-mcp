// whatsapp-bridge-project/bridge-go/internal/eventbus/bus.go

// Package eventbus fans out bridge events to live SSE subscribers and to any
// registered relay (e.g. the outbound HTTP webhook).
package eventbus

import (
	"encoding/json"
	"sync"
	"time"
)

// Relay receives every published event; used for webhook forwarding.
type Relay func(eventType string, data any)

// Bus is a concurrency-safe publish/subscribe hub.
type Bus struct {
	mu     sync.Mutex
	subs   map[chan []byte]struct{}
	relays []Relay
}

// New creates an empty bus.
func New() *Bus {
	return &Bus{subs: make(map[chan []byte]struct{})}
}

// AddRelay registers an extra sink that gets the structured event payload.
func (b *Bus) AddRelay(r Relay) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.relays = append(b.relays, r)
}

// Subscribe registers a new SSE listener and returns it with its cancel func.
func (b *Bus) Subscribe(buffer int) (chan []byte, func()) {
	if buffer <= 0 {
		buffer = 64
	}
	ch := make(chan []byte, buffer)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

// Publish sends an event to every subscriber and relay. Slow subscribers are
// skipped rather than blocking the caller.
func (b *Bus) Publish(eventType string, data any) {
	frame, err := b.frame(eventType, data)
	if err != nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subs {
		select {
		case ch <- frame:
		default:
		}
	}

	for _, r := range b.relays {
		r(eventType, data)
	}
}

// frame renders one SSE payload.
func (b *Bus) frame(eventType string, data any) ([]byte, error) {
	payload, err := json.Marshal(map[string]any{
		"event":     eventType,
		"timestamp": time.Now().Unix(),
		"data":      data,
	})
	if err != nil {
		return nil, err
	}
	return []byte("data: " + string(payload) + "\n\n"), nil
}
