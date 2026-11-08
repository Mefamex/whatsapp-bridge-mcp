// whatsapp-bridge-project/bridge-go/internal/voip/calls.go

package voip

import (
	"log"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types/events"
)

// HandleEvent translates a raw whatsmeow event into a call stanza and relays it
// to the sidecar. It reports whether the event was a call event at all.
func (s *Sidecar) HandleEvent(evt any) bool {
	from, kind, node, ok := callEventOf(evt)
	if !ok {
		return false
	}

	log.Printf("📞 [CALL] %s from %s", kind, from)
	s.RelayCall(node, from, kind)
	return true
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
