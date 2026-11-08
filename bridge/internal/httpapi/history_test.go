// whatsapp-bridge-project/bridge-go/internal/httpapi/history_test.go

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHistoryCountDefaultsAndClamps(t *testing.T) {
	cases := []struct {
		query string
		want  int
	}{
		{"", 20},
		{"?count=5", 5},
		{"?count=abc", 20},
		{"?count=0", 20},
		{"?count=-3", 20},
		{"?count=500", 100},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodGet, "/messages/x"+tc.query, nil)
		if got := historyCount(r); got != tc.want {
			t.Errorf("historyCount(%q) = %d, want %d", tc.query, got, tc.want)
		}
	}
}
