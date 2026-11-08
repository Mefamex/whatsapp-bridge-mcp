// whatsapp-bridge-project/bridge-go/internal/httpapi/messages.go

package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/msgstore"
)

// historyWait bounds how long a request blocks waiting for WhatsApp history
// when the buffer is COMPLETELY empty — the only case where waiting can still
// produce an answer that is not already in hand.
const historyWait = 10 * time.Second

// historyGrace bounds the extra wait when the buffer already holds messages.
//
// ÖLÇÜM GEREKÇESİ: historyWait (10 sn) ile tetikleyicinin çift gönderim
// denetimindeki bakış bütçesi (8 sn) ÇELİŞİYORDU. Çağıran 8 saniyede
// vazgeçiyor, sunucu 10 saniye bekliyor: istek her seferinde zaman aşımına
// uğruyor ve denetim "geçmiş alınamadı" deyip bastırmadan geçiyordu — yani
// koruma sessizce ÖLÜYDÜ.
//
// Tampon doluysa beklemeye değer çok az: elde mesaj var, soru "daha eskisi
// de geldi mi" değil. historyGrace, tam senkronizasyon şansa yer bırakır ve
// yine de çağıranın bütçesinin ALTINDA kalır (1,5 sn < 8 sn).
const historyGrace = 1500 * time.Millisecond

// historyPoll is the re-check interval while waiting for a history sync.
const historyPoll = 200 * time.Millisecond

// maxHistoryCount caps the ?count= parameter.
const maxHistoryCount = 100

// messagesHandler serves recent messages, topping the buffer up from WhatsApp
// history when the chat is not fully cached locally.
//
// ?topup=off → WhatsApp'a hiç gitmez, yerel tampon anında döner. Hızlı
// sorgular (örn. "bu metin az önce gitti mi?") için: tam geçmiş istemiyorlar,
// sadece hızlı yanıt istiyorlar.
func (s *Server) messagesHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireConnection(w) {
		return
	}

	rawJID := mux.Vars(r)["jid"]
	if rawJID == "" {
		http.Error(w, "jid required", http.StatusBadRequest)
		return
	}

	count := historyCount(r)
	jid := appcfg.NormalizeJID(rawJID)

	messages := s.messages.Recent(jid, count)
	if len(messages) < count && topUpIstekVar(r) {
		messages = s.topUpFromWhatsApp(jid, count, messages)
	}

	writeJSON(w, map[string]any{
		"jid":      jid,
		"count":    len(messages),
		"messages": messages,
	})
}

// topUpIstekVar, ?topup=off ile tamamlanmışsa false döner.
//
// Bu parametre YENİDEN SORMAYLA değil, ZAMANLA İLGİLİ: varsayılan olarak
// uç, tampon yetmiyorsa WhatsApp'a gider ve 10 saniye bekleyebilir. 10 sn,
// çağıranın (tetikleyicinin çift gönderim denetimi) kendi baktığı 8 saniyeden
// uzun — yani varsayılan davranış, o denetimi TIKAN bir cevap üretir.
func topUpIstekVar(r *http.Request) bool {
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("topup"))) {
	case "off", "0", "hayir", "no":
		return false
	}
	return true
}

// topUpFromWhatsApp requests older messages and polls the buffer until they
// arrive or the wait budget is spent. It returns the freshest view available.
//
// BEKLEME BÜTÇESİ UYUMLUĞU: tampon HİÇ boş değilse yalnızca kısa bir süre
// beklenir (historyGrace), istenen sayıya hiç ulaşılamasa bile eldeki
// mesajlar döndürülür. Tampon boşsa tam bütçe (historyWait) harcanır, çünkü
// o durumda döndürülecek hiçbir şey yoktur ve bir şeyin gelmesi beklemeye
// değer.
//
// Neden: tam bütçe her istekte yanıtı 10 saniyeye sabitliyordu. Ölçüldü:
// /messages {count=12} tam olarak 10,12 sn sürüyordu, üstelik tetikleyicinin
// kendi bakış bütçesi 8 sn olduğu için çift gönderim denetimi her seferinde
// "geçmiş alınamadı" deyip sessizce çalışmıyordu.
func (s *Server) topUpFromWhatsApp(jid string, count int, current []msgstore.Message) []msgstore.Message {
	log.Printf("📜 [HISTORY] Requesting %d messages for %s (buffered %d)", count, jid, len(current))

	if err := s.wa.RequestHistory(jid, count); err != nil {
		log.Printf("⚠️ [HISTORY] %v", err)
		return current
	}

	budget := historyWait
	if len(current) > 0 {
		budget = historyGrace
	}

	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		time.Sleep(historyPoll)

		if fresh := s.messages.Recent(jid, count); len(fresh) >= count {
			return fresh
		}
	}
	return s.messages.Recent(jid, count)
}

// historyCount parses ?count=, clamping it to a sane range.
func historyCount(r *http.Request) int {
	count := 20
	if raw := r.URL.Query().Get("count"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			count = parsed
		}
	}
	if count > maxHistoryCount {
		count = maxHistoryCount
	}
	return count
}

// mediaDownload streams a buffered message's attachment.
func (s *Server) mediaDownload(w http.ResponseWriter, r *http.Request) {
	if !s.requireConnection(w) {
		return
	}

	id := mux.Vars(r)["messageId"]
	if id == "" {
		http.Error(w, "messageId required", http.StatusBadRequest)
		return
	}

	msg, ok := s.messages.ByID(id)
	if !ok {
		http.Error(w, "Message not found in buffer", http.StatusNotFound)
		return
	}
	if msg.Media == nil {
		http.Error(w, "Message has no media", http.StatusBadRequest)
		return
	}

	data, err := s.wa.DownloadMedia(r.Context(), msg.Media)
	if err != nil {
		http.Error(w, "Failed to download media: "+err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := msg.Media.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	_, _ = w.Write(data)
}
