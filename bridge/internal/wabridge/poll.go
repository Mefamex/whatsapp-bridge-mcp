// whatsapp-bridge-project/bridge-go/internal/wabridge/poll.go

package wabridge

import (
	"context"
	"fmt"
	"strings"
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
