// whatsapp-bridge-project/bridge-go/internal/httpapi/chatstate_test.go
//
// /mute, /star ve /chat-state: sohbet durumu uçları.
//
// Bu uçlar mesaj gönderMEZ, WhatsApp'ın kendi ayarlarını değiştirir. Testte
// ölçülen şey: doğru servise doğru değerlerle gidip gitmediği ve — daha
// önemlisi — yazma işleminin OKUNABİLİR bir kanıtı olup olmadığı.
//
// "ok:true" tek başına kanıt DEĞİLDİR. Bu proje kuralı: gönderim başarısı
// teslimat kanıtı değildir. /chat-state bunu okunur kılar; testler de tam
// olarak o kombinasyonu sınar.

package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"whatsapp-bridge/internal/wabridge"
)

// ── yardımcılar ───────────────────────────────────────────────────────────

// istek, bir handler'a gövde gönderir ve kaydı döndürür.
func istek(t *testing.T, h http.HandlerFunc, govde string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(govde)))
	return rec
}

// cevap, JSON gövdeyi haritala olarak okur.
func cevap(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("yanıt JSON değil (%d): %s", rec.Code, rec.Body.String())
	}
	return m
}

// ── mute ──────────────────────────────────────────────────────────────────

func TestMuteSustururVeSuresiGider(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).mute,
		`{"jid":"905451483283@s.whatsapp.net","muted":true,"for":"8h"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d: %s", rec.Code, rec.Body.String())
	}
	if wa.muteJID != "905451483283@s.whatsapp.net" {
		t.Errorf("jid = %q", wa.muteJID)
	}
	if !wa.muteMuted {
		t.Error("susturulmadı")
	}
	if wa.muteSure != 8*time.Hour {
		t.Errorf("süre = %v, want 8h", wa.muteSure)
	}
	if got := cevap(t, rec)["muted"]; got != true {
		t.Errorf("yanıt muted = %v", got)
	}
}

// SÜRESİZ SUSTURMA 0 GÖNDERİLMELİ: servis kalıcıyı bir yıla çevirir. Kullanıcı
// "kalıcı" dediğinde 0 demek doğru; "365 gün" demek onun yerine karar
// vermek olurdu.
func TestMuteSuresizSusturmaSifirGider(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).mute,
		`{"jid":"a@s.whatsapp.net","muted":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d", rec.Code)
	}
	if wa.muteSure != 0 {
		t.Errorf("süre = %v, want 0 (kalıcı)", wa.muteSure)
	}
}

// SÜRE YANITTA İNSAN OKUNUR OLMALI.
//
// Canlı deneme bunu yakaladı: ilk sürüm time.Duration'ı doğrudan JSON'a
// yazıyordu ve yanıt şu çıkıyordu:
//
//	{"ok":true,"muted":true,"until":3600000000000}
//
// Değer DOĞRUYDU (3.6e12 nanosaniye = 1 saat) ama kimse bunu okuyamaz. Sessiz
// hataya dönüşmesi de mümkün: kullanıcı "bu ne?" deyip susturmayı açık
// sandığında kalır.
func TestMuteSuresizYanitindaKaliciYazar(t *testing.T) {
	rec := istek(t, newServer(&fakeWA{connected: true}, nil).mute,
		`{"jid":"a@s.whatsapp.net","muted":true}`)

	got, _ := cevap(t, rec)["until"].(string)
	if got != kaliciSureAdi {
		t.Errorf("until = %q, want %q", got, kaliciSureAdi)
	}
}

func TestMuteSuresiYanitindaInsanOkur(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).mute,
		`{"jid":"a@s.whatsapp.net","muted":true,"for":"8h"}`)

	got, _ := cevap(t, rec)["until"].(string)
	if got == "" {
		t.Fatal("until boş")
	}
	// Ham nanosaniye biçimi ("3600000000000") kabul edilmez: bu tam olarak
	// canlıda görülen ve okunamayan biçim.
	if strings.Count(got, "0") >= 6 && !strings.ContainsAny(got, "hms") {
		t.Errorf("until = %q: ham nanosaniye gibi görünüyor", got)
	}
	if got != (8 * time.Hour).String() {
		t.Errorf("until = %q, want %q", got, (8 * time.Hour).String())
	}
}

// Sessizliği KALDIRMA: unset ve unmute aynı iş. İkisi de verilmişse
// çelişki vardır; susturmayı kaldırmak güvenli taraftır (kullanıcı
// bildirim almak istiyor).
func TestMuteUnmuteSusturmayuKaldirir(t *testing.T) {
	for _, ad := range []string{"unset", "unmute"} {
		t.Run(ad, func(t *testing.T) {
			wa := &fakeWA{connected: true}
			var govde string
			if ad == "unset" {
				govde = `{"jid":"a@s.whatsapp.net","muted":false}`
			} else {
				govde = `{"jid":"a@s.whatsapp.net","unmute":true}`
			}

			rec := istek(t, newServer(wa, nil).mute, govde)
			if rec.Code != http.StatusOK {
				t.Fatalf("kod = %d", rec.Code)
			}
			if wa.muteMuted {
				t.Errorf("%s: hâlâ susturuldu", ad)
			}
			if wa.muteSure != 0 {
				t.Errorf("%s: süre = %v, want 0", ad, wa.muteSure)
			}
		})
	}
}

func TestMuteCiftBayrakCeliskisindeKaldirir(t *testing.T) {
	wa := &fakeWA{connected: true}
	// muted:true VE unmute:true → susturulmuş olmak istenmez.
	rec := istek(t, newServer(wa, nil).mute,
		`{"jid":"a@s.whatsapp.net","muted":true,"unmute":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d", rec.Code)
	}
	if wa.muteMuted {
		t.Error("çelişkide susturuldu — kaldırması bekleniyordu")
	}
}

func TestMuteJidYoksa400(t *testing.T) {
	rec := istek(t, newServer(&fakeWA{connected: true}, nil).mute, `{"muted":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
}

func TestMuteBozukGovde400(t *testing.T) {
	rec := istek(t, newServer(&fakeWA{connected: true}, nil).mute, `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
}

// Bağlantı yoksa 503: susturma "gönderildi" gibi görünmemeli.
func TestMuteBaglantiYoksa503(t *testing.T) {
	wa := &fakeWA{connected: false}
	rec := istek(t, newServer(wa, nil).mute, `{"jid":"a@s.whatsapp.net","muted":true}`)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("kod = %d, want 503", rec.Code)
	}
	if wa.muteMuted {
		t.Error("bağlantısızken susturuldu")
	}
}

func TestMuteServisHatasiniYuzeyeCikarir(t *testing.T) {
	wa := &fakeWA{connected: true, stateErr: errors.New("app state rejected")}
	rec := istek(t, newServer(wa, nil).mute, `{"jid":"a@s.whatsapp.net","muted":true}`)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("kod = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "app state rejected") {
		t.Errorf("gövde sebebi taşımıyor: %s", rec.Body.String())
	}
}

// ── star ──────────────────────────────────────────────────────────────────

func TestStarYildizlar(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).star,
		`{"jid":"a@s.whatsapp.net","message_id":"3EB0ABC","starred":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d: %s", rec.Code, rec.Body.String())
	}
	if wa.starID != "3EB0ABC" {
		t.Errorf("message_id = %q", wa.starID)
	}
	if !wa.starred {
		t.Error("yıldızlanmadı")
	}
	if got := cevap(t, rec)["starred"]; got != true {
		t.Errorf("yanıt starred = %v", got)
	}
}

func TestStarUnstarYildiziKaldirir(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).star,
		`{"jid":"a@s.whatsapp.net","message_id":"M1","starred":false}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d", rec.Code)
	}
	if wa.starred {
		t.Error("yıldız kaldırılmadı")
	}
}

// KİMLİKSİZ YILDIZ 400. Yıldız bir MESAJI işaretler; kimliksiz istek
// sessizce başarısız olurdu ve kullanıcı telefonunda "yıldızlamadı" görür,
// sebebini bilmez.
func TestStarKimliksiz400(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).star, `{"jid":"a@s.whatsapp.net","starred":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
	if wa.starID != "" {
		t.Error("kimliksiz istek servise gitti")
	}
}

func TestStarJidYoksa400(t *testing.T) {
	rec := istek(t, newServer(&fakeWA{connected: true}, nil).star, `{"message_id":"M1","starred":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
}

func TestStarBaglantiYoksa503(t *testing.T) {
	rec := istek(t, newServer(&fakeWA{connected: false}, nil).star,
		`{"jid":"a@s.whatsapp.net","message_id":"M1","starred":true}`)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("kod = %d, want 503", rec.Code)
	}
}

// Gönderen kimliği taşınmalı: LID'li sohbetlerde mesajın From'u gerekir,
// aksi hâlde yıldız sessizce kaybolur.
func TestStarGonderenKimligiTasir(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).star,
		`{"jid":"47575939964948@lid","message_id":"M1","sender":"47575939964948@lid","starred":true,"from_me":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d", rec.Code)
	}
	if wa.starSender != "47575939964948@lid" {
		t.Errorf("sender = %q", wa.starSender)
	}
	if !wa.fromMe {
		t.Error("from_me taşınmadı")
	}
}

// ── chat-state (yazmanın okunabilir kanıtı) ───────────────────────────────

func TestChatStateOkunurDurumuDoner(t *testing.T) {
	bitis := time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC)
	wa := &fakeWA{connected: true, chatState: wabridge.ChatState{
		JID:        "a@s.whatsapp.net",
		MutedUntil: bitis,
		Pinned:     true,
		Archived:   false,
	}}
	srv := newServer(wa, nil)
	rec := httptest.NewRecorder()
	srv.chatState(rec, httptest.NewRequest(http.MethodGet, "/chat-state?jid=a@s.whatsapp.net", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("kod = %d: %s", rec.Code, rec.Body.String())
	}
	yanit := cevap(t, rec)
	if yanit["muted"] != true {
		t.Errorf("muted = %v, want true", yanit["muted"])
	}
	if yanit["pinned"] != true {
		t.Errorf("pinned = %v, want true", yanit["pinned"])
	}
	if yanit["archived"] != false {
		t.Errorf("archived = %v, want false", yanit["archived"])
	}
	if yanit["muted_until"] != bitis.Format(time.RFC3339) {
		t.Errorf("muted_until = %v, want %s", yanit["muted_until"], bitis.Format(time.RFC3339))
	}
}

// YANIT KANONİK KİMLİĞİ YANKILAMALI, ham girdiyi değil.
//
// Canlı denemede görüldü: istek "?jid=905451483283" idi, yanıt da
// "jid":"905451483283" diyordu — oysa durum "905451483283@s.whatsapp.net"
// üzerinden okunmuştu. Bir kanıt endpoint'i başka sohbeti gösterip kendi
// girdisini yankılarsa, kanıt hangi sohbete ait belirsizdir.
func TestChatStateKanonikJidYankilar(t *testing.T) {
	wa := &fakeWA{connected: true, chatState: wabridge.ChatState{
		JID: "905451483283@s.whatsapp.net",
	}}
	srv := newServer(wa, nil)
	rec := httptest.NewRecorder()
	// Sunucusuz kimlik: köprü @s.whatsapp.net ekler.
	srv.chatState(rec, httptest.NewRequest(http.MethodGet, "/chat-state?jid=905451483283", nil))

	got, _ := cevap(t, rec)["jid"].(string)
	if got != "905451483283@s.whatsapp.net" {
		t.Errorf("jid = %q, want kanonik biçim", got)
	}
}

// SUSTURULMAMIŞ sohbette muted_until BOŞ OLMALI: sıfır zaman "1 Ocak 1970"
// yazmak kullanıcıya 56 yıllık bir susturma izlenimi verir.
func TestChatStateSusturulmamisZamanBos(t *testing.T) {
	wa := &fakeWA{connected: true}
	srv := newServer(wa, nil)
	rec := httptest.NewRecorder()
	srv.chatState(rec, httptest.NewRequest(http.MethodGet, "/chat-state?jid=a@s.whatsapp.net", nil))

	yanit := cevap(t, rec)
	if yanit["muted"] != false {
		t.Errorf("muted = %v, want false", yanit["muted"])
	}
	if yanit["muted_until"] != "" {
		t.Errorf("muted_until = %v, want boş", yanit["muted_until"])
	}
}

func TestChatStateJidYoksa400(t *testing.T) {
	srv := newServer(&fakeWA{connected: true}, nil)
	rec := httptest.NewRecorder()
	srv.chatState(rec, httptest.NewRequest(http.MethodGet, "/chat-state", nil))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
}

func TestChatStateBaglantiYoksa503(t *testing.T) {
	srv := newServer(&fakeWA{connected: false}, nil)
	rec := httptest.NewRecorder()
	srv.chatState(rec, httptest.NewRequest(http.MethodGet, "/chat-state?jid=a@s.whatsapp.net", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("kod = %d, want 503", rec.Code)
	}
}

// Yaz → OKU döngüsü: mute çağrıldıktan sonra okuma SUSTURULMUŞ göstermeli.
// Uçlar birbirinden bağımsız çalıştığı için bu, servisin yazdığını
// okuduğunu tek testte kanıtlar.
func TestMuteSonrasiOkumaSusturulmusGosterir(t *testing.T) {
	wa := &fakeWA{connected: true}
	srv := newServer(wa, nil)

	// Yaz.
	rec := istek(t, srv.mute, `{"jid":"a@s.whatsapp.net","muted":true,"for":"8h"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mute kod = %d", rec.Code)
	}
	// Yazma, sahte serviste okuma durumunu günceller (gerçek WhatsApp'ın
	// yaptığı gibi).
	wa.chatState = wabridge.ChatState{MutedUntil: time.Now().Add(8 * time.Hour)}

	// Oku.
	oku := httptest.NewRecorder()
	srv.chatState(oku, httptest.NewRequest(http.MethodGet, "/chat-state?jid=a@s.whatsapp.net", nil))

	if cevap(t, oku)["muted"] != true {
		t.Error("yazdıktan sonra okuma susturulmuş göstermedi — kanıt okunamıyor")
	}
}

// ── süre ayrıştırma ───────────────────────────────────────────────────────

func TestSureSusturBirimleri(t *testing.T) {
	vakalar := []struct {
		girilen string
		want    time.Duration
	}{
		{"8h", 8 * time.Hour},
		{"24h", 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"30d", 30 * 24 * time.Hour},
		{"1y", 365 * 24 * time.Hour},
		{" 8 H ", 8 * time.Hour},
		{"8SAAT", 8 * time.Hour},
		{"7GUN", 7 * 24 * time.Hour},
		{"30dk", 30 * time.Minute},
		{"", 0}, // kalıcı
		{"  ", 0},
	}

	for _, v := range vakalar {
		t.Run(v.girilen, func(t *testing.T) {
			got, err := sureSustur(v.girilen)
			if err != nil {
				t.Fatalf("hata: %v", err)
			}
			if got != v.want {
				t.Errorf("got %v, want %v", got, v.want)
			}
		})
	}
}

// Birim yanlış yorumlanırsa kullanıcı bildirimi yanlış zamanda geri alır ve
// hatanın nerede olduğunu bulamaz. Bu yüzden bozuk girdi 400 döner.
func TestSureSusturBozukGirdiHata(t *testing.T) {
	for _, bozuk := range []string{"abc", "8x", "h", "-5h", "0h", "8", "99999999999y"} {
		t.Run(bozuk, func(t *testing.T) {
			if _, err := sureSustur(bozuk); err == nil {
				t.Errorf("%q kabul edildi, hata bekleniyordu", bozuk)
			}
		})
	}
}

func TestMuteBozukSure400(t *testing.T) {
	wa := &fakeWA{connected: true}
	rec := istek(t, newServer(wa, nil).mute, `{"jid":"a@s.whatsapp.net","muted":true,"for":"bir hafta"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("kod = %d, want 400", rec.Code)
	}
	if wa.muteMuted {
		t.Error("bozuk süreye rağmen susturuldu")
	}
}

// ── yönlendirme ───────────────────────────────────────────────────────────

// Uçlar router'a bağlı olmalı: bağlı değilse kod derlenir ama çağrılamaz.
// Yöntem de sınanır: POST uçları GET'te, GET ucu POST'ta reddedilmeli.
func TestSohbetDurumuUclariRouterda(t *testing.T) {
	router := newServer(&fakeWA{connected: true}, nil).Router()

	postUclari := []string{"/mute", "/star"}
	for _, yol := range postUclari {
		t.Run(yol, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, yol, bytes.NewReader([]byte(`{}`)))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotFound {
				t.Errorf("%s router'da yok (404)", yol)
			}
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s kod = %d, want 400 (yol çalışıyor, gövde eksik)", yol, rec.Code)
			}
		})
	}

	t.Run("/chat-state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/chat-state?jid=a@s.whatsapp.net", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("kod = %d, want 200", rec.Code)
		}
	})
}

func TestSohbetDurumuYalnizDogruYontemiKabul(t *testing.T) {
	router := newServer(&fakeWA{connected: true}, nil).Router()

	vakalar := []struct {
		yol     string
		yontem  string
		wantKod int
	}{
		{"/mute", http.MethodGet, http.StatusMethodNotAllowed},
		{"/star", http.MethodGet, http.StatusMethodNotAllowed},
		{"/chat-state", http.MethodPost, http.StatusMethodNotAllowed},
	}

	for _, v := range vakalar {
		t.Run(v.yol+" "+v.yontem, func(t *testing.T) {
			req := httptest.NewRequest(v.yontem, v.yol, bytes.NewReader(nil))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != v.wantKod {
				t.Errorf("kod = %d, want %d", rec.Code, v.wantKod)
			}
		})
	}
}
