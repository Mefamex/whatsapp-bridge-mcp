// whatsapp-bridge-project/bridge-go/internal/wabridge/poll_test.go

package wabridge

import (
	"strings"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// Telefon V3 anket gönderiyor (canlı kanıt 2026-10-06 02:47). V1 de
// aynı struct. İkisi de "poll" türüne düşmeli, seçenekler metinde olmalı.
func TestClassifyPollV1veV3(t *testing.T) {
	anket := &waE2E.PollCreationMessage{
		Name: proto.String("pıttık mı?"),
		Options: []*waE2E.PollCreationMessage_Option{
			{OptionName: proto.String("pıttık1")},
			{OptionName: proto.String("pıttık2")},
		},
	}

	cases := []struct {
		name string
		msg  *waE2E.Message
	}{
		{"v1", &waE2E.Message{PollCreationMessage: anket}},
		{"v3", &waE2E.Message{PollCreationMessageV3: anket}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tur, metin, _ := classify(tc.msg)
			if tur != "poll" {
				t.Fatalf("tür = %q, poll olmalı", tur)
			}
			for _, want := range []string{"pıttık mı?", "pıttık1", "pıttık2"} {
				if !strings.Contains(metin, want) {
					t.Errorf("metin %q içermiyor: %q", want, metin)
				}
			}
		})
	}
}

// V4 FutureProofMessage'dır: içeriği çözülemez, unsupported kalmalı.
func TestClassifyPollV4Cozulemez(t *testing.T) {
	tur, _, _ := classify(&waE2E.Message{
		PollCreationMessageV4: &waE2E.FutureProofMessage{},
	})
	if tur == "poll" {
		t.Errorf("V4 poll sayıldı, unsupported kalmalı (tür=%q)", tur)
	}
}

func TestHazirlaAnketDogrulama(t *testing.T) {
	cases := []struct {
		name    string
		req     PollRequest
		wantErr string
	}{
		{"bos hedef", PollRequest{Name: "x", Options: []string{"a", "b"}}, "to required"},
		{"bos ad", PollRequest{To: "9051", Options: []string{"a", "b"}}, "name required"},
		{"tek secenek", PollRequest{To: "9051", Name: "x", Options: []string{"a"}}, "at least 2"},
		{"bos secenekler sayilmaz", PollRequest{To: "9051", Name: "x", Options: []string{"a", " ", ""}}, "at least 2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := HazirlaAnket(tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("hata = %v, %q içermeli", err, tc.wantErr)
			}
		})
	}
}

func TestHazirlaAnket13SecenekReddedilir(t *testing.T) {
	secenekler := make([]string, 13)
	for i := range secenekler {
		secenekler[i] = strings.Repeat("s", 3) + string(rune('a'+i))
	}
	_, _, err := HazirlaAnket(PollRequest{To: "9051", Name: "x", Options: secenekler})
	if err == nil || !strings.Contains(err.Error(), "at most 12") {
		t.Errorf("13 seçenek kabul edildi: %v", err)
	}
}

func TestHazirlaAnketVarsayilanTekSecim(t *testing.T) {
	secenekler, sayilabilir, err := HazirlaAnket(PollRequest{
		To: "9051", Name: "x", Options: []string{"a", "b", "c"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sayilabilir != 1 {
		t.Errorf("sayilabilir = %d, varsayılan 1 olmalı", sayilabilir)
	}
	if len(secenekler) != 3 {
		t.Errorf("seçenek = %d, 3 olmalı", len(secenekler))
	}
}

func TestHazirlaAnketYineleneniAtar(t *testing.T) {
	secenekler, _, err := HazirlaAnket(PollRequest{
		To: "9051", Name: "x", Options: []string{"a", "a", "b", " "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(secenekler) != 2 {
		t.Errorf("yinelenen/boş atılmadı: %q", secenekler)
	}
}

func TestRenderPollBicimi(t *testing.T) {
	got := renderPoll("pıttık mı?", []string{"pıttık1", "pıttık2"})
	want := "📊 pıttık mı?\n• pıttık1\n• pıttık2"
	if got != want {
		t.Errorf("render = %q, want %q", got, want)
	}
}
