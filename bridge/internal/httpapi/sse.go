// whatsapp-bridge-project/bridge-go/internal/httpapi/sse.go

package httpapi

import (
	"fmt"
	"net/http"
	"time"
)

// heartbeatInterval keeps idle SSE connections alive through proxies.
//
// It must be comfortably shorter than any proxy or load-balancer idle
// timeout on the path. While the server's own write deadline was 30s this
// ticker fired at exactly 30s too, so the connection always lost that race.
const heartbeatInterval = 15 * time.Second

// subscribeBuffer is the per-client SSE queue depth.
const subscribeBuffer = 100

// events streams live bridge events as Server-Sent Events.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	if !s.requireConnection(w) {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("Access-Control-Allow-Origin", "*")

	updates, cancel := s.bus.Subscribe(subscribeBuffer)
	defer cancel()

	fmt.Fprintf(w, "data: {\"event\":\"connected\",\"timestamp\":%d}\n\n", time.Now().Unix())
	flusher.Flush()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case frame := <-updates:
			if _, err := w.Write(frame); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
