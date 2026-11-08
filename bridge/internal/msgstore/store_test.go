// whatsapp-bridge-project/bridge-go/internal/msgstore/store_test.go

package msgstore

import "testing"

func msg(id, from, sender string, ts int64) Message {
	return Message{ID: id, From: from, Sender: sender, Timestamp: ts}
}

func TestAddEvictsOldestPastCapacity(t *testing.T) {
	s := New(3)
	for i, id := range []string{"a", "b", "c", "d"} {
		s.Add(msg(id, "chat@s.whatsapp.net", "x@s.whatsapp.net", int64(i)))
	}

	if got := s.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}
	if _, ok := s.ByID("a"); ok {
		t.Error("oldest message 'a' should have been evicted")
	}
	if _, ok := s.ByID("d"); !ok {
		t.Error("newest message 'd' should be retained")
	}
}

func TestRecentReturnsNewestFirstAndFiltersByChat(t *testing.T) {
	s := New(10)
	s.Add(msg("1", "chatA@s.whatsapp.net", "peer@s.whatsapp.net", 1))
	s.Add(msg("2", "chatB@s.whatsapp.net", "peer@s.whatsapp.net", 2))
	s.Add(msg("3", "chatA@s.whatsapp.net", "peer@s.whatsapp.net", 3))

	got := s.Recent("chatA@s.whatsapp.net", 10)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "3" || got[1].ID != "1" {
		t.Errorf("order = %s,%s want 3,1", got[0].ID, got[1].ID)
	}
}

func TestRecentMatchesOnSenderJID(t *testing.T) {
	s := New(10)
	s.Add(msg("1", "group@g.us", "905551234567@s.whatsapp.net", 1))

	if got := s.Recent("905551234567@s.whatsapp.net", 10); len(got) != 1 {
		t.Fatalf("len = %d, want 1 (sender match)", len(got))
	}
}

func TestRecentHonoursCountLimit(t *testing.T) {
	s := New(10)
	for _, id := range []string{"1", "2", "3", "4"} {
		s.Add(msg(id, "chat@s.whatsapp.net", "peer@s.whatsapp.net", 1))
	}
	if got := s.Recent("chat@s.whatsapp.net", 2); len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestLastAnchorReturnsNewestMessage(t *testing.T) {
	s := New(10)
	s.Add(msg("old", "chat@s.whatsapp.net", "peer@s.whatsapp.net", 100))
	s.Add(msg("new", "chat@s.whatsapp.net", "peer@s.whatsapp.net", 200))

	anchor, ok := s.LastAnchor("chat@s.whatsapp.net")
	if !ok {
		t.Fatal("expected an anchor")
	}
	if anchor.ID != "new" || anchor.Timestamp != 200 {
		t.Errorf("anchor = %+v, want new@200", anchor)
	}
}

func TestLastAnchorMissingChat(t *testing.T) {
	s := New(10)
	s.Add(msg("1", "chatA@s.whatsapp.net", "peer@s.whatsapp.net", 1))

	if _, ok := s.LastAnchor("unknown@s.whatsapp.net"); ok {
		t.Error("expected no anchor for an unbuffered chat")
	}
}

func TestByIDUnknownReturnsFalse(t *testing.T) {
	s := New(10)
	if _, ok := s.ByID("nope"); ok {
		t.Error("expected ok=false for unknown ID")
	}
}

func TestNewRejectsNonPositiveCapacity(t *testing.T) {
	if got := New(0).capacity; got != DefaultCapacity {
		t.Errorf("capacity = %d, want %d", got, DefaultCapacity)
	}
	if got := New(-5).capacity; got != DefaultCapacity {
		t.Errorf("capacity = %d, want %d", got, DefaultCapacity)
	}
}
