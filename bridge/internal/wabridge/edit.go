package wabridge

import (
	"context"
	"log"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"

	"whatsapp-bridge/internal/msgstore"
)

// handleEditMessage resolves a SecretEncryptedMessage carrying a message edit
// and folds the new text into the store.
//
// WhatsApp delivers edits inside the normal Message event — there is no
// separate edit event. whatsmeow does not decrypt them automatically; the
// plaintext must be requested explicitly. Without this step the edit lands in
// the store as "[unsupported: MessageContextInfo, SecretEncryptedMessage]"
// with empty text, and the new content is lost everywhere (log, store, MCP).
//
// It returns true when the event was an edit and was handled (whether or not
// the original was found). False means "not an edit, handle normally".
func (s *Service) handleEditMessage(v *events.Message) bool {
	if v.Message.GetSecretEncryptedMessage() == nil {
		return false
	}

	decrypted, err := s.client.DecryptSecretEncryptedMessage(context.Background(), v)
	if err != nil {
		log.Printf("⚠️ [EDIT] çözümleme başarısız (%s): %v", v.Info.ID, err)
		// Fall through to the normal path so the raw edit is at least visible
		// as unknown, instead of vanishing silently.
		return false
	}

	targetID, text, kind, media, ok := unwrapEdit(decrypted)
	if !ok {
		log.Printf("⚠️ [EDIT] içerik çıkarılamadı (%s)", v.Info.ID)
		return true
	}

	editedAt := v.Info.Timestamp.Unix()
	if editedAt == 0 {
		editedAt = time.Now().Unix()
	}

	if s.messages.ApplyEdit(targetID, text, kind, media, editedAt) {
		log.Printf("✏️ [EDIT] %s güncellendi: %q", targetID, truncateEdit(text))

		// Re-publish so SSE listeners see the new text without re-reading.
		if updated, found := s.messages.ByID(targetID); found {
			s.bus.Publish("message", updated)
		}
		return true
	}

	// The original is not buffered (restart wiped the store, or the edit
	// arrived before the history sync). Store the new text as its own record
	// under the target ID so the content is at least visible.
	log.Printf("✏️ [EDIT] orijinal bulunamadı (%s), yeni metin ayrı kayıt", targetID)
	fallback := Translate(v)
	fallback.ID = targetID
	fallback.Text = text
	fallback.Type = kind
	fallback.Media = media
	fallback.Edited = true
	fallback.EditedAt = editedAt
	s.messages.Add(fallback)
	s.bus.Publish("message", fallback)
	return true
}

// unwrapEdit extracts the edited content from a decrypted message.
//
// DecryptSecretEncryptedMessage returns the inner waE2E.Message, which for a
// MESSAGE_EDIT carries a ProtocolMessage whose EditedMessage (a
// FutureProofMessage) wraps the replacement content. Only MESSAGE_EDIT is
// handled here; poll and event edits have their own shapes and are out of
// scope until someone asks for them.
func unwrapEdit(decrypted *waE2E.Message) (targetID, text, kind string, media *msgstore.Media, ok bool) {
	if decrypted == nil {
		return "", "", "", nil, false
	}

	proto := decrypted.GetProtocolMessage()
	if proto == nil || proto.GetType() != waE2E.ProtocolMessage_MESSAGE_EDIT {
		return "", "", "", nil, false
	}

	targetID = proto.GetKey().GetID()
	if targetID == "" {
		return "", "", "", nil, false
	}

	inner := proto.GetEditedMessage()
	if inner == nil {
		return "", "", "", nil, false
	}

	kind, text, media = classify(inner)
	return targetID, text, kind, media, true
}

// truncateEdit shortens edit text for the log line.
func truncateEdit(s string) string {
	const max = 80
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
