// whatsapp-bridge-project/bridge-go/internal/httpapi/search_test.go

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"whatsapp-bridge/internal/msgstore"
)

func TestSearch_ReturnsMatchingMessages(t *testing.T) {
	store := msgstore.New(10)
	store.Add(msgstore.Message{ID: "m1", Text: "Merhaba Dünya", Direction: "in", From: "11223344@s.whatsapp.net", Timestamp: 1})
	store.Add(msgstore.Message{ID: "m2", Text: "merhaba dünya tekrar", Direction: "in", From: "11223344@s.whatsapp.net", Timestamp: 2})
	store.Add(msgstore.Message{ID: "m3", Text: "alakasız konu", Direction: "in", From: "11223344@s.whatsapp.net", Timestamp: 3})

	srv := newServer(&fakeWA{connected: true}, store)
	rec := httptest.NewRecorder()
	srv.search(rec, httptest.NewRequest(http.MethodGet, "/search?q=D%C3%9CNYA", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	cevapT := cevap(t, rec)
	if cevapT["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", cevapT["count"])
	}
	if cevapT["query"] != "DÜNYA" {
		t.Errorf("query = %v", cevapT["query"])
	}
}

func TestSearch_EmptyQueryStillOK(t *testing.T) {
	srv := newServer(&fakeWA{connected: true}, msgstore.New(5))
	rec := httptest.NewRecorder()
	srv.search(rec, httptest.NewRequest(http.MethodGet, "/search?q=", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (boş sonuç)", rec.Code)
	}
	if c := cevap(t, rec)["count"].(float64); c != 0 {
		t.Errorf("count = %v, want 0", c)
	}
}

func TestSearch_NoStoreReturns500(t *testing.T) {
	srv := &Server{wa: &fakeWA{connected: true}, messages: nil}
	rec := httptest.NewRecorder()
	srv.search(rec, httptest.NewRequest(http.MethodGet, "/search?q=x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestSearch_RouteRegistered(t *testing.T) {
	// /search route'u Router'da olmalı (yol testi).
	srv := newServer(&fakeWA{connected: true}, msgstore.New(5))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/search?q=test", nil))

	if rec.Code == http.StatusNotFound {
		t.Fatal("/search route'u yok")
	}
}
