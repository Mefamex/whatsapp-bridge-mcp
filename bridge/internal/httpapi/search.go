// whatsapp-bridge-project/bridge-go/internal/httpapi/search.go

package httpapi

import (
	"net/http"
	"strconv"
)

// maxSearchResults caps the ?limit= parameter.
const maxSearchResults = 50

// search scans the buffered message log for a case-insensitive text query and
// returns matches newest first. Store-backed only — no WhatsApp round trip.
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	if s.messages == nil {
		http.Error(w, `{"error":"mesaj deposu yok"}`, http.StatusInternalServerError)
		return
	}

	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > maxSearchResults {
		limit = maxSearchResults
	}

	results := s.messages.Search(q, limit)
	writeJSON(w, map[string]any{
		"query":   q,
		"count":   len(results),
		"results": results,
	})
}
