// whatsapp-bridge-project/bridge-go/internal/wabridge/classify_test.go

package wabridge

import (
	"strings"
	"testing"

	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// A history sync can hand us an event whose payload never arrived. Whatsmeow
// recovers the panic, but the whole batch dies with it, so every message in
// that sync is lost — not just the broken entry.
func TestTranslateSurvivesMissingPayload(t *testing.T) {
	cases := []struct {
		name  string
		event *events.Message
	}{
		{"nil event", nil},
		{"nil payload", &events.Message{
			Info: types.MessageInfo{Chat: types.JID{User: "9051", Server: "s.whatsapp.net"}},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := Translate(tc.event)

			if msg.Type != "unknown" {
				t.Errorf("type = %q, want %q", msg.Type, "unknown")
			}
			if msg.ID != "" {
				t.Errorf("id = %q, want empty", msg.ID)
			}
		})
	}
}

// Without the payload there is no kind to report, but the identity that
// survived it must still reach the store so the gap stays traceable.
func TestTranslateKeepsIdentityWhenPayloadIsMissing(t *testing.T) {
	chat := types.JID{User: "12345678901234", Server: "lid"}
	msg := Translate(&events.Message{
		Info: types.MessageInfo{
			ID:       "WAID123",
			Chat:     chat,
			Sender:   chat,
			PushName: "Mefamex",
		},
	})

	if msg.Text != "[unsupported: nilPayload]" {
		t.Errorf("text = %q, want the nil-payload marker", msg.Text)
	}
	if msg.From != chat.String() {
		t.Errorf("from = %q, want %q", msg.From, chat.String())
	}
	if msg.PushName != "Mefamex" {
		t.Errorf("pushName = %q, want %q", msg.PushName, "Mefamex")
	}
}

// The default branch is the only place we learn what WhatsApp is sending us,
// so it must name the offending field instead of hiding it behind a constant.
func TestClassifyNamesTheUnknownProtoField(t *testing.T) {
	cases := []struct {
		name string
		msg  *waE2E.Message
		want string // expected kind
	}{
		{"reaction", &waE2E.Message{ReactionMessage: &waE2E.ReactionMessage{
			Key:               &waCommon.MessageKey{ID: proto.String("ABC"), FromMe: proto.Bool(true)},
			Text:              proto.String("👍"),
			SenderTimestampMS: proto.Int64(1700000000),
		}}, "reaction"},
		{"empty", &waE2E.Message{}, "noPayload"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, text, media := classify(tc.msg)

			if kind != tc.want {
				t.Errorf("kind = %q, want %q", kind, tc.want)
			}
			if media != nil {
				t.Error("media should be nil for unhandled types")
			}
			if tc.name == "reaction" && !strings.Contains(text, "👍") {
				t.Errorf("text = %q, want it to contain 👍", text)
			}
			if tc.name == "empty" && !strings.Contains(text, "noPayload") {
				t.Errorf("text = %q, want it to contain noPayload", text)
			}
		})
	}
}

// "extendedText" and "viewOnceMessageV2Extension" share a prefix; a bare list
// would be unreadable.
func TestDescribeUnsupportedListsDistinctFields(t *testing.T) {
	_, text := describeUnsupported(&waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{Text: proto.String("👍")},
		ProtocolMessage: &waE2E.ProtocolMessage{Key: &waCommon.MessageKey{ID: proto.String("DEF")}},
	})

	if !strings.Contains(text, "ReactionMessage") || !strings.Contains(text, "ProtocolMessage") {
		t.Errorf("text = %q, want both field names", text)
	}
}

// Protocol messages are control traffic (revokes, sync notifications), not
// content. They used to fall through to the same "[unsupported]" as a
// reaction or a poll, which made the two indistinguishable.
func TestClassifyNamesProtocolControlMessages(t *testing.T) {
	cases := []struct {
		name string
		msg  *waE2E.Message
		want string
	}{
		{"history sync notification", &waE2E.Message{ProtocolMessage: &waE2E.ProtocolMessage{
			Type: waE2E.ProtocolMessage_HISTORY_SYNC_NOTIFICATION.Enum(),
		}}, "protocol:HISTORY_SYNC_NOTIFICATION"},
		{"revoke", &waE2E.Message{ProtocolMessage: &waE2E.ProtocolMessage{
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
		}}, "protocol:REVOKE"},
		{"ephemeral setting", &waE2E.Message{ProtocolMessage: &waE2E.ProtocolMessage{
			Type: waE2E.ProtocolMessage_EPHEMERAL_SETTING.Enum(),
		}}, "protocol:EPHEMERAL_SETTING"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, text, media := classify(tc.msg)

			if kind != "protocol" {
				t.Errorf("kind = %q, want %q", kind, "protocol")
			}
			if !strings.Contains(text, tc.want) {
				t.Errorf("text = %q, want it to contain %q", text, tc.want)
			}
			if media != nil {
				t.Error("media should be nil for protocol messages")
			}
		})
	}
}

// Guards against the new fallback shadowing a branch that already worked.
func TestClassifyStillHandlesTheKnownTypes(t *testing.T) {
	cases := []struct {
		name string
		msg  *waE2E.Message
		want string
	}{
		{"conversation", &waE2E.Message{Conversation: proto.String("selam")}, "text"},
		{"extended", &waE2E.Message{ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("uzun")}}, "extended_text"},
		{"image", &waE2E.Message{ImageMessage: &waE2E.ImageMessage{Mimetype: proto.String("image/jpeg")}}, "image"},
		{"audio", &waE2E.Message{AudioMessage: &waE2E.AudioMessage{Mimetype: proto.String("audio/ogg")}}, "audio"},
		{"document", &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{Title: proto.String("f.pdf")}}, "document"},
		{"sticker", &waE2E.Message{StickerMessage: &waE2E.StickerMessage{Mimetype: proto.String("image/webp")}}, "sticker"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, _, _ := classify(tc.msg)
			if kind != tc.want {
				t.Errorf("kind = %q, want %q", kind, tc.want)
			}
		})
	}
}

func TestTranslateExtractsQuotedMessage(t *testing.T) {
	stanza := "ORIG123"
	sender := "905551234567@s.whatsapp.net"
	quotedText := "Önceki soru metni"

	extMsg := &waE2E.ExtendedTextMessage{
		Text: proto.String("Tamam, anladım"),
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:    proto.String(stanza),
			Participant: proto.String(sender),
			QuotedMessage: &waE2E.Message{
				Conversation: proto.String(quotedText),
			},
		},
	}

	evt := &events.Message{
		Info: types.MessageInfo{
			ID:   "MSG456",
			Chat: types.JID{User: "905551234567", Server: "s.whatsapp.net"},
		},
		Message: &waE2E.Message{
			ExtendedTextMessage: extMsg,
		},
	}

	msg := Translate(evt)
	if msg.Quoted == nil {
		t.Fatal("expected Quoted message, got nil")
	}
	if msg.Quoted.ID != stanza {
		t.Errorf("Quoted.ID = %q, want %q", msg.Quoted.ID, stanza)
	}
	if msg.Quoted.Sender != sender {
		t.Errorf("Quoted.Sender = %q, want %q", msg.Quoted.Sender, sender)
	}
	if msg.Quoted.Text != quotedText {
		t.Errorf("Quoted.Text = %q, want %q", msg.Quoted.Text, quotedText)
	}
	if msg.Quoted.Type != "text" {
		t.Errorf("Quoted.Type = %q, want text", msg.Quoted.Type)
	}
}
