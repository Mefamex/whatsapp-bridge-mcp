// whatsapp-bridge-project/bridge-go/internal/httpapi/handlers_test.go

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/msgstore"
	"whatsapp-bridge/internal/wabridge"
)

// fakeWA is an offline WhatsAppService for handler tests.
type fakeWA struct {
	loggedIn  bool
	connected bool

	sendErr      error
	mediaErr     error
	lastTo       string
	lastText     string
	historyCalls int

	// Sohbet durumu çağrıları: /mute ve /star.
	stateErr   error
	muteJID    string
	muteMuted  bool
	muteSure   time.Duration
	starJID    string
	starSender string
	starID     string
	starred    bool
	fromMe     bool
	chatState  wabridge.ChatState
}

func (f *fakeWA) IsLoggedIn() bool { return f.loggedIn }

func (f *fakeWA) Connected() bool { return f.connected }

func (f *fakeWA) SendText(_ context.Context, to, text string) (string, error) {
	if f.sendErr != nil {
		return "", f.sendErr
	}
	f.lastTo, f.lastText = to, text
	return "905551234567@s.whatsapp.net", nil
}

func (f *fakeWA) SendMedia(_ context.Context, _ wabridge.MediaRequest) (string, string, error) {
	if f.mediaErr != nil {
		return "", "", f.mediaErr
	}
	return "905551234567@s.whatsapp.net", "dp", nil
}

func (f *fakeWA) SendPoll(_ context.Context, req wabridge.PollRequest) (string, error) {
	if f.sendErr != nil {
		return "", f.sendErr
	}
	f.lastTo, f.lastText = req.To, req.Name
	return "905551234567@s.whatsapp.net", nil
}

func (f *fakeWA) SendCall(_ context.Context, to string, _ int) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.lastTo = to
	return nil
}

func (f *fakeWA) SendReaction(_ context.Context, to, targetID, emoji, _ string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.lastTo = to
	f.lastText = emoji
	return nil
}

func (f *fakeWA) RevokeMessage(_ context.Context, to, _ string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.lastTo = to
	return nil
}

func (f *fakeWA) Contacts(context.Context) ([]wabridge.Contact, error) {
	return []wabridge.Contact{{JID: "a@s.whatsapp.net", Name: "A"}}, nil
}

func (f *fakeWA) Groups(context.Context) ([]wabridge.Group, error) {
	return []wabridge.Group{{JID: "g@g.us", Name: "G", Participants: 3}}, nil
}

func (f *fakeWA) MarkRead(context.Context, string, []string) error { return nil }

func (f *fakeWA) MuteChat(_ context.Context, jid string, muted bool, d time.Duration) error {
	if f.stateErr != nil {
		return f.stateErr
	}
	f.muteJID, f.muteMuted, f.muteSure = jid, muted, d
	return nil
}

func (f *fakeWA) StarMessage(_ context.Context, jid, sender, messageID string, fromMe, starred bool) error {
	if f.stateErr != nil {
		return f.stateErr
	}
	f.starJID, f.starSender, f.starID = jid, sender, messageID
	f.fromMe, f.starred = fromMe, starred
	return nil
}

func (f *fakeWA) ChatSettings(_ context.Context, _ string) (wabridge.ChatState, error) {
	if f.stateErr != nil {
		return wabridge.ChatState{}, f.stateErr
	}
	return f.chatState, nil
}

func (f *fakeWA) RequestHistory(string, int) error {
	f.historyCalls++
	return errors.New("no anchor in test")
}

func (f *fakeWA) DownloadMedia(context.Context, *msgstore.Media) ([]byte, error) {
	return []byte("bytes"), nil
}

// newServer builds a Server backed by the fake service and an offline store.
func newServer(wa WhatsAppService, store *msgstore.Store) *Server {
	if store == nil {
		store = msgstore.New(10)
	}
	return New(wa, store, eventbus.New(), time.Now())
}

func TestHealthReportsSessionState(t *testing.T) {
	rec := httptest.NewRecorder()
	newServer(&fakeWA{loggedIn: true, connected: true}, nil).
		health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.Status != "ok" || !body.Connected || !body.Paired {
		t.Errorf("body = %+v, want ok + paired + connected", body)
	}
}

func TestSendTextRejectsMalformedJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)
	srv.sendText(rec, httptest.NewRequest(http.MethodPost, "/send", strings.NewReader("{not json")))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSendTextRequiresToAndText(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	for _, body := range []string{`{"to":"","text":"x"}`, `{"to":"905551234567","text":""}`} {
		rec := httptest.NewRecorder()
		srv.sendText(rec, httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}

func TestSendTextRequiresLiveConnection(t *testing.T) {
	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true}, nil)
	srv.sendText(rec, httptest.NewRequest(http.MethodPost, "/send",
		bytes.NewBufferString(`{"to":"905551234567","text":"hi"}`)))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSendTextSuccessPassesThroughPayload(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendText(rec, httptest.NewRequest(http.MethodPost, "/send",
		bytes.NewBufferString(`{"to":"905551234567","text":"hello"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if wa.lastText != "hello" {
		t.Errorf("sent text = %q, want hello", wa.lastText)
	}
	if !strings.Contains(rec.Body.String(), "905551234567@s.whatsapp.net") {
		t.Errorf("body = %s, want resolved JID", rec.Body)
	}
}

func TestSendTextSurfacesUpstreamError(t *testing.T) {
	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true, sendErr: errors.New("boom")}, nil)

	srv.sendText(rec, httptest.NewRequest(http.MethodPost, "/send",
		bytes.NewBufferString(`{"to":"905551234567","text":"hi"}`)))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestSendMediaRejectsInvalidBase64(t *testing.T) {
	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	srv.sendMedia(rec, httptest.NewRequest(http.MethodPost, "/send-media",
		bytes.NewBufferString(`{"to":"905551234567","media_base64":"!!!","mime_type":"image/jpeg"}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSendMediaRequiresFields(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	for _, body := range []string{
		`{"to":"","media_base64":"aGk=","mime_type":"image/jpeg"}`,
		`{"to":"905551234567","media_base64":"","mime_type":"image/jpeg"}`,
		`{"to":"905551234567","media_base64":"aGk=","mime_type":""}`,
	} {
		rec := httptest.NewRecorder()
		srv.sendMedia(rec, httptest.NewRequest(http.MethodPost, "/send-media", bytes.NewBufferString(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}

func TestContactsAndGroupsReturnCounts(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	rec := httptest.NewRecorder()
	srv.contacts(rec, httptest.NewRequest(http.MethodGet, "/contacts", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"count":1`) {
		t.Errorf("contacts: status=%d body=%s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	srv.groups(rec, httptest.NewRequest(http.MethodGet, "/groups", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"participants":3`) {
		t.Errorf("groups: status=%d body=%s", rec.Code, rec.Body)
	}
}

func TestDirectoryEndpointsRequireLiveConnection(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true}, nil)

	rec := httptest.NewRecorder()
	srv.contacts(rec, httptest.NewRequest(http.MethodGet, "/contacts", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("contacts status = %d, want 503", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.groups(rec, httptest.NewRequest(http.MethodGet, "/groups", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("groups status = %d, want 503", rec.Code)
	}
}

func TestMarkReadRequiresFields(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	for _, body := range []string{`{"jid":"","message_ids":["a"]}`, `{"jid":"x","message_ids":[]}`} {
		rec := httptest.NewRecorder()
		srv.markRead(rec, httptest.NewRequest(http.MethodPost, "/mark-read", bytes.NewBufferString(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}

func TestWebhookRegisterGetDelete(t *testing.T) {
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)

	rec := httptest.NewRecorder()
	srv.webhook(rec, httptest.NewRequest(http.MethodPost, "/webhook",
		bytes.NewBufferString(`{"url":"http://127.0.0.1:9000/hook"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("register status = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.webhook(rec, httptest.NewRequest(http.MethodGet, "/webhook", nil))
	if !strings.Contains(rec.Body.String(), "127.0.0.1:9000/hook") {
		t.Errorf("GET body = %s, want registered URL", rec.Body)
	}

	rec = httptest.NewRecorder()
	srv.webhook(rec, httptest.NewRequest(http.MethodDelete, "/webhook", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.webhook(rec, httptest.NewRequest(http.MethodGet, "/webhook", nil))
	if strings.Contains(rec.Body.String(), "127.0.0.1:9000/hook") {
		t.Error("URL should be cleared after DELETE")
	}
}

func TestWebhookPostRejectsEmptyURL(t *testing.T) {
	rec := httptest.NewRecorder()
	newServer(&fakeWA{loggedIn: true, connected: true}, nil).webhook(rec,
		httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{"url":""}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestMediaDownloadUnknownMessageIs404(t *testing.T) {
	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, nil)
	srv.routerVars(rec, "/media/nope", map[string]string{"messageId": "nope"}, srv.mediaDownload)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestMediaDownloadStreamsAttachment(t *testing.T) {
	store := msgstore.New(10)
	store.Add(msgstore.Message{ID: "MID1", Media: &msgstore.Media{MimeType: "image/jpeg"}})

	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, store)
	srv.routerVars(rec, "/media/MID1", map[string]string{"messageId": "MID1"}, srv.mediaDownload)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", got)
	}
	if rec.Body.String() != "bytes" {
		t.Errorf("body = %q, want 'bytes'", rec.Body)
	}
}

func TestMediaDownloadMessageWithoutMediaIs400(t *testing.T) {
	store := msgstore.New(10)
	store.Add(msgstore.Message{ID: "MID2"})

	rec := httptest.NewRecorder()
	srv := newServer(&fakeWA{loggedIn: true, connected: true}, store)
	srv.routerVars(rec, "/media/MID2", map[string]string{"messageId": "MID2"}, srv.mediaDownload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestMessagesHandlerServesBuffer(t *testing.T) {
	store := msgstore.New(10)
	store.Add(msgstore.Message{
		ID: "M1", From: "905551234567@s.whatsapp.net",
		Sender: "905551234567@s.whatsapp.net", Text: "selam",
	})

	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()
	srv := newServer(wa, store)
	srv.routerVars(rec, "/messages/905551234567?count=1", map[string]string{"jid": "905551234567"}, srv.messagesHandler)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "selam") {
		t.Errorf("body = %s, want the buffered message", rec.Body)
	}
	if wa.historyCalls != 0 {
		t.Error("history should not be requested when the buffer is sufficient")
	}
}

func TestMessagesHandlerTopsUpFromWhatsApp(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()
	srv := newServer(wa, msgstore.New(10))
	srv.routerVars(rec, "/messages/905551234567?count=5", map[string]string{"jid": "905551234567"}, srv.messagesHandler)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if wa.historyCalls != 1 {
		t.Errorf("history calls = %d, want 1", wa.historyCalls)
	}
}

func TestRouterRegistersEveryDocumentedRoute(t *testing.T) {
	router := newServer(&fakeWA{loggedIn: true, connected: true}, nil).Router()

	cases := []struct{ method, path string }{
		{http.MethodGet, "/health"},
		{http.MethodPost, "/send"},
		{http.MethodPost, "/send-media"},
		{http.MethodGet, "/contacts"},
		{http.MethodGet, "/groups"},
		{http.MethodPost, "/mark-read"},
		{http.MethodGet, "/messages/905551234567@s.whatsapp.net"},
		{http.MethodGet, "/media/ABC123"},
		{http.MethodGet, "/events"},
		{http.MethodPost, "/webhook"},
		{http.MethodDelete, "/webhook"},
	}

	for _, tc := range cases {
		var match mux.RouteMatch
		if !router.Match(httptest.NewRequest(tc.method, tc.path, nil), &match) {
			t.Errorf("no route for %s %s", tc.method, tc.path)
		}
	}
}

// routerVars runs a path-variable handler with mux vars preset, so handlers
// can be exercised directly without going through the router.
func (s *Server) routerVars(rec *httptest.ResponseRecorder, target string, vars map[string]string, h http.HandlerFunc) {
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, target, nil), vars)
	h(rec, req)
}

func TestSendPollRequiresValidPayload(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"bos ad", `{"to":"905551234567","name":"","options":["a","b"]}`},
		{"tek secenek", `{"to":"905551234567","name":"x","options":["a"]}`},
		{"seceneksiz", `{"to":"905551234567","name":"x","options":[]}`},
		{"hedefsiz", `{"name":"x","options":["a","b"]}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wa := &fakeWA{loggedIn: true, connected: true}
			rec := httptest.NewRecorder()

			newServer(wa, nil).sendPoll(rec, httptest.NewRequest(http.MethodPost, "/send-poll",
				bytes.NewBufferString(tc.body)))

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400: %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestSendPollSuccessPassesThroughPayload(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendPoll(rec, httptest.NewRequest(http.MethodPost, "/send-poll",
		bytes.NewBufferString(`{"to":"905551234567","name":"pıttık mı?","options":["pıttık1","pıttık2"]}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if wa.lastText != "pıttık mı?" {
		t.Errorf("sent name = %q, want anket adı", wa.lastText)
	}
}

func TestSendPollRequiresLiveConnection(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: false}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendPoll(rec, httptest.NewRequest(http.MethodPost, "/send-poll",
		bytes.NewBufferString(`{"to":"905551234567","name":"x","options":["a","b"]}`)))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSendCallRequiresTo(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendCall(rec, httptest.NewRequest(http.MethodPost, "/call",
		bytes.NewBufferString(`{"duration_sec":10}`)))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSendCallSuccessPassesThrough(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendCall(rec, httptest.NewRequest(http.MethodPost, "/call",
		bytes.NewBufferString(`{"to":"905551234567","duration_sec":10}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if wa.lastTo != "905551234567" {
		t.Errorf("sent to = %q, want hedef", wa.lastTo)
	}
}

func TestSendCallRequiresLiveConnection(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: false}
	rec := httptest.NewRecorder()

	newServer(wa, nil).sendCall(rec, httptest.NewRequest(http.MethodPost, "/call",
		bytes.NewBufferString(`{"to":"905551234567","duration_sec":10}`)))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestSendReaction(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	srv := newServer(wa, nil)

	// Hatalı gövde: to veya message_id eksik
	for _, body := range []string{`{"to":"","message_id":"1"}`, `{"to":"123","message_id":""}`} {
		rec := httptest.NewRecorder()
		srv.sendReaction(rec, httptest.NewRequest(http.MethodPost, "/reaction", bytes.NewBufferString(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for %s, got %d", body, rec.Code)
		}
	}

	// Başarılı reaksiyon
	rec := httptest.NewRecorder()
	srv.sendReaction(rec, httptest.NewRequest(http.MethodPost, "/reaction",
		bytes.NewBufferString(`{"to":"905551234567","message_id":"MSG123","emoji":"👍"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if wa.lastText != "👍" {
		t.Errorf("expected emoji 👍, got %s", wa.lastText)
	}
}

func TestRevokeMessage(t *testing.T) {
	wa := &fakeWA{loggedIn: true, connected: true}
	srv := newServer(wa, nil)

	// Hatalı gövde
	rec := httptest.NewRecorder()
	srv.revokeMessage(rec, httptest.NewRequest(http.MethodPost, "/revoke", bytes.NewBufferString(`{"to":""}`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Başarılı geri çekme
	rec = httptest.NewRecorder()
	srv.revokeMessage(rec, httptest.NewRequest(http.MethodPost, "/revoke",
		bytes.NewBufferString(`{"to":"905551234567","message_id":"MSG123"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
