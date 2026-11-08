// whatsapp-bridge-project/bridge-go/internal/wabridge/history.go

package wabridge

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"whatsapp-bridge/internal/appcfg"
)

// errNoAnchor means WhatsApp cannot walk history without a known message.
var errNoAnchor = errors.New("no previous message to anchor a history request")

// errNotPaired means the client has no paired identity yet.
var errNotPaired = errors.New("not paired")

// errMessageIDRequired means a star request came without a message id.
//
// Yıldız bir MESAJI işaretler, sohbeti değil. Kimliksiz istek sessizce
// geçersiz sayılabilirdi ve kullanıcı telefonunda "yıldızlamadı" görür
// ama nedenini bilmez.
var errMessageIDRequired = errors.New("message_id required")

// errJIDRequired means no chat identity was supplied at all.
var errJIDRequired = errors.New("jid required")

// errJIDMalformed means the chat identity could not be a real WhatsApp chat.
//
// Ayrım önemli: "required" eksiklik, "malformed" yanlış biçim. Kullanıcı
// ikisini de düzeltebilir ama hatayı farklı okur.
var errJIDMalformed = errors.New("jid malformed")

// RequestHistory asks WhatsApp for older messages of a chat. WhatsApp only
// serves history relative to a message the client already knows, so the newest
// buffered message of that chat is used as the anchor.
func (s *Service) RequestHistory(jid string, count int) error {
	target, err := types.ParseJID(appcfg.NormalizeJID(jid))
	if err != nil {
		return err
	}

	anchor, ok := s.messages.LastAnchor(target.String())
	if !ok {
		return errNoAnchor
	}

	req := s.client.BuildHistorySyncRequest(&types.MessageInfo{
		Chat:      target,
		ID:        types.MessageID(anchor.ID),
		Timestamp: time.Unix(anchor.Timestamp, 0),
		IsFromMe:  anchor.IsFromMe,
	}, count)
	if req == nil {
		return errors.New("failed to build history sync request")
	}

	if _, err := s.client.SendPeerMessage(context.Background(), req); err != nil {
		return err
	}

	log.Printf("📜 [HISTORY] History request sent for %s (count: %d)", target, count)
	return nil
}

// handleHistorySync buffers every message WhatsApp pushed as history and
// announces the result on the event bus.
func (s *Service) handleHistorySync(v *events.HistorySync) {
	if v.Data == nil {
		return
	}

	conversations := v.Data.GetConversations()
	if len(conversations) == 0 {
		return
	}
	log.Printf("📜 [HISTORY] History sync with %d conversations", len(conversations))

	stored, failed := 0, 0
	for _, conv := range conversations {
		if conv == nil {
			continue
		}
		chat, err := types.ParseJID(conv.GetID())
		if err != nil {
			log.Printf("⚠️ [HISTORY] Unusable conversation id %q: %v", conv.GetID(), err)
			failed++
			continue
		}
		for _, hmsg := range conv.GetMessages() {
			if hmsg == nil {
				continue
			}
			web := hmsg.GetMessage()
			if web == nil {
				continue
			}
			ev, err := s.client.ParseWebMessage(chat, web)
			if err != nil {
				failed++
				log.Printf("⚠️ [HISTORY] Cannot decode message %s in %s: %v", web.GetKey().GetID(), chat, err)
				continue
			}
			if ev == nil {
				failed++
				continue
			}
			// History can also carry edits; fold them the same way as live ones.
			if s.handleEditMessage(ev) {
				stored++
				continue
			}
			s.messages.Add(Translate(ev))
			stored++
		}
	}

	log.Printf("📜 [HISTORY] Buffered %d historical messages (%d failed)", stored, failed)
	s.bus.Publish("history_sync", map[string]any{"stored": stored, "failed": failed})
}
