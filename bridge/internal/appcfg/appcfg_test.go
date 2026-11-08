// whatsapp-bridge-project/bridge-go/internal/appcfg/appcfg_test.go

package appcfg

import (
	"strings"
	"testing"
)

func TestNormalizeJID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"905451483283", "905451483283@s.whatsapp.net"},
		{"905451483283@s.whatsapp.net", "905451483283@s.whatsapp.net"},
		{"120363123456789@g.us", "120363123456789@g.us"},
		{"12345@lid", "12345@lid"},
		{"", "@s.whatsapp.net"},
	}
	for _, tc := range cases {
		if got := NormalizeJID(tc.in); got != tc.want {
			t.Errorf("NormalizeJID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDerefHelpers(t *testing.T) {
	if got := DerefString(nil); got != "" {
		t.Errorf("DerefString(nil) = %q, want empty", got)
	}
	s := "value"
	if got := DerefString(&s); got != "value" {
		t.Errorf("DerefString(&s) = %q, want value", got)
	}

	if got := DerefUint64(nil); got != 0 {
		t.Errorf("DerefUint64(nil) = %d, want 0", got)
	}
	u := uint64(42)
	if got := DerefUint64(&u); got != 42 {
		t.Errorf("DerefUint64(&u) = %d, want 42", got)
	}
}

func TestListenAddrDefaultAndOverride(t *testing.T) {
	if got := ListenAddr(); got != DefaultListenAddr {
		t.Errorf("ListenAddr() = %q, want %q", got, DefaultListenAddr)
	}
	t.Setenv("WA_BRIDGE_ADDR", "127.0.0.1:9999")
	if got := ListenAddr(); got != "127.0.0.1:9999" {
		t.Errorf("ListenAddr() with env = %q, want override", got)
	}
}

func TestStoreDirIsCreatedAndDSNEndsWithDB(t *testing.T) {
	dir, err := StoreDir()
	if err != nil {
		t.Fatalf("StoreDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("StoreDir() returned an empty path")
	}

	dsn, err := DBDSN()
	if err != nil {
		t.Fatalf("DBDSN() error: %v", err)
	}
	if !strings.HasPrefix(dsn, "file:") {
		t.Errorf("DBDSN() = %q, want a file: DSN", dsn)
	}
	if !strings.Contains(dsn, DBFileName) || !strings.Contains(dsn, "_foreign_keys=on") {
		t.Errorf("DBDSN() = %q, want %s plus foreign keys enabled", dsn, DBFileName)
	}
}
