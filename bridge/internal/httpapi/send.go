// whatsapp-bridge-project/bridge-go/internal/httpapi/send.go

package httpapi

import (
	"context"
	"encoding/base64"
	"net/http"

	"whatsapp-bridge/internal/wabridge"
)

// textRequest is the /send payload.
type textRequest struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

// mediaRequest is the /send-media payload.
type mediaRequest struct {
	To       string `json:"to"`
	MediaB64 string `json:"media_base64"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
	FileName string `json:"file_name,omitempty"`
}

// sendText delivers a plain text message.
func (s *Server) sendText(w http.ResponseWriter, r *http.Request) {
	var req textRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.To == "" || req.Text == "" {
		http.Error(w, "to and text required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	jid, err := s.wa.SendText(context.Background(), req.To, req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "to": jid})
}

// sendMedia uploads and delivers an attachment.
func (s *Server) sendMedia(w http.ResponseWriter, r *http.Request) {
	var req mediaRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.To == "" || req.MediaB64 == "" || req.MimeType == "" {
		http.Error(w, "to, media_base64, mime_type required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.MediaB64)
	if err != nil {
		http.Error(w, "Invalid base64: "+err.Error(), http.StatusBadRequest)
		return
	}

	jid, directPath, err := s.wa.SendMedia(context.Background(), wabridge.MediaRequest{
		To:       req.To,
		Data:     data,
		MimeType: req.MimeType,
		Caption:  req.Caption,
		FileName: req.FileName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"ok":          true,
		"to":          jid,
		"direct_path": directPath,
	})
}

// pollRequest is the /send-poll payload.
type pollRequest struct {
	To              string   `json:"to"`
	Name            string   `json:"name"`
	Options         []string `json:"options"`
	SelectableCount int      `json:"selectable_count,omitempty"`
}

// callRequest is the /call payload.
type callRequest struct {
	To          string `json:"to"`
	DurationSec int    `json:"duration_sec,omitempty"` // 0 = ring until answered/hangup
	IsVideo     bool   `json:"is_video,omitempty"`
}

// sendPoll delivers a poll (survey) message.
func (s *Server) sendPoll(w http.ResponseWriter, r *http.Request) {
	var req pollRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if _, _, err := wabridge.HazirlaAnket(wabridge.PollRequest{
		To:              req.To,
		Name:            req.Name,
		Options:         req.Options,
		SelectableCount: req.SelectableCount,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	jid, err := s.wa.SendPoll(context.Background(), wabridge.PollRequest{
		To:              req.To,
		Name:            req.Name,
		Options:         req.Options,
		SelectableCount: req.SelectableCount,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "to": jid})
}

// sendCall initiates an outbound WhatsApp call (ring the user).
func (s *Server) sendCall(w http.ResponseWriter, r *http.Request) {
	var req callRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.To == "" {
		http.Error(w, "to required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	duration := req.DurationSec
	if duration <= 0 {
		duration = 30 // default 30 seconds
	}

	if err := s.wa.SendCall(r.Context(), req.To, duration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "to": req.To, "duration_sec": duration})
}
