// whatsapp-bridge-project/bridge-go/internal/appcfg/appcfg.go

// Package appcfg holds process-wide configuration: paths, defaults and small
// value helpers shared across the bridge packages.
package appcfg

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultListenAddr is the loopback address the HTTP API binds to.
	DefaultListenAddr = "127.0.0.1:31415"

	// StoreDirName / DBDirName locate the whatsmeow session database.
	StoreSubDir = ".local/share/whatsapp-bridge"
	DBFileName  = "session.db"

	// MessagesFileName is the JSONL log backing the in-memory message store.
	// Each line is one JSON Message; the store replays the last capacity
	// entries on startup so /messages survives a bridge restart.
	MessagesFileName = "messages.jsonl"

	// VoIP sidecar layout (Node.js process driving WhatsApp Web call code).
	SidecarHomeSubDir = "Desktop/whatsapp-bridge-project/wa-voip-sidecar"
	SidecarEntry      = "index.js"
	SidecarStateDir   = ".wa-voip-sidecar"
	SidecarWAMSubDir  = "cache"
	SidecarLogFile    = "sidecar.log"
	SidecarLoaderFile = "WAWebVoipWebWasmLoader.js"

	// ProjectSubDir is the whatsapp-bridge checkout.
	ProjectSubDir = "Desktop/whatsapp-bridge-project"

	// StateDirName holds everything the helper commands write: trigger state,
	// payload dumps, logs. Kept inside the project so nothing litters $HOME
	// or /tmp.
	StateDirName = ".tetikleyici"

	// SeenFileName records which message IDs already fired.
	SeenFileName = "seen.txt"
	// PayloadFileName is the default JSON dump of the last trigger.
	PayloadFileName = "payload.json"
	// LogFileName is the default log file.
	LogFileName = "tetikleyici.log"
)

// ListenAddr returns the API bind address, overridable with WA_BRIDGE_ADDR.
func ListenAddr() string {
	if v := os.Getenv("WA_BRIDGE_ADDR"); v != "" {
		return v
	}
	return DefaultListenAddr
}

// StoreDir returns (and creates) the directory holding session.db.
func StoreDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, filepath.FromSlash(StoreSubDir))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// DBDSN returns the SQLite DSN for the session store.
func DBDSN() (string, error) {
	dir, err := StoreDir()
	if err != nil {
		return "", err
	}
	return "file:" + filepath.Join(dir, DBFileName) + "?_foreign_keys=on", nil
}

// MessagesLogPath returns the JSONL file backing the message store. It lives
// next to session.db so the bridge owns its whole state directory.
func MessagesLogPath() (string, error) {
	dir, err := StoreDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, MessagesFileName), nil
}

// SidecarDir returns the Node.js sidecar project directory.
func SidecarDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, filepath.FromSlash(SidecarHomeSubDir))
}

// ProjectDir returns the project checkout directory.
func ProjectDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, filepath.FromSlash(ProjectSubDir)), nil
}

// StateDir returns the project-local directory for helper command output,
// creating it if needed. WA_TRIGGER_STATE_DIR overrides the location.
//
// Everything the tetikleyici writes lives here rather than in $HOME or /tmp,
// so the state is easy to inspect and easy to delete.
func StateDir() (string, error) {
	if override := os.Getenv("WA_TRIGGER_STATE_DIR"); override != "" {
		return ensureDir(override)
	}

	project, err := ProjectDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(project, StateDirName))
}

// ensureDir creates dir with owner-only permissions and returns it.
func ensureDir(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// SidecarWAMCache returns the directory holding the cached call WASM binary.
func SidecarWAMCache() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, SidecarStateDir, SidecarWAMSubDir)
}

// SidecarLogPath returns the sidecar stderr log file path.
func SidecarLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, SidecarStateDir, SidecarLogFile)
}

// NormalizeJID appends the default WhatsApp server suffix when missing.
func NormalizeJID(jid string) string {
	if strings.Contains(jid, "@") {
		return jid
	}
	return jid + "@s.whatsapp.net"
}

// DerefString safely dereferences an optional proto string.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefUint64 safely dereferences an optional proto uint64.
func DerefUint64(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}
