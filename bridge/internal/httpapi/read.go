// whatsapp-bridge-project/bridge-go/internal/httpapi/read.go

package httpapi

import (
	"context"
	"net/http"
)

// markReadRequest is the /mark-read payload.
type markReadRequest struct {
	JID        string   `json:"jid"`
	MessageIDs []string `json:"message_ids"`
}

// markRead sends read receipts for a batch of messages.
func (s *Server) markRead(w http.ResponseWriter, r *http.Request) {
	var req markReadRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.JID == "" || len(req.MessageIDs) == 0 {
		http.Error(w, "jid and message_ids required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	if err := s.wa.MarkRead(context.Background(), req.JID, req.MessageIDs); err != nil {
		http.Error(w, "Failed to mark read: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
