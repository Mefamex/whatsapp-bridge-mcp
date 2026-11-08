// whatsapp-bridge-project/bridge-go/internal/wabridge/dispatch.go

package wabridge

import (
	"context"
	"log"
	"os"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// Dispatch handles one raw whatsmeow event. It is registered as the client's
// event handler and is the single place where events enter the bridge.
func (s *Service) Dispatch(evt any) {
	switch v := evt.(type) {
	case *events.Connected:
		log.Println("✅ [BRIDGE] Connected to WhatsApp")
		s.bus.Publish("connected", map[string]any{})
		s.fireConnected()

	case *events.Disconnected:
		log.Println("⚠️ [BRIDGE] Disconnected")
		s.bus.Publish("disconnected", map[string]any{})

	case *events.LoggedOut:
		log.Printf("❌ [BRIDGE] Logged out: OnConnect=%v, Reason=%v", v.OnConnect, v.Reason)
		s.bus.Publish("logged_out", map[string]any{
			"on_connect": v.OnConnect,
			"reason":     v.Reason,
		})
		os.Exit(1)

	case *events.Message:
		s.handleMessage(v)

	case *events.Receipt:
		s.bus.Publish("receipt", map[string]any{
			"type":        string(v.Type),
			"message_ids": v.MessageIDs,
			"chat":        v.Chat.String(),
		})

	case *events.Presence:
		s.bus.Publish("presence", map[string]any{
			"from":        v.From.String(),
			"unavailable": v.Unavailable,
			"last_seen":   v.LastSeen.Unix(),
		})

	case *events.HistorySync:
		s.handleHistorySync(v)

	case *events.AppStateSyncComplete:
		s.bus.Publish("app_state_sync_complete", map[string]any{})

	default:
		s.handleCall(evt)
	}
}

// handleCall extracts call metadata from raw whatsmeow events and publishes
// them to the event bus. Non-call events are silently ignored.
func (s *Service) handleCall(evt any) {
	from, kind, node, ok := callEventOf(evt)
	if !ok {
		return
	}

	log.Printf("📞 [CALL] %s from %s", kind, from)

	// Extract call ID from the stanza if available
	callID := ""
	if node != nil {
		if id, ok := node.Attrs["call-id"].(string); ok {
			callID = id
		}
	}

	s.bus.Publish("call", map[string]any{
		"from":    from,
		"kind":    kind,
		"call_id": callID,
	})
}

// callEventOf extracts the call stanza and metadata from a raw event.
func callEventOf(evt any) (from, kind string, node *waBinary.Node, ok bool) {
	switch v := evt.(type) {
	case *events.CallOffer:
		return v.From.String(), "offer", v.Data, true
	case *events.CallPreAccept:
		return v.From.String(), "preaccept", v.Data, true
	case *events.CallAccept:
		return v.From.String(), "accept", v.Data, true
	case *events.CallTransport:
		return v.From.String(), "transport", v.Data, true
	case *events.CallRelayLatency:
		return v.From.String(), "relaylatency", v.Data, true
	case *events.CallOfferNotice:
		return v.From.String(), "offer_notice", v.Data, true
	case *events.CallReject:
		return v.From.String(), "reject", v.Data, true
	case *events.CallTerminate:
		return v.From.String(), "terminate", v.Data, true
	case *events.UnknownCallEvent:
		return unknownCallFrom(v.Node), "unknown", v.Node, v.Node != nil
	default:
		return "", "", nil, false
	}
}

// unknownCallFrom digs the sender out of an unrecognised call stanza.
func unknownCallFrom(node *waBinary.Node) string {
	if node == nil || node.Attrs == nil {
		return ""
	}
	if from, ok := node.Attrs["from"].(string); ok {
		return from
	}
	return ""
}

// lidTelefonu, LID kimliğin store'da bilinen telefon karşılığını döner.
//
// Eşleşme yoksa (veya istemci/store hazır değilse) "" döner — gelen mesajın
// From alanı asla değiştirilmez, sonuç yalnız TakmaAd'a eklenir. Olay
// döngüsünü kilitlememek için sorgu 2 sn ile sınırlıdır (CachedLIDMap
// zaten bellekten karşılar).
func (s *Service) lidTelefonu(jidStr string) string {
	if s == nil || s.client == nil || s.client.Store == nil {
		return ""
	}
	lid, err := types.ParseJID(jidStr)
	if err != nil || lid.Server != types.HiddenUserServer || lid.User == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pn, err := s.client.Store.LIDs.GetPNForLID(ctx, lid)
	if err != nil || pn.Server != types.DefaultUserServer || pn.User == "" {
		return ""
	}
	return pn.String()
}

// takmaAdVar, takma ad listesinde tekrarı önler.
func takmaAdVar(liste []string, aday string) bool {
	for _, v := range liste {
		if v == aday {
			return true
		}
	}
	return false
}

// handleMessage stores a live message and announces it to listeners.
//
// Edits arrive inside the normal Message event, so they are intercepted first:
// handleEditMessage decrypts, updates the store, re-publishes, and reports
// whether the event was an edit at all.
func (s *Service) handleMessage(v *events.Message) {
	if s.handleEditMessage(v) {
		return
	}

	if pollMsg, ok := s.handlePollUpdate(v); ok {
		for _, aday := range []string{pollMsg.From, pollMsg.Sender} {
			if pn := s.lidTelefonu(aday); pn != "" && pn != pollMsg.From && !takmaAdVar(pollMsg.TakmaAd, pn) {
				pollMsg.TakmaAd = append(pollMsg.TakmaAd, pn)
			}
		}
		log.Printf("📊 [%s] POLL VOTE %s -> %s: %s",
			time.Unix(pollMsg.Timestamp, 0).Format("15:04:05"),
			pollMsg.Sender, pollMsg.From, pollMsg.Text,
		)
		s.messages.Add(pollMsg)
		s.bus.Publish("message", pollMsg)
		return
	}

	msg := Translate(v)

	// LID takma adı: gelen mesaj LID ile gelir (`123@lid`); store'da
	// bilinen telefon karşılığı varsa TakmaAd'a eklenir. From DEĞİŞMEZ —
	// store anahtarları, dedup ve tetikleyici eşleşmesi etkilenmez; dış
	// istemci (tetikleyici) telefonu ek kimlik olarak görür.
	for _, aday := range []string{msg.From, msg.Sender} {
		if pn := s.lidTelefonu(aday); pn != "" && pn != msg.From && !takmaAdVar(msg.TakmaAd, pn) {
			msg.TakmaAd = append(msg.TakmaAd, pn)
		}
	}

	log.Printf("📨 [%s] %s %s -> %s: %s",
		time.Unix(msg.Timestamp, 0).Format("15:04:05"),
		msg.Direction, msg.Sender, msg.From, msg.Text,
	)

	// Control traffic stays out of the store. WhatsApp emits a stream of
	// protocol frames (history-sync acks, receipts, ephemeral settings) —
	// a couple per minute — and the store keeps only the last 1000 entries.
	// Storing them evicts real conversations, which is how a chat silently
	// disappears from /messages. The log line above still records them.
	if msg.Type == "protocol" {
		return
	}

	s.messages.Add(msg)
	s.bus.Publish("message", msg)
}
