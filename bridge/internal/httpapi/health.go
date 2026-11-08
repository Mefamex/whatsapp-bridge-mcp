// whatsapp-bridge-project/bridge-go/internal/httpapi/health.go

package httpapi

import (
	"net/http"
	"time"
)

// healthResponse is the /health payload. It separates a stored session from a
// live connection so a caller can tell "not paired yet" from "socket down".
type healthResponse struct {
	Status    string `json:"status"`
	Paired    bool   `json:"paired"`
	Connected bool   `json:"connected"`
	Uptime    string `json:"uptime"`
	Buffered  int    `json:"buffered_messages"`
}

// health reports liveness, session state and uptime.
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, healthResponse{
		Status:    "ok",
		Paired:    s.wa.IsLoggedIn(),
		Connected: s.wa.Connected(),
		Uptime:    time.Since(s.started).String(),
		Buffered:  s.messages.Len(),
	})
}
