// whatsapp-bridge-project/bridge-go/internal/wabridge/takma_test.go
//
// takmaAdlar: istenen kimlik ne zaman mesaja eklenir.
//
// Bu kural yanlışsa iki hata birden çıkar:
//   · Gereksiz takma ad → mesaj, sorgulanmayan sohbetlerde de görünür.
//   · Eksik takma ad → LID'li sohbetlerde giden mesaj kaybolur (canlı
//     ölçümde bu olan: çift gönderim denetimi ve ajan geçmişi ölüydü).

package wabridge

import (
	"testing"

	"go.mau.fi/whatsmeow/types"
)

func TestTakmaAdYoksaUretilmez(t *testing.T) {
	target := types.JID{User: "905550000001", Server: "s.whatsapp.net"}

	vakalar := []struct {
		ad      string
		istendi string
	}{
		{"isteyen ile çözülen aynı", target.String()},
		{"isteyen boş", ""},
		{"isteyen yalnız boşluk", "   "},
	}

	for _, v := range vakalar {
		t.Run(v.ad, func(t *testing.T) {
			if got := takmaAdlar(v.istendi, target); got != nil {
				t.Errorf("takma ad üretildi: %v (gereksiz — mesaj başka sohbetlerde de görünür)", got)
			}
		})
	}
}

func TestTakmaAdCozulmeyenKimlikIcinUretilir(t *testing.T) {
	target := types.JID{User: "905550000001", Server: "s.whatsapp.net"}
	istendi := "12345678901234@lid"

	got := takmaAdlar(istendi, target)
	if len(got) != 1 || got[0] != istendi {
		t.Fatalf("takma ad = %v, want [%s]", got, istendi)
	}
}

// Boşluklu istek KIRPILMALI: " 475…@lid " ile "475…@lid" aynı sohbettir.
// Kırpılmazsa mesaj " boşluklu" kimlikle kaydedilir ve normal sorgu onu
// bulamaz — yani takma ad yine işe yaramaz.
func TestTakmaAdKirpilir(t *testing.T) {
	target := types.JID{User: "905550000001", Server: "s.whatsapp.net"}

	got := takmaAdlar("  12345678901234@lid  ", target)
	if len(got) != 1 {
		t.Fatalf("takma ad = %v, want 1 öğe", got)
	}
	if got[0] != "12345678901234@lid" {
		t.Errorf("takma ad = %q, kırpılmamış", got[0])
	}
}
