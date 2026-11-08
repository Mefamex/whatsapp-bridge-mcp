// whatsapp-bridge-project/bridge-go/internal/wabridge/poll.go

package wabridge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"go.mau.fi/whatsmeow/types/events"

	"whatsapp-bridge/internal/msgstore"
)

// Poll sınırları WhatsApp'tan gelir: en az 2 seçenek, en çok 12.
// Kaynak: whatsmeow BuildPollCreation + canlı vaka (2026-10-06 02:47,
// telefon V3 anket gönderdi, köprü `[unsupported]` düşürdü).
const (
	minPollOptions = 2
	maxPollOptions = 12
)

// PollRequest, giden bir anketi tanımlar.
type PollRequest struct {
	To              string
	Name            string
	Options         []string
	SelectableCount int
}

// SendPoll bir anket gönderir ve çözümlenmiş JID'yi döndürür.
//
// Doğrulama istemciden ÖNCE yapılır: geçersiz istek hiçbir ağ işi
// tetiklemez, bu yüzden saf birim testiyle de denetlenebilir.
func (s *Service) SendPoll(ctx context.Context, req PollRequest) (string, error) {
	secenekler, sayilabilir, err := HazirlaAnket(req)
	if err != nil {
		return "", err
	}

	target, err := s.resolveTarget(ctx, req.To)
	if err != nil {
		return "", err
	}

	msg := s.client.BuildPollCreation(strings.TrimSpace(req.Name), secenekler, sayilabilir)
	resp, err := s.client.SendMessage(ctx, target, msg)
	if err != nil {
		return "", err
	}

	metin := renderPoll(strings.TrimSpace(req.Name), secenekler)
	s.recordOwnAdi(target, req.To, resp, "poll", metin, nil)
	return target.String(), nil
}

// HazirlaAnket isteği doğrular ve gönderime hazır hâle getirir.
// Hem HTTP katmanı (400 için) hem SendPoll kullanır: kural tek yerdedir.
func HazirlaAnket(req PollRequest) ([]string, int, error) {
	if strings.TrimSpace(req.To) == "" {
		return nil, 0, fmt.Errorf("to required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, 0, fmt.Errorf("name required")
	}
	secenekler := temizSecenekler(req.Options)
	if len(secenekler) < minPollOptions {
		return nil, 0, fmt.Errorf("at least %d options required, got %d", minPollOptions, len(secenekler))
	}
	if len(secenekler) > maxPollOptions {
		return nil, 0, fmt.Errorf("at most %d options allowed, got %d", maxPollOptions, len(secenekler))
	}

	sayilabilir := req.SelectableCount
	if sayilabilir <= 0 {
		sayilabilir = 1
	}
	if sayilabilir > len(secenekler) {
		sayilabilir = len(secenekler)
	}
	return secenekler, sayilabilir, nil
}

// temizSecenekler boş ve yinelenen seçenekleri atar, sırayı korur.
// WhatsApp boş seçeneği sessizce yutar; sayı sınırına boşlar dahil
// olsaydı kullanıcı "2 seçenek verdim" sanır, tek seçenek giderdi.
func temizSecenekler(giren []string) []string {
	gorulen := make(map[string]bool, len(giren))
	cikti := make([]string, 0, len(giren))
	for _, secenek := range giren {
		secenek = strings.TrimSpace(secenek)
		if secenek == "" || gorulen[secenek] {
			continue
		}
		gorulen[secenek] = true
		cikti = append(cikti, secenek)
	}
	return cikti
}

// renderPoll anketi store/log için tek metne indirir.
// classify'deki gelen yönle AYNI biçim kullanılır: iki yönde de
// "📊 soru\no1\no2" görülür, yönden değil içerikten tanınır.
func renderPoll(ad string, secenekler []string) string {
	var b strings.Builder
	b.WriteString("📊 ")
	b.WriteString(ad)
	for _, secenek := range secenekler {
		b.WriteString("\n• ")
		b.WriteString(secenek)
	}
	return b.String()
}

// ayristirSecenekler renderPoll metninden orijinal seçenekleri ayıklar.
func ayristirSecenekler(metin string) []string {
	var secenekler []string
	satirlar := strings.Split(metin, "\n")
	for _, satir := range satirlar {
		satir = strings.TrimSpace(satir)
		if strings.HasPrefix(satir, "• ") {
			secenekler = append(secenekler, strings.TrimPrefix(satir, "• "))
		}
	}
	return secenekler
}

// handlePollUpdate gelen bir anket oyunu deşifre eder ve normalize mesaja çevirir.
func (s *Service) handlePollUpdate(v *events.Message) (msgstore.Message, bool) {
	pollUpdate := v.Message.GetPollUpdateMessage()
	if pollUpdate == nil || s.client == nil {
		return msgstore.Message{}, false
	}

	pollVote, err := s.client.DecryptPollVote(context.Background(), v)
	if err != nil {
		log.Printf("⚠️ anket oyu deşifre edilemedi: %v", err)
		return msgstore.Message{}, false
	}

	pollKey := pollUpdate.GetPollCreationMessageKey()
	var pollID string
	if pollKey != nil {
		pollID = pollKey.GetID()
	}

	// Orijinal anket mesajından seçenekleri çöz
	var secilenler []string
	var soruMetni string
	if pollID != "" && s.messages != nil {
		if origMsg, ok := s.messages.ByID(pollID); ok {
			soruMetni = origMsg.Text
			secenekler := ayristirSecenekler(origMsg.Text)
			hashHaritasi := make(map[string]string)
			for _, sec := range secenekler {
				h := sha256.Sum256([]byte(sec))
				hashHaritasi[hex.EncodeToString(h[:])] = sec
			}

			for _, optHash := range pollVote.GetSelectedOptions() {
				hashHex := hex.EncodeToString(optHash)
				if secenekAdi, varMi := hashHaritasi[hashHex]; varMi {
					secilenler = append(secilenler, secenekAdi)
				}
			}
		}
	}

	secimMetni := strings.Join(secilenler, ", ")
	if secimMetni == "" {
		secimMetni = "Oylama güncellendi"
	}

	from := v.Info.Chat.String()
	sender := v.Info.Sender.String()
	direction := "in"
	if v.Info.IsFromMe {
		direction = "out"
	}

	var quoted *msgstore.QuotedMessage
	if pollID != "" {
		quoted = &msgstore.QuotedMessage{
			ID:   pollID,
			Text: soruMetni,
		}
	}

	return msgstore.Message{
		ID:        v.Info.ID,
		Type:      "poll_vote",
		Direction: direction,
		From:      from,
		Sender:    sender,
		Timestamp: v.Info.Timestamp.Unix(),
		Text:      secimMetni,
		Quoted:    quoted,
	}, true
}
