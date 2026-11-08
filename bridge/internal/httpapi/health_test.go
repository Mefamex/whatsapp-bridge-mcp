// whatsapp-bridge-project/bridge-go/internal/httpapi/health_test.go

package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/msgstore"
)

// pairedButDisconnected has a stored session but no live socket.
func pairedButDisconnected() *Server {
	return New(&fakeWA{loggedIn: true, connected: false},
		msgstore.New(10), eventbus.New(), time.Now())
}

func TestHealthDistinguishesPairedFromConnected(t *testing.T) {
	rec := httptest.NewRecorder()
	pairedButDisconnected().health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"paired":true`) {
		t.Errorf("body = %s, want paired:true", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"connected":false`) {
		t.Errorf("body = %s, want connected:false", rec.Body)
	}
}

func TestEndpointsRefuseWhenOnlyPairedNotConnected(t *testing.T) {
	srv := pairedButDisconnected()

	rec := httptest.NewRecorder()
	srv.sendText(rec, httptest.NewRequest(http.MethodPost, "/send",
		bytes.NewBufferString(`{"to":"905451483283","text":"hi"}`)))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("send status = %d, want 503 when paired but not connected", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.contacts(rec, httptest.NewRequest(http.MethodGet, "/contacts", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("contacts status = %d, want 503", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.routerVars(rec, "/messages/905451483283", map[string]string{"jid": "905451483283"}, srv.messagesHandler)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("messages status = %d, want 503", rec.Code)
	}
}

func TestHealthStillAnswersWithoutASession(t *testing.T) {
	rec := httptest.NewRecorder()
	New(&fakeWA{}, msgstore.New(10), eventbus.New(), time.Now()).
		health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 even when unpaired", rec.Code)
	}
}
