// whatsapp-bridge-project/bridge-go/internal/httpapi/server.go

// Package httpapi exposes the bridge's local REST/SSE interface.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/msgstore"
	"whatsapp-bridge/internal/wabridge"
)

// WhatsAppService is the slice of the bridge service the HTTP layer needs.
// *wabridge.Service satisfies it; tests substitute a fake.
type WhatsAppService interface {
	IsLoggedIn() bool
	Connected() bool
	SendText(ctx context.Context, to, text string) (string, error)
	SendMedia(ctx context.Context, req wabridge.MediaRequest) (string, string, error)
	SendPoll(ctx context.Context, req wabridge.PollRequest) (string, error)
	SendReaction(ctx context.Context, to, targetID, emoji, senderJID string) error
	RevokeMessage(ctx context.Context, to, targetID string) error
	SendCall(ctx context.Context, to string, durationSec int) error
	Contacts(ctx context.Context) ([]wabridge.Contact, error)
	Groups(ctx context.Context) ([]wabridge.Group, error)
	MarkRead(ctx context.Context, jid string, messageIDs []string) error
	MuteChat(ctx context.Context, jid string, muted bool, duration time.Duration) error
	StarMessage(ctx context.Context, jid, senderJID, messageID string, fromMe, starred bool) error
	ChatSettings(ctx context.Context, jid string) (wabridge.ChatState, error)
	RequestHistory(jid string, count int) error
	DownloadMedia(ctx context.Context, m *msgstore.Media) ([]byte, error)
}

// Server binds the WhatsApp service to HTTP handlers.
type Server struct {
	wa       WhatsAppService
	messages *msgstore.Store
	bus      *eventbus.Bus
	started  time.Time
}

// New creates a Server. started feeds the uptime field of /health.
func New(wa WhatsAppService, messages *msgstore.Store, bus *eventbus.Bus, started time.Time) *Server {
	return &Server{wa: wa, messages: messages, bus: bus, started: started}
}

// Router returns the fully wired HTTP router.
func (s *Server) Router() *mux.Router {
	r := mux.NewRouter()
	s.registerCoreRoutes(r)
	s.registerDirectoryRoutes(r)
	s.registerHistoryRoutes(r)
	s.registerEventRoutes(r)
	return r
}

// registerCoreRoutes wires health, sending and webhook endpoints.
func (s *Server) registerCoreRoutes(r *mux.Router) {
	r.HandleFunc("/health", s.health).Methods(http.MethodGet)
	r.HandleFunc("/send", s.sendText).Methods(http.MethodPost)
	r.HandleFunc("/send-media", s.sendMedia).Methods(http.MethodPost)
	r.HandleFunc("/send-poll", s.sendPoll).Methods(http.MethodPost)
	r.HandleFunc("/reaction", s.sendReaction).Methods(http.MethodPost)
	r.HandleFunc("/revoke", s.revokeMessage).Methods(http.MethodPost)
	r.HandleFunc("/call", s.sendCall).Methods(http.MethodPost)
	r.HandleFunc("/webhook", s.webhook).Methods(http.MethodPost, http.MethodDelete, http.MethodGet)
}

// registerDirectoryRoutes wires contacts, groups and read receipts.
func (s *Server) registerDirectoryRoutes(r *mux.Router) {
	r.HandleFunc("/contacts", s.contacts).Methods(http.MethodGet)
	r.HandleFunc("/groups", s.groups).Methods(http.MethodGet)
	r.HandleFunc("/mark-read", s.markRead).Methods(http.MethodPost)
	r.HandleFunc("/mute", s.mute).Methods(http.MethodPost)
	r.HandleFunc("/star", s.star).Methods(http.MethodPost)
	r.HandleFunc("/chat-state", s.chatState).Methods(http.MethodGet)
}

// registerHistoryRoutes wires message history and media download.
func (s *Server) registerHistoryRoutes(r *mux.Router) {
	r.HandleFunc("/messages/{jid}", s.messagesHandler).Methods(http.MethodGet)
	r.HandleFunc("/media/{messageId}", s.mediaDownload).Methods(http.MethodGet)
	r.HandleFunc("/search", s.search).Methods(http.MethodGet)
}

// registerEventRoutes wires the live event stream.
func (s *Server) registerEventRoutes(r *mux.Router) {
	r.HandleFunc("/events", s.events).Methods(http.MethodGet)
}

// requireConnection guards endpoints that need a live socket.
func (s *Server) requireConnection(w http.ResponseWriter) bool {
	if s.wa.Connected() {
		return true
	}
	http.Error(w, "Not connected", http.StatusServiceUnavailable)
	return false
}

// writeJSON renders v as a JSON response.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON parses a JSON request body, reporting a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}
