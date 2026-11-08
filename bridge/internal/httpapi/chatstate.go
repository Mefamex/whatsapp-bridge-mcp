// whatsapp-bridge-project/bridge-go/internal/httpapi/chatstate.go
//
// Sohbet durumu uçları: /mute ve /star.
//
// İkisi de mesaj DEĞİL, WhatsApp'ın kendi sohbet ayarlarıdır. Bu ayrım
// önemli: mesaj gönderimi `/send` üzerinden, susturma/favori ise
// app-state yaması üzerinden gider ve köprünün store'una düşmez.

package httpapi

import (
	"net/http"
	"strings"
	"time"
)

// muteRequest is the /mute payload.
//
// Süre metin olarak gelir ("8h", "7d", "30d"). Saniye sayısı göndermek
// kullanıcıyı hesap yapmaya zorlar; WhatsApp'ın kendi arayüzü de metin
// kullanıcı dostu birimler sunar.
type muteRequest struct {
	JID    string `json:"jid"`
	Muted  bool   `json:"muted"`
	For    string `json:"for,omitempty"` // "8h", "7d", "30d"; boşsa kalıcı
	Unmute bool   `json:"unmute"`        // kısayol: muted'i tersine çevirir
}

// starRequest is the /star payload.
type starRequest struct {
	JID       string `json:"jid"`
	MessageID string `json:"message_id"`
	Sender    string `json:"sender,omitempty"`
	FromMe    bool   `json:"from_me"`
	Starred   bool   `json:"starred"`
	Unstar    bool   `json:"unstar"`
}

// mute silences or unsilences a chat.
func (s *Server) mute(w http.ResponseWriter, r *http.Request) {
	var req muteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.JID == "" {
		http.Error(w, "jid required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	// unset ve unmute aynı iş: susturmayı KALDIR. İkisi de verilmişse
	// çelişkidir ve susturmayı kaldırmak daha güvenli (kullanıcı
	// bildirim almak istiyor).
	muted := req.Muted && !req.Unmute

	sure, err := sureSustur(req.For)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.wa.MuteChat(r.Context(), req.JID, muted, sure); err != nil {
		http.Error(w, "Failed to change mute: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if muted {
		writeJSON(w, map[string]any{
			"ok":    true,
			"muted": true,
			// Süre OKUNABİLİR metin olarak döner, ham nanosaniye olarak değil.
			// Canlı denemede görüldü: "until":3600000000000 — 1 saat, ama
			// kimse bunu okuyamaz. Süre metni aynı zamanda "kalıcı" ile
			// ayırt edilir (özel bir değer).
			"until": sureMetni(sure),
		})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "muted": false})
}

// chatState reads a chat's settings back, so a write can be verified instead
// of trusted. GET /chat-state?jid=...
func (s *Server) chatState(w http.ResponseWriter, r *http.Request) {
	rawJID := r.URL.Query().Get("jid")
	if rawJID == "" {
		http.Error(w, "jid required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	jid := strings.TrimSpace(rawJID)

	// sohbetJID normalizasyonu (LID ekleme, bozuk kimlik reddi) servis
	// katmanında; burada yalnız aynı biçimi yanıtta kullanmak için durum
	// okunuyor, kimlik yeniden kurulmuyor.
	durum, err := s.wa.ChatSettings(r.Context(), jid)
	if err != nil {
		http.Error(w, "Failed to read chat state: "+err.Error(), http.StatusInternalServerError)
		return
	}

	yanit := map[string]any{
		// Kanonik kimlik: durumun okunduğu sohbet. İstek "905451483283"
		// ise burada tam biçim görünür.
		"jid":      durum.JID,
		"muted":    durum.Muted(),
		"pinned":   durum.Pinned,
		"archived": durum.Archived,
	}
	if durum.Muted() {
		yanit["muted_until"] = durum.MutedUntil.UTC().Format(time.RFC3339)
	} else {
		yanit["muted_until"] = ""
	}
	writeJSON(w, yanit)
}

// star marks a message as a favourite.
func (s *Server) star(w http.ResponseWriter, r *http.Request) {
	var req starRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.JID == "" || req.MessageID == "" {
		http.Error(w, "jid and message_id required", http.StatusBadRequest)
		return
	}
	if !s.requireConnection(w) {
		return
	}

	starred := req.Starred && !req.Unstar

	if err := s.wa.StarMessage(r.Context(), req.JID, req.Sender, req.MessageID, req.FromMe, starred); err != nil {
		http.Error(w, "Failed to change star: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "starred": starred})
}

// sureMetni, süreyi insanın okuyabileceği biçime çevirir.
//
// NEDEN: time.Duration bir sayı olarak JSON'a girdiğinde 1 saat "3600000000000"
// olur. Bu doğru ama kullanılamaz — kimse 1 saatin nanosaniye olduğunu
// tanımaz. Yanıtta insan biçimi döner: "8h0m0s", "168h0m0s".
//
// Kalıcı susturma 0 gelir ve "kalıcı" yazılır; ayrım önemli çünkü kullanıcı
// "süresiz" dediğinde ne kadar bekleyeceğini bilmelidir.
func sureMetni(sure time.Duration) string {
	if sure <= 0 {
		return kaliciSureAdi
	}
	return sure.String()
}

// kaliciSureAdi, süresiz susturmanın yanıttaki karşılığı.
//
// Servis tarafında kalıcı 365 güne çevrilir (WhatsApp'ta kalıcı susturma
// yok, sonsuz süre bazı istemcilerde susturmayı KALDIRIYOR). Yanıtta "365
// gün" yazmak kullanıcıyı yanıltırdı: o 1 yıl sonra kendiliğinden açılacak,
// kullanıcı ise süresiz sandı.
const kaliciSureAdi = "kalıcı"

// sureSustur, kullanıcının yazdığı süreyi süreye çevirir.
//
// Boş = kalıcı (0 döner, servis tarafında bir yıla çevrilir).
// Kabul: "8h", "24h", "7d", "1w", "30d", "1y" ve bunların "h" "d" "w" "y"
// varyantları. Büyük harf ve boşluk kabul edilir (" 8 H ").
//
// Neden ayrıştırma elle: süre birimi yanlış yorumlanırsa ("8d" → 8 saniye)
// kullanıcı bir haftalık susturma istiyorken saniyeler sonra bildirim
// almaya başlar ve hatanın nerede olduğunu bulamaz.
func sureSustur(girilen string) (time.Duration, error) {
	metin := normalizeSure(girilen)
	if metin == "" {
		return 0, nil
	}

	var sayi int
	var birim string

	// Sondaki harfler birimi, geri kalanı sayıdır.
	i := len(metin)
	for i > 0 && !isRakam(metin[i-1]) {
		i--
	}
	birim = metin[i:]
	if sayiAlan := metin[:i]; sayiAlan != "" {
		sayi = atoiGuvenli(sayiAlan)
	}
	if sayi <= 0 {
		return 0, hataliSure(hataMetni(girilen))
	}

	switch birim {
	case "h", "saat":
		return time.Duration(sayi) * time.Hour, nil
	case "d", "gun":
		return time.Duration(sayi) * 24 * time.Hour, nil
	case "w", "hafta":
		return time.Duration(sayi) * 7 * 24 * time.Hour, nil
	case "a", "y", "yil":
		return time.Duration(sayi) * 365 * 24 * time.Hour, nil
	case "dk", "m", "dk.":
		return time.Duration(sayi) * time.Minute, nil
	default:
		return 0, hataliSure(hataMetni(girilen))
	}
}

// normalizeSure, girdiyi küçük harfe indirir ve boşlukları siler.
func normalizeSure(girilen string) string {
	sonuc := make([]byte, 0, len(girilen))
	for i := 0; i < len(girilen); i++ {
		c := girilen[i]
		switch {
		case c == ' ' || c == '\t':
			continue
		case c >= 'A' && c <= 'Z':
			sonuc = append(sonuc, c+('a'-'A'))
		default:
			sonuc = append(sonuc, c)
		}
	}
	return string(sonuc)
}

func isRakam(c byte) bool { return c >= '0' && c <= '9' }

// atoiGuvenli, rakam dışı karakterde durur (taşma paniği yerine).
func atoiGuvenli(s string) int {
	var n int
	for i := 0; i < len(s); i++ {
		if !isRakam(s[i]) {
			return 0
		}
		n = n*10 + int(s[i]-'0')
		if n > 1_000_000 {
			return 0
		}
	}
	return n
}

// hataMetni ve hataliSure, ayrıştırma hatalarını tek yerde toplar.
func hataMetni(girilen string) string {
	return "süre anlaşılamadı: " + girilen + " (örn. 8h, 7d, 1y)"
}

func hataliSure(mesaj string) error { return &hataSur{mesaj} }

type hataSur struct{ mesaj string }

func (h *hataSur) Error() string { return h.mesaj }
