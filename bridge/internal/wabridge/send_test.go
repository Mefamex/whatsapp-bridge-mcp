// whatsapp-bridge-project/bridge-go/internal/wabridge/send_test.go

package wabridge

import (
	"context"
	"errors"
	"testing"

	"go.mau.fi/whatsmeow/types"
)

// fakeLIDs is an in-memory stand-in for store.LIDStore.
type fakeLIDs struct {
	byLID map[string]types.JID
	err   error
}

func (f fakeLIDs) GetPNForLID(_ context.Context, lid types.JID) (types.JID, error) {
	if f.err != nil {
		return types.EmptyJID, f.err
	}
	return f.byLID[lid.String()], nil
}

// LID sohbetlerine ham LID ile cevap verildiğinde mesaj kullanıcının
// gördüğü konuşmaya düşmüyordu. Bu test, LID'in telefon numarasına
// çözülmesini zorunlu kılıyor.
func TestResolveTargetMapsLIDToPhoneNumber(t *testing.T) {
	lids := fakeLIDs{byLID: map[string]types.JID{
		"12345678901234@lid": types.NewJID("905550000001", types.DefaultUserServer),
	}}

	got, err := resolveTarget(context.Background(), lids, "12345678901234@lid")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}

	if got.User != "905550000001" || got.Server != types.DefaultUserServer {
		t.Errorf("got %s, want 905550000001@s.whatsapp.net", got)
	}
}

// Telefon numarası zaten doğruysa hiçbir çözümleme yapılmamalı; gereksiz
// bir depo sorgusu LID'in telefon numarasına çevrilmesini bozabilir.
func TestResolveTargetLeavesPhoneNumbersAlone(t *testing.T) {
	// Hata veren depo, "sorgu yapılmadı" kanıtıdır.
	lids := fakeLIDs{err: errors.New("depo hiç sorgulanmamalıydı")}

	got, err := resolveTarget(context.Background(), lids, "905550000001@s.whatsapp.net")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got.String() != "905550000001@s.whatsapp.net" {
		t.Errorf("got %s, want unchanged", got)
	}
}

// Grup kimlikleri de dokunulmadan geçmeli.
func TestResolveTargetLeavesGroupsAlone(t *testing.T) {
	lids := fakeLIDs{err: errors.New("depo hiç sorgulanmamalıydı")}

	got, err := resolveTarget(context.Background(), lids, "12345-67890@g.us")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got.String() != "12345-67890@g.us" {
		t.Errorf("got %s, want unchanged", got)
	}
}

// Eşleme yoksa gönderim ENGELLENMEZ; ham LID'e düşülür. Aksi hâlde tek bir
// eksik eşleme tüm yanıtları sustururdu.
func TestResolveTargetFallsBackToRawLIDWhenUnmapped(t *testing.T) {
	lids := fakeLIDs{byLID: map[string]types.JID{}}

	got, err := resolveTarget(context.Background(), lids, "99999@lid")
	if err != nil {
		t.Fatalf("resolveTarget should not fail on an unmapped LID: %v", err)
	}
	if got.String() != "99999@lid" {
		t.Errorf("got %s, want the raw LID", got)
	}
}

func TestResolveTargetFallsBackWhenTheStoreErrors(t *testing.T) {
	lids := fakeLIDs{err: errors.New("db kapalı")}

	got, err := resolveTarget(context.Background(), lids, "99999@lid")
	if err != nil {
		t.Fatalf("resolveTarget should not fail on a store error: %v", err)
	}
	if got.String() != "99999@lid" {
		t.Errorf("got %s, want the raw LID", got)
	}
}

// Depo boş bir JID döndürürse bu "bulunamadı" demektir; çözülmüş sayılmamalı.
func TestResolveTargetRejectsAnEmptyMapping(t *testing.T) {
	lids := fakeLIDs{byLID: map[string]types.JID{
		"99999@lid": types.EmptyJID,
	}}

	got, err := resolveTarget(context.Background(), lids, "99999@lid")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got.String() != "99999@lid" {
		t.Errorf("got %s, want the raw LID for an empty mapping", got)
	}
}

// @s.whatsapp.net yerine yanlış sunucu dönen bir eşleme kabul edilmemeli.
func TestResolveTargetRejectsANonPhoneMapping(t *testing.T) {
	lids := fakeLIDs{byLID: map[string]types.JID{
		"99999@lid": types.NewJID("905550000001", "example.com"),
	}}

	got, err := resolveTarget(context.Background(), lids, "99999@lid")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got.String() != "99999@lid" {
		t.Errorf("got %s, want the raw LID when the mapping is not a phone JID", got)
	}
}
