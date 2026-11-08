// whatsapp-bridge-project/bridge-go/internal/webhook/webhook.go

// Package webhook stores a single outbound webhook URL and relays bridge
// events to it as JSON POST requests.
package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Sink is the minimal relay behaviour the bus expects.
type Sink func(eventType string, data any)

var (
	mu  sync.RWMutex
	url string
)

// Register sets (or clears, when empty) the target webhook URL.
func Register(target string) {
	mu.Lock()
	defer mu.Unlock()
	url = target
}

// Current returns the registered webhook URL, if any.
func Current() string {
	mu.RLock()
	defer mu.RUnlock()
	return url
}

// Relay implements eventbus.Relay: POST every event to the registered URL.
// It is a no-op while no URL is configured.
func Relay(eventType string, data any) {
	target := Current()
	if target == "" {
		return
	}

	body, err := json.Marshal(map[string]any{
		"event": eventType,
		"data":  data,
	})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(target, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
