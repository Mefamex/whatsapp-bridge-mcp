// whatsapp-bridge-project/bridge-go/internal/msgstore/store.go

// Package msgstore keeps a bounded, in-memory ring buffer of the most recent
// messages so the HTTP API can serve history without touching WhatsApp.
//
// When opened with a path (Open), every Add is also appended to a JSONL log;
// a later Open replays the last capacity entries, so /messages survives a
// bridge restart without a WhatsApp history-sync round trip.
package msgstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

// DefaultCapacity is the number of most recent messages retained.
const DefaultCapacity = 1000

// Message is the canonical JSON shape of a chat message.
type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Direction string `json:"direction"` // "in" | "out"
	From      string `json:"from"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
	IsFromMe  bool   `json:"is_from_me"`
	IsGroup   bool   `json:"is_group"`
	PushName  string `json:"push_name,omitempty"`
	Media     *Media `json:"media,omitempty"`
	// Quoted holds the referenced message if this was a reply (swipe right).
	Quoted *QuotedMessage `json:"quoted,omitempty"`
	// Edited is true when the sender later edited the message and this
	// record carries the newest text. EditedAt is the edit's timestamp.
	Edited   bool  `json:"edited,omitempty"`
	EditedAt int64 `json:"edited_at,omitempty"`

	// TakmaAd, bu mesajın AYRICA ait olduğu kimlikler.
	//
	// NEDEN VAR (2026-10-06 canlı ölçümü): bir sohbet iki kimlikle
	// bilinebilir — gelen mesaj `47575939964948@lid` ile gelir, giden
	// mesaj ise çözülmüş `905…@s.whatsapp.net` olarak kaydedilir. Tam
	// eşleşme arayan store, LID'li görünümde giden mesajları GÖRMEDİ.
	//
	// Bunun bedeli sessizdi: çift gönderim denetimi hiç tetiklenmiyordu ve
	// ajan kendi önceki cevabını göremiyordu. Kimlik eşleştirme yapılmadığı
	// için hata yoktu — sadece mesaj yoktu.
	//
	// Bu alan JSON'a da çıkar: dışarıdaki istemci de (tetikleyici) hangi
	// kimliklerin aynı sohbete gittiğini görebilir.
	TakmaAd []string `json:"aliases,omitempty"`
}

// QuotedMessage describes the referenced/replied message in a quote context.
type QuotedMessage struct {
	ID     string `json:"id,omitempty"`
	Sender string `json:"sender,omitempty"`
	Type   string `json:"type,omitempty"`
	Text   string `json:"text,omitempty"`
}

// Media holds the encrypted-media metadata needed to download an attachment.
type Media struct {
	URL           string `json:"url"`
	DirectPath    string `json:"direct_path"`
	MediaKey      string `json:"media_key"`
	FileEncSHA256 string `json:"file_enc_sha256"`
	FileSHA256    string `json:"file_sha256"`
	FileLength    uint64 `json:"file_length"`
	MimeType      string `json:"mime_type"`
}

// Anchor identifies a known message, used as the starting point for
// WhatsApp history-sync requests.
type Anchor struct {
	ID        string
	Timestamp int64
	IsFromMe  bool
}

// Store is a concurrency-safe bounded message buffer.
type Store struct {
	mu       sync.RWMutex
	messages []Message
	capacity int

	// Path-backed persistence (zero value = memory only).
	path             string
	file             *os.File
	addsSinceCompact int
}

// New creates a store retaining at most capacity messages. It is
// memory-only; use Open to attach a JSONL log that survives restarts.
func New(capacity int) *Store {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Store{capacity: capacity}
}

// Open creates a store backed by the JSONL file at path. If the file exists,
// the last capacity entries are replayed into memory; corrupt tail lines are
// skipped. Errors (unreadable directory, unwritable file) are returned so the
// caller can fall back to a memory-only store knowingly.
func Open(capacity int, path string) (*Store, error) {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}

	s := &Store{capacity: capacity, path: path}

	if data, err := os.ReadFile(path); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
		ok, corrupt := 0, 0
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var m Message
			if err := json.Unmarshal(line, &m); err != nil {
				corrupt++
				continue
			}
			ok++
			s.messages = append(s.messages, m)
		}
		if ok > capacity {
			s.messages = s.messages[len(s.messages)-capacity:]
		}
		if corrupt > 0 {
			log.Printf("msgstore: %d bozuk satır atlandı (%s)", corrupt, path)
		}
		if len(s.messages) > 0 {
			log.Printf("msgstore: %d mesaj %s'ten yüklendi", len(s.messages), path)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	s.file = f
	return s, nil
}

// Close flushes and closes the backing log file. Safe on memory-only stores.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Sync()
	if cerr := s.file.Close(); err == nil {
		err = cerr
	}
	s.file = nil
	return err
}

// Add appends a message, evicting the oldest entries past capacity.
//
// KAPASİTE 0 KORUMASI: sıfır kapasite, "her şeyi hemen sil" demektir.
// `&Store{}` ile kurulan bir depo sessizce HİÇBİR mesajı tutmaz ve hata da
// vermez — /messages boş döner, geçmiş alınamaz, çift gönderim denetimi
// çalışmaz, ajan geçmişi görmez. Hepsi aynı belirti, hiçbiri nedenini
// söylemez. Bu tuzak test yazarken bir kez yakalandı; burada kapatılıyor.
func (s *Store) Add(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	kapasite := s.capacity
	if kapasite <= 0 {
		kapasite = DefaultCapacity
	}

	s.messages = append(s.messages, msg)
	s.appendLocked(msg)

	if len(s.messages) > kapasite {
		s.messages = s.messages[len(s.messages)-kapasite:]
		s.addsSinceCompact++
	}

	// Dosya, belleği aşan her eklemede büyür; kapasite kadar add birikince
	// log, bellekteki son kapasiteyle yeniden yazılır (tmp + rename =
	// atomik). Böylece dosya kabaca 2×kapasite satırı aşmaz ve restart
	// replaysi ucuz kalır.
	if s.file != nil && s.addsSinceCompact >= kapasite {
		s.addsSinceCompact = 0
		s.compactLocked()
	}
}

// appendLocked appends one JSON line to the backing log. Caller holds s.mu.
func (s *Store) appendLocked(msg Message) {
	if s.file == nil {
		return
	}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("msgstore: marshall hatası (%v) — loga yazılmadı", err)
		return
	}
	if _, err := s.file.Write(append(b, '\n')); err != nil {
		log.Printf("msgstore: log yazma hatası (%v) — bellek devam", err)
	}
}

// compactLocked rewrites the log with the buffered messages. Caller holds mu.
// The file is swapped via a temp file + rename so a crash never leaves a
// truncated log; the old descriptor is replaced so appends keep working.
func (s *Store) compactLocked() {
	tmp := s.path + ".tmp"
	tf, err := os.Create(tmp)
	if err != nil {
		log.Printf("msgstore: compact açamadı (%v)", err)
		return
	}
	tf.Chmod(0o600)
	enc := json.NewEncoder(tf)
	for i := range s.messages {
		if err := enc.Encode(&s.messages[i]); err != nil {
			log.Printf("msgstore: compact encode hatası (%v)", err)
			tf.Close()
			os.Remove(tmp)
			return
		}
	}
	if err := tf.Sync(); err != nil {
		log.Printf("msgstore: compact sync hatası (%v)", err)
		tf.Close()
		os.Remove(tmp)
		return
	}
	if err := tf.Close(); err != nil {
		log.Printf("msgstore: compact close hatası (%v)", err)
		os.Remove(tmp)
		return
	}
	if err := os.Rename(tmp, s.path); err != nil {
		log.Printf("msgstore: compact rename hatası (%v)", err)
		os.Remove(tmp)
		return
	}
	nf, err := os.OpenFile(s.path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		log.Printf("msgstore: compact yeniden açamadı (%v) — log askıda", err)
		s.file = nil
		return
	}
	s.file = nf
}

// Len returns the number of buffered messages.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}

// Recent returns up to count messages of the given chat, newest first.
func (s *Store) Recent(jid string, count int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Message, 0, count)
	for i := len(s.messages) - 1; i >= 0 && len(result) < count; i-- {
		if s.belongsTo(&s.messages[i], jid) {
			result = append(result, s.messages[i])
		}
	}
	return result
}

// Search returns up to limit messages whose text contains the query
// (case-insensitive, trimmed), newest first. Query boşsa boş sonuç döner.
func (s *Store) Search(query string, limit int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	if limit <= 0 {
		limit = DefaultCapacity
	}

	result := make([]Message, 0, limit)
	for i := len(s.messages) - 1; i >= 0 && len(result) < limit; i-- {
		m := &s.messages[i]
		if strings.Contains(strings.ToLower(m.Text), q) {
			result = append(result, *m)
		}
	}
	return result
}

// ByID finds a message by its WhatsApp message ID.
func (s *Store) ByID(id string) (Message, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.messages {
		if s.messages[i].ID == id {
			return s.messages[i], true
		}
	}
	return Message{}, false
}

// LastAnchor returns the newest buffered message of a chat, for history sync.
func (s *Store) LastAnchor(jid string) (Anchor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.messages) - 1; i >= 0; i-- {
		if s.belongsTo(&s.messages[i], jid) {
			m := s.messages[i]
			return Anchor{ID: m.ID, Timestamp: m.Timestamp, IsFromMe: m.IsFromMe}, true
		}
	}
	return Anchor{}, false
}

// ApplyEdit replaces the text of the message with the given ID and marks it
// edited. It returns false when the original is not buffered — the edit then
// has nothing to attach to, and the caller should store it as its own record
// instead of silently dropping it.
func (s *Store) ApplyEdit(id, text, kind string, media *Media, editedAt int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.messages {
		if s.messages[i].ID == id {
			s.messages[i].Text = text
			s.messages[i].Type = kind
			s.messages[i].Media = media
			s.messages[i].Edited = true
			s.messages[i].EditedAt = editedAt
			return true
		}
	}
	return false
}

// belongsTo reports whether a message concerns the given chat JID.
// belongsTo, mesajın bu sohbete ait olup olmadığını söyler.
//
// Üç kimlik denenir: From, Sender ve TakmaAd. TakmaAd olmadan LID'li
// sohbetlerde giden mesajlar görünmez — bkz. Message.TakmaAd yorumu.
func (s *Store) belongsTo(m *Message, jid string) bool {
	if jid == "" {
		return false
	}
	if m.From == jid || m.Sender == jid {
		return true
	}
	for _, ad := range m.TakmaAd {
		if ad == jid {
			return true
		}
	}
	return false
}
