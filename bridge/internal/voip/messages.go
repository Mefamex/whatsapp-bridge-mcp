// whatsapp-bridge-project/bridge-go/internal/voip/messages.go

package voip

import (
	"fmt"
	"log"
	"strings"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// missingAssetError describes sidecar prereqs that are not on disk.
type missingAssetError struct {
	what string
	path string
}

func (e *missingAssetError) Error() string {
	return fmt.Sprintf("%s not found at %s", e.what, e.path)
}

// jidTriple derives the phone and LID variants of the account's own JID.
func jidTriple(selfJID string) (self, phone, lid string) {
	if strings.Contains(selfJID, "@lid") {
		return selfJID, strings.Replace(selfJID, "@lid", "@s.whatsapp.net", 1), selfJID
	}
	return selfJID, selfJID, strings.Replace(selfJID, "@s.whatsapp.net", "@lid", 1)
}

// handleMessage processes one IPC message coming from the sidecar.
func (s *Sidecar) handleMessage(msg map[string]any) {
	msgType, _ := msg["type"].(string)

	switch msgType {
	case "ready":
		s.mu.Lock()
		s.ready = true
		s.mu.Unlock()
		log.Println("✅ [Sidecar] Ready for calls")

	case "call_stanza_out":
		if stanza, ok := msg["stanza"].(map[string]any); ok {
			s.EmitStanza(stanza)
		}

	case "call_state":
		callID, _ := msg["callId"].(string)
		state, _ := msg["state"].(string)
		from, _ := msg["from"].(string)
		log.Printf("📞 [Sidecar] Call %s: %s (from: %s)", callID, state, from)
		s.bus.Publish("call_state", msg)

	case "status":
		text, _ := msg["message"].(string)
		log.Printf("ℹ️ [Sidecar] %s", text)

	case "error":
		text, _ := msg["message"].(string)
		log.Printf("❌ [Sidecar] Error: %s", text)

	default:
		log.Printf("❓ [Sidecar] Unknown message type: %q", msgType)
	}
}

// EmitStanza converts a sidecar stanza map and sends it to WhatsApp.
func (s *Sidecar) EmitStanza(stanza map[string]any) {
	node := MapToNode(stanza)
	if node == nil {
		return
	}
	if s.sender == nil {
		log.Println("⚠️ [Sidecar] No stanza sender configured")
		return
	}
	if err := s.sender.SendNode(node); err != nil {
		log.Printf("❌ [Sidecar] Failed to send call stanza: %v", err)
	}
}

// RelayCall forwards an inbound call stanza from WhatsApp to the sidecar.
func (s *Sidecar) RelayCall(node *waBinary.Node, from, eventType string) {
	if node == nil {
		return
	}
	stanza := NodeToMap(node)
	stanza["from"] = from
	stanza["event_type"] = eventType

	s.Send(map[string]any{"type": "call_stanza", "stanza": stanza})
}
