// whatsapp-bridge-project/bridge-go/internal/wabridge/translate.go

package wabridge

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/msgstore"
)

// Translate converts a whatsmeow message event into the canonical store shape.
//
// A missing payload is not an error: history syncs carry placeholder entries
// whose WebMessageInfo never decoded into a real waE2E.Message. Classifying a
// nil payload used to panic here, and because whatsmeow recovers event-handler
// panics, the entire history batch was lost instead of just the broken entry.
func Translate(v *events.Message) msgstore.Message {
	if v == nil || v.Message == nil {
		return missingPayloadMessage(v)
	}

	kind, text, media := classify(v.Message)
	quoted := extractQuoted(v.Message)

	direction := "in"
	if v.Info.IsFromMe {
		direction = "out"
	}

	return msgstore.Message{
		ID:        string(v.Info.ID),
		Type:      kind,
		Direction: direction,
		From:      v.Info.Chat.String(),
		Sender:    v.Info.Sender.String(),
		Timestamp: v.Info.Timestamp.Unix(),
		Text:      text,
		IsFromMe:  v.Info.IsFromMe,
		IsGroup:   v.Info.IsGroup,
		PushName:  v.Info.PushName,
		Media:     media,
		Quoted:    quoted,
	}
}

// missingPayloadMessage describes an event that carried no payload, keeping
// every field the event still has so the gap stays traceable in the log.
func missingPayloadMessage(v *events.Message) msgstore.Message {
	kind, text := describeUnsupported(nil)

	if v == nil {
		return msgstore.Message{Type: kind, Direction: "unknown", Text: text}
	}

	direction := "in"
	if v.Info.IsFromMe {
		direction = "out"
	}

	return msgstore.Message{
		ID:        string(v.Info.ID),
		Type:      kind,
		Direction: direction,
		From:      v.Info.Chat.String(),
		Sender:    v.Info.Sender.String(),
		Timestamp: v.Info.Timestamp.Unix(),
		Text:      text,
		IsFromMe:  v.Info.IsFromMe,
		IsGroup:   v.Info.IsGroup,
		PushName:  v.Info.PushName,
	}
}

// classify extracts the message kind, human-readable text and media metadata.
func classify(m *waE2E.Message) (kind, text string, media *msgstore.Media) {
	if m == nil {
		return "noPayload", "[unsupported: nilPayload]", nil
	}
	// Check for empty message (no fields set)
	v := reflect.ValueOf(m).Elem()
	if v.IsValid() && v.NumField() > 0 {
		hasContent := false
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsZero() {
				hasContent = true
				break
			}
		}
		if !hasContent {
			return "noPayload", "[unsupported: noPayload]", nil
		}
	}

	switch {
	case m.Conversation != nil:
		return "text", m.GetConversation(), nil

	case m.ExtendedTextMessage != nil:
		return "extended_text", m.GetExtendedTextMessage().GetText(), nil

	case m.GetImageMessage() != nil:
		img := m.GetImageMessage()
		return "image", appcfg.DerefString(img.Caption), newMedia(img.URL, img.DirectPath,
			img.MediaKey, img.FileEncSHA256, img.FileSHA256, img.FileLength, img.Mimetype)

	case m.GetVideoMessage() != nil:
		vid := m.GetVideoMessage()
		return "video", appcfg.DerefString(vid.Caption), newMedia(vid.URL, vid.DirectPath,
			vid.MediaKey, vid.FileEncSHA256, vid.FileSHA256, vid.FileLength, vid.Mimetype)

	case m.GetAudioMessage() != nil:
		aud := m.GetAudioMessage()
		return "audio", "", newMedia(aud.URL, aud.DirectPath,
			aud.MediaKey, aud.FileEncSHA256, aud.FileSHA256, aud.FileLength, aud.Mimetype)

	case m.GetDocumentMessage() != nil:
		doc := m.GetDocumentMessage()
		return "document", appcfg.DerefString(doc.Title), newMedia(doc.URL, doc.DirectPath,
			doc.MediaKey, doc.FileEncSHA256, doc.FileSHA256, doc.FileLength, doc.Mimetype)

	case m.GetStickerMessage() != nil:
		stk := m.GetStickerMessage()
		return "sticker", "", newMedia(stk.URL, stk.DirectPath,
			stk.MediaKey, stk.FileEncSHA256, stk.FileSHA256, stk.FileLength, stk.Mimetype)

	// Anket: telefon V3 gönderiyor (canlı kanıt 2026-10-06 02:47),
	// V1/V2/V5/V6 aynı struct. V4 FutureProofMessage'dır, çözülemez.
	case anketBul(m) != nil:
		anket := anketBul(m)
		secenekler := make([]string, 0, len(anket.GetOptions()))
		for _, secenek := range anket.GetOptions() {
			secenekler = append(secenekler, secenek.GetOptionName())
		}
		return "poll", renderPoll(anket.GetName(), secenekler), nil

	// Anket oy kullanımı (kullanıcı anket seçeneğine tıkladığında)
	case m.GetPollUpdateMessage() != nil:
		return "poll_update", "Anket oyu kullanıldı", nil

	// Reaksiyon (emoji tepkisi)
	case m.GetReactionMessage() != nil:
		reaction := m.GetReactionMessage()
		return "reaction", reaction.GetText(), nil

	// Konum (Location)
	case m.GetLocationMessage() != nil:
		loc := m.GetLocationMessage()
		metin := fmt.Sprintf("📍 Konum: %.6f, %.6f", loc.GetDegreesLatitude(), loc.GetDegreesLongitude())
		if name := strings.TrimSpace(loc.GetName()); name != "" {
			metin += fmt.Sprintf(" (%s)", name)
		}
		if addr := strings.TrimSpace(loc.GetAddress()); addr != "" {
			metin += fmt.Sprintf(" - %s", addr)
		}
		metin += fmt.Sprintf("\nHarita: https://maps.google.com/?q=%.6f,%.6f", loc.GetDegreesLatitude(), loc.GetDegreesLongitude())
		return "location", metin, nil

	// Canlı Konum (Live Location)
	case m.GetLiveLocationMessage() != nil:
		loc := m.GetLiveLocationMessage()
		metin := fmt.Sprintf("📍 Canlı Konum: %.6f, %.6f\nHarita: https://maps.google.com/?q=%.6f,%.6f",
			loc.GetDegreesLatitude(), loc.GetDegreesLongitude(), loc.GetDegreesLatitude(), loc.GetDegreesLongitude())
		if cap := strings.TrimSpace(loc.GetCaption()); cap != "" {
			metin = cap + "\n" + metin
		}
		return "live_location", metin, nil

	// Kişi Kartı (vCard)
	case m.GetContactMessage() != nil:
		cont := m.GetContactMessage()
		metin := fmt.Sprintf("📇 Kişi Kartı: %s", cont.GetDisplayName())
		if vcard := strings.TrimSpace(cont.GetVcard()); vcard != "" {
			metin += "\n" + vcard
		}
		return "contact", metin, nil

	// Kişi Listesi (Contacts Array)
	case m.GetContactsArrayMessage() != nil:
		arr := m.GetContactsArrayMessage()
		metin := fmt.Sprintf("📇 Kişi Listesi: %s (%d kişi)", arr.GetDisplayName(), len(arr.GetContacts()))
		return "contact_array", metin, nil

	// Control traffic, not content: revokes, disappearing-message settings,
	// history-sync acknowledgements. Worth naming so the log can tell a
	// sync notification from an actual revoke.
	case m.GetProtocolMessage() != nil:
		return "protocol", "protocol:" + protocolMessageKind(m.GetProtocolMessage()), nil

	default:
		kind, text = describeUnsupported(m)
		return kind, text, nil
	}
}

// protocolMessageKind names the control message so the log shows what
// WhatsApp was actually telling us. These carry no readable content.
func protocolMessageKind(pm *waE2E.ProtocolMessage) string {
	if pm == nil {
		return "nil"
	}
	// REVOKE is the zero value, so an unset type reads as REVOKE.
	return strings.TrimPrefix(pm.GetType().String(), "ProtocolMessage_")
}

// anketBul gelen anket sürümlerini tek yapıya indirir.
// Sıra önemli değil: bir mesajda tek anket alanı dolu olur.
func anketBul(m *waE2E.Message) *waE2E.PollCreationMessage {
	for _, al := range []*waE2E.PollCreationMessage{
		m.GetPollCreationMessage(),
		m.GetPollCreationMessageV2(),
		m.GetPollCreationMessageV3(),
		m.GetPollCreationMessageV5(),
		m.GetPollCreationMessageV6(),
	} {
		if al != nil {
			return al
		}
	}
	return nil
}

// newMedia builds the media metadata block for an attachment.
func newMedia(url, directPath *string, mediaKey, encHash, fileHash []byte, length *uint64, mime *string) *msgstore.Media {
	return &msgstore.Media{
		URL:           appcfg.DerefString(url),
		DirectPath:    appcfg.DerefString(directPath),
		MediaKey:      base64.StdEncoding.EncodeToString(mediaKey),
		FileEncSHA256: base64.StdEncoding.EncodeToString(encHash),
		FileSHA256:    base64.StdEncoding.EncodeToString(fileHash),
		FileLength:    appcfg.DerefUint64(length),
		MimeType:      appcfg.DerefString(mime),
	}
}

// getContextInfo returns the ContextInfo from whichever message type carries it.
func getContextInfo(m *waE2E.Message) *waE2E.ContextInfo {
	if m == nil {
		return nil
	}
	if ext := m.GetExtendedTextMessage(); ext != nil && ext.ContextInfo != nil {
		return ext.ContextInfo
	}
	if img := m.GetImageMessage(); img != nil && img.ContextInfo != nil {
		return img.ContextInfo
	}
	if vid := m.GetVideoMessage(); vid != nil && vid.ContextInfo != nil {
		return vid.ContextInfo
	}
	if aud := m.GetAudioMessage(); aud != nil && aud.ContextInfo != nil {
		return aud.ContextInfo
	}
	if doc := m.GetDocumentMessage(); doc != nil && doc.ContextInfo != nil {
		return doc.ContextInfo
	}
	if stk := m.GetStickerMessage(); stk != nil && stk.ContextInfo != nil {
		return stk.ContextInfo
	}
	return nil
}

// extractQuoted pulls the quoted message context (native WhatsApp swipe reply).
func extractQuoted(m *waE2E.Message) *msgstore.QuotedMessage {
	ci := getContextInfo(m)
	if ci == nil {
		return nil
	}

	var stanzaID, participant string
	if ci.StanzaID != nil {
		stanzaID = *ci.StanzaID
	}
	if ci.Participant != nil {
		participant = *ci.Participant
	}

	if ci.QuotedMessage == nil && stanzaID == "" {
		return nil
	}

	var qKind, qText string
	if ci.QuotedMessage != nil {
		qKind, qText, _ = classify(ci.QuotedMessage)
	}

	return &msgstore.QuotedMessage{
		ID:     stanzaID,
		Sender: participant,
		Type:   qKind,
		Text:   qText,
	}
}
