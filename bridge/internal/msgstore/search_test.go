// whatsapp-bridge-project/bridge-go/internal/msgstore/search_test.go

package msgstore

import "testing"

func TestSearch_FindsCaseInsensitiveText(t *testing.T) {
	s := New(10)
	s.Add(Message{ID: "1", Text: "Hello Dünya", Direction: "in", From: "44556677@s.whatsapp.net"})
	s.Add(Message{ID: "2", Text: "merhaba dünya", Direction: "in", From: "11223344@s.whatsapp.net"})
	s.Add(Message{ID: "3", Text: "başka konu", Direction: "in", From: "11223344@s.whatsapp.net"})

	got := s.Search("DÜNYA", 10)
	if len(got) != 2 {
		t.Fatalf("Search(\"DÜNYA\") = %d sonuç, want 2", len(got))
	}
	// Yeni→eski: "2" sonradan eklendi → önce gelmeli.
	if got[0].ID != "2" || got[1].ID != "1" {
		t.Errorf("sıra bozuk: %s, %s", got[0].ID, got[1].ID)
	}
}

func TestSearch_EmptyQueryReturnsNil(t *testing.T) {
	s := New(4)
	s.Add(Message{ID: "1", Text: "x", Direction: "in", From: "a@s.whatsapp.net"})
	if got := s.Search("   ", 10); got != nil {
		t.Errorf("boş sorgu nil dönmeli, %v geldi", got)
	}
}

func TestSearch_NoMatchReturnsEmpty(t *testing.T) {
	s := New(4)
	s.Add(Message{ID: "1", Text: "elma", Direction: "in", From: "a@s.whatsapp.net"})
	if got := s.Search("muz", 10); len(got) != 0 {
		t.Errorf("eşleşme yokken %d sonuç", len(got))
	}
}

func TestSearch_HonoursLimit(t *testing.T) {
	s := New(10)
	for i := 0; i < 5; i++ {
		s.Add(Message{ID: string(rune('a' + i)), Text: "ortak", Direction: "in", From: "a@s.whatsapp.net"})
	}
	got := s.Search("ortak", 2)
	if len(got) != 2 {
		t.Fatalf("limit 2 → %d sonuç", len(got))
	}
	if got[0].ID != "e" {
		t.Errorf("en yeni önce gelmeli, %q geldi", got[0].ID)
	}
}

func TestSearch_WhitespaceQueryTrimmed(t *testing.T) {
	s := New(4)
	s.Add(Message{ID: "1", Text: "selam dünya", Direction: "in", From: "a@s.whatsapp.net"})
	if got := s.Search("  dünya  ", 10); len(got) != 1 {
		t.Errorf("trim edilmeli, %d sonuç", len(got))
	}
}
