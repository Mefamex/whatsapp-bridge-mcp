// whatsapp-bridge-project/bridge-go/internal/wabridge/send.go

package wabridge

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/msgstore"
)

// MediaRequest describes an outbound attachment.
type MediaRequest struct {
	To       string
	Data     []byte
	MimeType string
	Caption  string
	FileName string
}

// SendText delivers a plain text message and returns the resolved JID.
func (s *Service) SendText(ctx context.Context, to, text string) (string, error) {
	target, err := s.resolveTarget(ctx, to)
	if err != nil {
		return "", err
	}

	msg := &waE2E.Message{Conversation: proto.String(text)}
	resp, err := s.client.SendMessage(ctx, target, msg)
	if err != nil {
		return "", err
	}

	s.recordOwnAdi(target, to, resp, "text", text, nil)
	return target.String(), nil
}

// SendReaction adds or removes an emoji reaction on a message.
func (s *Service) SendReaction(ctx context.Context, to, targetID, emoji, senderJID string) error {
	target, err := s.resolveTarget(ctx, to)
	if err != nil {
		return err
	}

	sender := target
	if senderJID != "" {
		if resolvedSender, err := s.resolveTarget(ctx, senderJID); err == nil {
			sender = resolvedSender
		}
	} else if s.messages != nil {
		// Mesaj bellekte varsa göndereni otomatik tespit et
		if msg, ok := s.messages.ByID(targetID); ok {
			if msg.Direction == "out" && s.client != nil && s.client.Store != nil && s.client.Store.ID != nil {
				sender = s.client.Store.ID.ToNonAD()
			} else if msg.Sender != "" {
				if parsed, err := types.ParseJID(msg.Sender); err == nil {
					sender = parsed
				}
			}
		}
	}

	msg := s.client.BuildReaction(target, sender, types.MessageID(targetID), emoji)
	_, err = s.client.SendMessage(ctx, target, msg)
	return err
}

// RevokeMessage unsends (deletes for everyone) an outgoing message.
func (s *Service) RevokeMessage(ctx context.Context, to, targetID string) error {
	target, err := s.resolveTarget(ctx, to)
	if err != nil {
		return err
	}

	msg := s.client.BuildRevoke(target, types.EmptyJID, types.MessageID(targetID))
	_, err = s.client.SendMessage(ctx, target, msg)
	return err
}

// resolveTarget turns a caller-supplied address into a JID WhatsApp will
// actually deliver to.
func (s *Service) resolveTarget(ctx context.Context, to string) (types.JID, error) {
	return resolveTarget(ctx, s.client.Store.LIDs, to)
}

// takmaAdlar, istenen kimlik çözülmüş kinden farklıysa onu da döner.
//
// Boş veya aynıysa takma ad ÜRETİLMEZ: boş bir takma ad, mesajı her sohbette
// göstermeye yol açardı (belongsTo boş string'le eşleşmez ama yanlış eşleşme
// riski yaratır ve hata ayıklamayı zorlaştırır).
func takmaAdlar(istendi string, target types.JID) []string {
	kesilmis := strings.TrimSpace(istendi)
	if kesilmis == "" || kesilmis == target.String() {
		return nil
	}
	return []string{kesilmis}
}

// lidResolver is the slice of store.LIDStore this file needs. Naming it keeps
// resolveTarget testable without standing up a real whatsmeow client.
type lidResolver interface {
	GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error)
}

// resolveTarget turns a caller-supplied address into a JID WhatsApp will
// actually deliver to.
//
// The important case is a LID chat. WhatsApp delivers incoming DMs in
// "LID chats" addressed as 123456@lid, and replying to that raw LID is
// unreliable: the message leaves the server without an error but does not
// surface in the conversation the human sees. whatsmeow records the
// LID↔phone mapping as messages arrive (see Client.StoreLIDPNMapping), so we
// resolve the LID back to the phone JID before sending and the reply lands
// in the same thread the message came from.
//
// An unresolvable LID is passed through unchanged. That is deliberate: a
// missing mapping should not block sending, and the server may still accept
// the raw LID — it just may not be visible to the user.
func resolveTarget(ctx context.Context, lids lidResolver, to string) (types.JID, error) {
	target, err := parseTarget(to)
	if err != nil {
		return types.JID{}, err
	}

	if target.Server != types.HiddenUserServer {
		return target, nil
	}

	pn, err := lids.GetPNForLID(ctx, target)
	if err != nil || pn.Server != types.DefaultUserServer || pn.User == "" {
		log.Printf("⚠️ [SEND] LID %s çözümlenemedi (%v); ham LID'e gönderiliyor", target, err)
		return target, nil
	}

	log.Printf("📤 [SEND] LID %s → %s çözüldü", target, pn)
	return pn, nil
}

// SendMedia uploads an attachment and sends it as the right message type.
func (s *Service) SendMedia(ctx context.Context, req MediaRequest) (string, string, error) {
	target, err := s.resolveTarget(ctx, req.To)
	if err != nil {
		return "", "", fmt.Errorf("media metadata could not be built: %w", err)
	}

	uploaded, err := s.client.Upload(ctx, req.Data, mediaTypeFromMime(req.MimeType))
	if err != nil {
		return "", "", fmt.Errorf("upload failed: %w", err)
	}

	kind, text, media := describeMedia(req, &uploaded)
	resp, err := s.client.SendMessage(ctx, target, buildMediaMessage(req, &uploaded))
	if err != nil {
		return "", "", err
	}

	s.recordOwnAdi(target, req.To, resp, kind, text, media)
	return target.String(), uploaded.DirectPath, nil
}

// recordOwn buffers a message we sent. whatsmeow does not emit an inbound
// event for our own traffic, so without this the buffer stays empty and
// history sync has no anchor to walk back from.
//
// istendi, target'ın ÇÖZÜLMÜŞ hâlidir; ikisi farklı olabilir (LID → telefon).
// ikisi de kaydedilir: bkz. recordOwnAliasNotu.
func (s *Service) recordOwn(target types.JID, resp whatsmeow.SendResponse, kind, text string, media *msgstore.Media) {
	s.recordOwnAdi(target, "", resp, kind, text, media)
}

// recordOwnAdi, mesajı istenen kimlikle de kaydeder. LID sohbetlerinde
// istENEN kimlik `123@lid`, ÇÖZÜLMÜŞ olan `905…@s.whatsapp.net` olur.
func (s *Service) recordOwnAdi(target types.JID, istendi string, resp whatsmeow.SendResponse, kind, text string, media *msgstore.Media) {
	if text == "" {
		text = "[media]"
	}

	// recordOwnAliasNotu
	//
	// ÖLÇÜM (2026-10-06, canlı): LID'li bir sohbette köprüye
	// `47575939964948@lid` diye sorulduğu HALDE /messages o kimlikte 0
	// mesaj döndürüyordu; aynı mesajlar `905…@s.whatsapp.net` altta duruyordu.
	//
	// Sebep: giden mesajın From alanı ÇÖZÜLMÜŞ kimlikle yazılıyor, gelen
	// mesajınki ise LID ile. Store.belongsTo tam eşleşme arıyor.
	//
	// Bu sessiz bir boşluk değil, iki özelliği öldürüyordu:
	//   1. ÇİFT GÖNDERİM DENETİMİ: yalnız fromMe mesajlara bakar, LID
	//      görünümünde fromMe mesaj SAYI YOK → bastırma hiç çalışmıyor.
	//   2. AJANIN GEÇMİŞ PENCERESİ: ajan kendi önceki cevabını görmüyor.
	//      "Kaç tane vardı?" sorusuna "3" dedikten sonra "kaçtı?" diye
	//      sorulduğunda önceki cevabı hatırlamıyor.
	//
	// Çözüm: mesajı İSTENEN kimlikle de kaydet (takma ad). Eklenebilir,
	// mevcut davranışı bozmaz: kimse tekrar sorgulamıyor.
	s.messages.Add(msgstore.Message{
		ID:        string(resp.ID),
		Type:      kind,
		Direction: "out",
		From:      target.String(),
		Sender:    s.SelfJID(),
		Timestamp: resp.Timestamp.Unix(),
		Text:      text,
		IsFromMe:  true,
		IsGroup:   target.Server == types.GroupServer,
		Media:     media,
		TakmaAd:   takmaAdlar(istendi, target),
	})
	s.bus.Publish("message_sent", map[string]any{
		"id":        string(resp.ID),
		"to":        target.String(),
		"type":      kind,
		"timestamp": resp.Timestamp.Unix(),
		"text":      text,
	})
}

// describeMedia renders the store-side view of an outbound attachment.
func describeMedia(req MediaRequest, up *whatsmeow.UploadResponse) (kind, text string, media *msgstore.Media) {
	media = &msgstore.Media{
		URL:           up.URL,
		DirectPath:    up.DirectPath,
		MediaKey:      base64.StdEncoding.EncodeToString(up.MediaKey),
		FileEncSHA256: base64.StdEncoding.EncodeToString(up.FileEncSHA256),
		FileSHA256:    base64.StdEncoding.EncodeToString(up.FileSHA256),
		FileLength:    up.FileLength,
		MimeType:      req.MimeType,
	}

	switch {
	case strings.HasPrefix(req.MimeType, "image/"):
		return "image", req.Caption, media
	case strings.HasPrefix(req.MimeType, "video/"):
		return "video", req.Caption, media
	case strings.HasPrefix(req.MimeType, "audio/"):
		return "audio", "", media
	default:
		if req.Caption != "" {
			return "document", req.Caption, media
		}
		return "document", req.FileName, media
	}
}

// parseTarget normalises a phone number or JID and parses it.
func parseTarget(to string) (types.JID, error) {
	return types.ParseJID(appcfg.NormalizeJID(to))
}

// SendCall initiates an outbound WhatsApp call (ring the user).
// durationSec: how long to ring (0 = until answered/hangup).
func (s *Service) SendCall(ctx context.Context, to string, durationSec int) error {
	target, err := s.resolveTarget(ctx, to)
	if err != nil {
		return fmt.Errorf("target could not be resolved: %w", err)
	}

	log.Printf("📞 [CALL] Outbound call to %s (duration: %ds)", target.String(), durationSec)

	// Build call offer stanza:
	// <call to="target" id="call-id" from="self">
	//   <offer platform="web" version="1" call-id="..." t="...">
	//     <encrypted_sdp>...</encrypted_sdp>
	//   </offer>
	// </call>
	// Note: The actual SDP generation + encryption happens in the sidecar/WASM.
	// We send a minimal offer stanza; the sidecar/WASM will handle the crypto.
	callID := fmt.Sprintf("call_%d", time.Now().UnixMilli())
	selfJID := s.SelfJID()

	callNode := &waBinary.Node{
		Tag: "call",
		Attrs: map[string]any{
			"to":   target.String(),
			"id":   callID,
			"from": selfJID,
		},
		Content: []*waBinary.Node{
			{
				Tag: "offer",
				Attrs: map[string]any{
					"platform": "web",
					"version":  "1",
					"call-id":  callID,
					"t":        fmt.Sprintf("%d", time.Now().Unix()),
				},
				Content: []*waBinary.Node{
					{
						Tag:     "encrypted_sdp",
						Content: []byte("placeholder_sdp"), // WASM will replace with encrypted SDP
					},
				},
			},
		},
	}

	log.Printf("📞 [CALL] Sending call offer to %s (callID: %s)", target.String(), callID)

	// Send via sidecar IPC (call_stanza_out)
	if err := s.SendCallStanza(callNode); err != nil {
		return fmt.Errorf("failed to send call stanza: %w", err)
	}

	log.Printf("✅ [CALL] Call offer sent to sidecar for %s", target.String())
	return nil
}

// LocationRequest describes an outbound location pin.
type LocationRequest struct {
	To        string
	Latitude  float64
	Longitude float64
	Name      string
	Address   string
}

// SendLocation sends a location pin and returns the resolved target JID.
func (s *Service) SendLocation(ctx context.Context, req LocationRequest) (string, error) {
	if strings.TrimSpace(req.To) == "" {
		return "", fmt.Errorf("to required")
	}
	target, err := s.resolveTarget(ctx, req.To)
	if err != nil {
		return "", err
	}

	locMsg := &waE2E.LocationMessage{
		DegreesLatitude:  proto.Float64(req.Latitude),
		DegreesLongitude: proto.Float64(req.Longitude),
	}
	if strings.TrimSpace(req.Name) != "" {
		locMsg.Name = proto.String(strings.TrimSpace(req.Name))
	}
	if strings.TrimSpace(req.Address) != "" {
		locMsg.Address = proto.String(strings.TrimSpace(req.Address))
	}

	msg := &waE2E.Message{LocationMessage: locMsg}
	resp, err := s.client.SendMessage(ctx, target, msg)
	if err != nil {
		return "", err
	}

	metin := fmt.Sprintf("📍 Konum: %.6f, %.6f", req.Latitude, req.Longitude)
	if req.Name != "" {
		metin += fmt.Sprintf(" (%s)", req.Name)
	}
	s.recordOwnAdi(target, req.To, resp, "location", metin, nil)
	return target.String(), nil
}

// ContactRequest describes an outbound vCard contact card.
type ContactRequest struct {
	To    string
	Name  string
	Phone string
	Vcard string
}

// SendContact sends a contact card and returns the resolved target JID.
func (s *Service) SendContact(ctx context.Context, req ContactRequest) (string, error) {
	if strings.TrimSpace(req.To) == "" {
		return "", fmt.Errorf("to required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return "", fmt.Errorf("name required")
	}
	target, err := s.resolveTarget(ctx, req.To)
	if err != nil {
		return "", err
	}

	vcard := req.Vcard
	if strings.TrimSpace(vcard) == "" {
		vcard = fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;TYPE=CELL:%s\nEND:VCARD", req.Name, req.Phone)
	}

	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(req.Name),
			Vcard:       proto.String(vcard),
		},
	}
	resp, err := s.client.SendMessage(ctx, target, msg)
	if err != nil {
		return "", err
	}

	metin := fmt.Sprintf("📇 Kişi Kartı: %s", req.Name)
	s.recordOwnAdi(target, req.To, resp, "contact", metin, nil)
	return target.String(), nil
}

// mediaTypeFromMime maps a MIME type to a whatsmeow upload category.
func mediaTypeFromMime(mime string) whatsmeow.MediaType {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return whatsmeow.MediaImage
	case strings.HasPrefix(mime, "video/"):
		return whatsmeow.MediaVideo
	case strings.HasPrefix(mime, "audio/"):
		return whatsmeow.MediaAudio
	default:
		return whatsmeow.MediaDocument
	}
}
