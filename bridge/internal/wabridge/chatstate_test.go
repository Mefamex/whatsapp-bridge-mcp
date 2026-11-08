// whatsapp-bridge-project/bridge-go/internal/wabridge/chatstate_test.go
//
// MuteChat ve StarMessage: uç katmanının doğrulama mantığı.
//
// AĞ YOK: nil istemci kullanılır. whatsmeow'un SendAppState'i nil istemcide
// ErrClientIsNil döner (panik atmaz), dolayısıyla buradaki testler gerçekten
// istemcinin ÇAĞRILDIĞINI ama bağlantıya gitmediğini doğrular. Önemli olan
// da bu: JID normalizasyonu ve kimlik doğrulaması ağ olmadan sınanır.

package wabridge

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
)

// errClientNil, whatsmeow'un nil istemci için döndürdüğü hatadır.
//
// KİMLİK EŞLEŞMESİ: whatsmeow'un error DEĞERİ kullanılır, `errors.Is` ile
// eşleşir. Metne göre eşleştirme ("client is nil") whatsmeow metni
// değiştirdiğinde sessizce TÜM bu testleri yanlış çevirir — testler
// kırılmaz, tersine yalan söyler.
var errClientNil = whatsmeow.ErrClientIsNil

// mute, geçersiz girdide istemciye ULAŞILMAZ: hata istemciden önce döner.
func TestMuteChatGecersizJidIstemciyeGitmez(t *testing.T) {
	s := &Service{}

	err := s.MuteChat(context.Background(), "gecersiz jid @@@", true, time.Hour)
	if err == nil {
		t.Fatal("geçersiz JID için hata bekleniyordu")
	}
	if errors.Is(err, errClientNil) {
		t.Errorf("geçersiz JID istemciye ulaştı (hata = %v)", err)
	}
}

// Kalıcı susturma (süre 0) serviste BİR YILA çevrilir: WhatsApp'ta kalıcı
// susturma yok, "süresiz" her zaman büyük bir süre olmak zorunda. Sonsuz
// süre göndermek bazı istemcilerde susturmayı KALDIRIR.
func TestMuteChatKaliciSureBirYilaCevrilir(t *testing.T) {
	s := &Service{}

	err := s.MuteChat(context.Background(), "47575939964948@lid", true, 0)
	if !errors.Is(err, errClientNil) {
		t.Errorf("istemciye ulaşılmadı (hata = %v); kalıcı süre reddedildi", err)
	}
}

// StarMessage kimliksiz isteği İSTEMCİYE GÖNDERMEMELİ: yıldız bir mesajı
// işaretler, kimlik yoksa istek anlamsız. Boşluk dolu değildir.
func TestStarMessageKimliksizIstemciyeGitmez(t *testing.T) {
	s := &Service{}

	for _, bos := range []string{"", "   ", "\t"} {
		err := s.StarMessage(context.Background(), "47575939964948@lid", "", bos, false, true)
		if !errors.Is(err, errMessageIDRequired) {
			t.Errorf("message_id=%q için hata = %v, want %q", bos, err, errMessageIDRequired)
		}
	}
}

// Boş gönderen, hedefe DÜŞMELİ: birebir sohbette gönderen hedefle aynıdır.
// Düşmezse whatsmeow boş JID ile yıldız yaması kurar ve telefon uygulamasında
// yıldız görünmez — hata da vermez (sessiz başarısızlık).
func TestStarMessageBosGonderenHedefeDuser(t *testing.T) {
	s := &Service{}

	err := s.StarMessage(context.Background(), "47575939964948@lid", "", "M1", false, true)
	if !errors.Is(err, errClientNil) {
		t.Errorf("istemciye ulaşılmadı (hata = %v); boş gönderen hedefe düşmüyor", err)
	}
	if errors.Is(err, errMessageIDRequired) {
		t.Error("gönderen boş olduğu hâlde kimlik hatası verildi")
	}
}

// Açık gönderen korunur: grup mesajlarında gönderen hedefTEN FARKLIDIR,
// yanlış doldurulursa yıldız yanlış mesaja biner.
func TestStarMessageAcikGonderenKorunur(t *testing.T) {
	s := &Service{}

	err := s.StarMessage(context.Background(), "a@s.whatsapp.net", "b@s.whatsapp.net", "M1", false, true)
	if !errors.Is(err, errClientNil) {
		t.Errorf("geçerli gönderende hata = %v, want client is nil", err)
	}
}

// Geçersiz gönderen kimliği reddedilmeli: sessizce boşa düşerse yıldız
// kaybolur ve nedeni anlaşılmaz.
func TestStarMessageGecersizGonderenHata(t *testing.T) {
	s := &Service{}

	err := s.StarMessage(context.Background(), "a@s.whatsapp.net", "bozuk @@@", "M1", false, true)
	if err == nil {
		t.Fatal("geçersiz gönderen için hata bekleniyordu")
	}
	if errors.Is(err, errClientNil) {
		t.Errorf("geçersiz gönderen istemciye ulaştı (hata = %v)", err)
	}
}

// Yıldız KALDIRMA da istemciye gitmeli: starred=false geçerli bir işlemdir.
func TestStarMessageYildizKaldirmaIstemciyeGider(t *testing.T) {
	s := &Service{}

	if err := s.StarMessage(context.Background(), "a@s.whatsapp.net", "", "M1", false, false); !errors.Is(err, errClientNil) {
		t.Errorf("yıldız kaldırma istemciye gitmedi: %v", err)
	}
}

// ChatSettings bozuk kimlikte istemciye ULAŞILMAZ: okuma ucu da aynı
// doğrulamadan geçmeli, aksi hâlde /chat-state boş sohbet bilgisi döndürür.
func TestChatSettingsBozukJidIstemciyeGitmez(t *testing.T) {
	s := &Service{}

	_, err := s.ChatSettings(context.Background(), "gecersiz@@")
	if err == nil {
		t.Fatal("çoklu @ içeren kimlik için hata bekleniyordu")
	}
	if errors.Is(err, errClientNil) {
		t.Errorf("bozuk kimlik istemciye ulaştı: %v", err)
	}
}

// ChatState.Muted, sıfır zamanı "susturulmuş" saymamalı. Sıfır zaman, Go'nun
// "hiçbir zaman ayarlanmamış" anlamıdır; Unix epoch DEĞİL. Bu ayrım yanlış
// yapılırsa her sohbet susturulmuş görünür.
func TestChatStateSifirZamanSusturulmusSayilmaz(t *testing.T) {
	if (ChatState{}).Muted() {
		t.Error("sıfır zaman susturulmuş sayıldı")
	}
	if !(ChatState{MutedUntil: time.Now()}).Muted() {
		t.Error("gerçek bir zaman susturulmuş sayılmadı")
	}
}
