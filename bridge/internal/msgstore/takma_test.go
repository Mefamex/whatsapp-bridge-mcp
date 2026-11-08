// whatsapp-bridge-project/bridge-go/internal/msgstore/takma_test.go
//
// TAKMA AD (alias) davranışı.
//
// Canlı ölçümün (2026-10-06) kaynağı: LID'li sohbette giden mesajlar LID
// görünümünde YOKTU. Çift gönderim denetimi bu yüzden hiç çalışmıyordu, ajan
// da kendi cevabını göremiyordu.
//
// Bu testler kırılmanın geri gelmesini yakalar.

package msgstore

import "testing"

// LID'li sohbet: giden mesajın From alanı ÇÖZÜLMÜŞ kimlik, takma ad ise
// istenen LID. İkisi de sorulduğunda da mesaj görünmeli.
func TestTakmaAdIkiKimligiBirdenKapsar(t *testing.T) {
	s := New(0)
	s.Add(Message{
		ID:       "M1",
		From:     "905550000001@s.whatsapp.net",
		Sender:   "905550000005:28@s.whatsapp.net",
		IsFromMe: true,
		Text:     "cevap",
		TakmaAd:  []string{"12345678901234@lid"},
	})

	for _, sorgu := range []string{
		"12345678901234@lid",          // istenen kimlik
		"905550000001@s.whatsapp.net", // çözülmüş kimlik
	} {
		got := s.Recent(sorgu, 10)
		if len(got) != 1 {
			t.Errorf("%s sorgusunda %d mesaj, want 1", sorgu, len(got))
		}
	}
}

// BU DÜZELTMENİN ASIL DEĞERİ: LID görünümünde fromMe mesaj bulunabilmeli.
// Çift gönderim denetimi yalnız fromMe mesajlara bakar; fromMe yoksa
// denetim sessizce her zaman "yeni" der.
func TestTakmaAdLidGorunumundeGidenMesajiGosterir(t *testing.T) {
	s := New(0)
	s.Add(Message{
		From:     "905550000001@s.whatsapp.net",
		Sender:   "905550000005:28@s.whatsapp.net",
		IsFromMe: true,
		Text:     "cevap",
		TakmaAd:  []string{"12345678901234@lid"},
	})

	got := s.Recent("12345678901234@lid", 10)
	if len(got) != 1 {
		t.Fatalf("LID görünümünde %d mesaj, want 1", len(got))
	}
	if !got[0].IsFromMe {
		t.Error("LID görünümündeki mesaj fromMe DEĞİL — çift gönderim denetimi yine çalışmaz")
	}
}

// Boş sorgu HİÇBİR mesajı getirmemeli. belongsTo'ya boş koruma konmadan
// önce boş jid'li istek, boş From/Sender alanlarıyla eşleşebilirdi.
func TestTakmaAdBosKimlikHicbirSeyiGetirmez(t *testing.T) {
	s := New(0)
	s.Add(Message{ID: "M1", From: "", Sender: "", TakmaAd: []string{""}})

	if got := s.Recent("", 10); len(got) != 0 {
		t.Errorf("boş kimlikle %d mesaj döndü, want 0", len(got))
	}
	if got := s.Recent("a@s.whatsapp.net", 10); len(got) != 0 {
		t.Errorf("takma adı boş olan mesaj yabancı sohbette göründü (%d)", len(got))
	}
}

// Takma ad YOKSA davranış eskisi gibi: başka kimlikteki mesaj görünmez.
// Düzeltme mevcut eşleşmeyi bozmamalı.
func TestTakmaAdsizMesajEskisiGibiDavranir(t *testing.T) {
	s := New(0)
	s.Add(Message{ID: "M1", From: "a@s.whatsapp.net", Text: "x"})

	if got := s.Recent("a@s.whatsapp.net", 10); len(got) != 1 {
		t.Errorf("kendi sohbetinde %d mesaj, want 1", len(got))
	}
	if got := s.Recent("b@s.whatsapp.net", 10); len(got) != 0 {
		t.Errorf("yabancı sohbette %d mesaj, want 0", len(got))
	}
	if got := s.Recent("a@lid", 10); len(got) != 0 {
		t.Errorf("takma adsız mesaj LID'de göründü (%d)", len(got))
	}
}

// Sıralama bozulmamalı: Recent en yeniden eskiye döner ve takma ad eşleşen
// mesajı doğru yere koyar.
func TestTakmaAdSiralamayiBozmaz(t *testing.T) {
	s := New(0)
	for _, id := range []string{"B", "C", "D"} {
		s.Add(Message{
			ID:       id,
			From:     "905550000001@s.whatsapp.net",
			IsFromMe: true,
			Text:     "m",
			TakmaAd:  []string{"12345678901234@lid"},
		})
	}

	got := s.Recent("12345678901234@lid", 10)
	if len(got) != 3 {
		t.Fatalf("%d mesaj, want 3", len(got))
	}
	if got[0].ID != "D" {
		t.Errorf("ilk mesaj %q, want D (en yeni önce)", got[0].ID)
	}
}

// SIFIR DEĞERLİ DEPO mesaj TUTMALI.
//
// Bu, test yazarken yaşanan tuzaktı: `&Store{}` kapasitesi 0'dır ve eski
// Add her ekleneni hemen siliyordu. Belirtiler yanıltıcıydı: /messages boş
// döner, "geçmiş alınamadı" yazar, çift gönderim denetimi çalışmaz, ajan
// geçmişi görmez. Hiçbiri kapasiteye işaret etmez.
func TestSifirDegerliDepoMesajTutar(t *testing.T) {
	s := &Store{}
	s.Add(Message{ID: "M1", From: "a@s.whatsapp.net", Text: "x"})

	if got := s.Recent("a@s.whatsapp.net", 10); len(got) != 1 {
		t.Fatalf("sıfır değerli depo %d mesaj tuttu, want 1", len(got))
	}
}
