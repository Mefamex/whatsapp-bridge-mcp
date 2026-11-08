// whatsapp-bridge-project/bridge-go/internal/msgstore/persistence_test.go

package msgstore

import (
	"os"
	"path/filepath"
	"testing"
)

// yardimciMesaj builds a minimal message with an id and timestamp.
func yardimciMesaj(id string, ts int64) Message {
	return Message{ID: id, Type: "text", Direction: "in", From: "905551234567@s.whatsapp.net", Timestamp: ts, Text: id}
}

func TestOpen_MissingFileStartsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "yok.jsonl")

	s, err := Open(10, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if s.Len() != 0 {
		t.Errorf("Len() = %d, want 0", s.Len())
	}
}

func TestOpen_ReplaysPersistedMessages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")

	s, err := Open(10, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for i := 0; i < 3; i++ {
		s.Add(yardimciMesaj(string(rune('a'+i)), int64(i)))
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Yeni süreç taklidi: dosyayı yeniden aç, mesajlar geri gelmeli.
	s2, err := Open(10, path)
	if err != nil {
		t.Fatalf("Open(2): %v", err)
	}
	defer s2.Close()

	if s2.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", s2.Len())
	}
	recent := s2.Recent("905551234567@s.whatsapp.net", 3)
	if len(recent) != 3 {
		t.Fatalf("Recent len = %d, want 3", len(recent))
	}
	// Recent yeni→eski döner; en yeni "c" olmalı.
	if recent[0].ID != "c" || recent[2].ID != "a" {
		t.Errorf("sıra bozuk: %v %v %v", recent[0].ID, recent[1].ID, recent[2].ID)
	}
}

func TestOpen_OnlyLastCapacityReplayed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")

	s, _ := Open(5, path)
	for i := 0; i < 8; i++ {
		s.Add(yardimciMesaj(string(rune('a'+i)), int64(i)))
	}
	s.Close()

	s2, _ := Open(5, path)
	defer s2.Close()

	if s2.Len() != 5 {
		t.Fatalf("Len() = %d, want 5 (son kapasite)", s2.Len())
	}
	// Bellekte de son 5 kaldı (eskiler "a","b","c" yüklendi mi? hayır — baştan
	// appende devam eden dosyada 8 satır var; son 5 alınır: d,e,f,g,h).
	recent := s2.Recent("905551234567@s.whatsapp.net", 5)
	if recent[4].ID != "d" {
		t.Errorf("en eski replay = %q, want 'd'", recent[4].ID)
	}
	if recent[0].ID != "h" {
		t.Errorf("en yeni replay = %q, want 'h'", recent[0].ID)
	}
}

func TestOpen_SkipsCorruptLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")

	// Elle bir dosya yaz: bozuk satır + geçerli satır + yarıda kesik satır.
	content := "bu bir json degil\n" + `{"id":"iyi","type":"text","direction":"in","from":"905551234567@s.whatsapp.net","timestamp":1,"text":"saglam"}` + "\nyarida-kesik-son-satir"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	s, err := Open(10, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if s.Len() != 1 {
		t.Fatalf("Len() = %d, want 1 (yalnızca geçerli satır)", s.Len())
	}
	if msg, ok := s.ByID("iyi"); !ok || msg.Text != "saglam" {
		t.Errorf("geçerli mesaj geri gelmedi: %+v ok=%v", msg, ok)
	}
}

func TestNew_MemoryOnlyWritesNothingAndCloses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bos.jsonl")

	s := New(10)
	s.Add(yardimciMesaj("x", 1))
	if err := s.Close(); err != nil {
		t.Fatalf("Close memory-only: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("New() hiç dosya oluşturmamalı, dosya var: %v", err)
	}
}

func TestAdd_AppendsToFileImmediately(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")

	s, _ := Open(10, path)
	s.Add(yardimciMesaj("canli", 42))
	s.Add(yardimciMesaj("canli2", 43))

	// Dosya içeriği satır satır kontrol: 2 satır JSON olmalı.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("dosyada %d satır, want 2", lines)
	}
	s.Close()
}

func TestCompact_KeepsFileBoundedAndAtomic(t *testing.T) {
	// Kapasite 4, 10 mesaj ekle: trim her eklentide olur, compact 4 add'de bir.
	// Dosya boyutu ~2×kapasite mertebesinde kalmalı ve tekrar açılınca son 4
	// mesaj gelmeli.
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")

	s, _ := Open(4, path)
	for i := 0; i < 10; i++ {
		s.Add(yardimciMesaj(string(rune('a'+i)), int64(i)))
	}
	s.Close()

	s2, _ := Open(4, path)
	defer s2.Close()
	if s2.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", s2.Len())
	}
	recent := s2.Recent("905551234567@s.whatsapp.net", 4)
	if recent[0].ID != "j" || recent[3].ID != "g" {
		t.Errorf("compact sonrası yanlış mesajlar: %v %v %v %v", recent[0].ID, recent[1].ID, recent[2].ID, recent[3].ID)
	}
}
