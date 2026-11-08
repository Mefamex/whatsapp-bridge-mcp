// whatsapp-bridge-project/bridge-go/internal/httpapi/directory.go

package httpapi

import (
	"context"
	"net/http"
)

// contacts lists every contact in the session store.
func (s *Server) contacts(w http.ResponseWriter, _ *http.Request) {
	if !s.requireConnection(w) {
		return
	}

	list, err := s.wa.Contacts(context.Background())
	if err != nil {
		http.Error(w, "Failed to get contacts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"count": len(list), "contacts": list})
}

// groups lists every joined group.
func (s *Server) groups(w http.ResponseWriter, _ *http.Request) {
	if !s.requireConnection(w) {
		return
	}

	list, err := s.wa.Groups(context.Background())
	if err != nil {
		http.Error(w, "Failed to get groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"count": len(list), "groups": list})
}
