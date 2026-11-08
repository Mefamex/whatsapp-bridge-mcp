// whatsapp-bridge-project/bridge-go/internal/httpapi/webhook.go

package httpapi

import (
	"net/http"

	"whatsapp-bridge/internal/webhook"
)

// webhookRequest is the /webhook registration payload.
type webhookRequest struct {
	URL string `json:"url"`
}

// webhook registers, inspects or clears the outbound event webhook.
func (s *Server) webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req webhookRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.URL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}
		webhook.Register(req.URL)
		writeJSON(w, map[string]any{"ok": true, "webhook": req.URL})

	case http.MethodDelete:
		webhook.Register("")
		writeJSON(w, map[string]any{"webhook": ""})

	default:
		writeJSON(w, map[string]any{"webhook": webhook.Current()})
	}
}
