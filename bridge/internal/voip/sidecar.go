// whatsapp-bridge-project/bridge-go/internal/voip/sidecar.go

// Package voip bridges WhatsApp call stanzas to the Node.js sidecar that
// executes WhatsApp Web's call code.
package voip

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/eventbus"
)

const (
	// initDelay gives the Node process time to boot before the handshake.
	initDelay = 3 * time.Second

	// selfJIDRetryDelay covers the case where the sidecar starts before the
	// session has reported its own JID.
	selfJIDRetryDelay = 2 * time.Second

	// selfJIDRetryLimit caps those retries.
	selfJIDRetryLimit = 15
)

// StanzaSender delivers raw stanzas back to WhatsApp.
type StanzaSender interface {
	SendNode(node *waBinary.Node) error
}

// Sidecar supervises the Node.js call process and its JSON-lines IPC channel.
type Sidecar struct {
	bus     *eventbus.Bus
	sender  StanzaSender
	selfJID func() string

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	ready  bool
	closed bool
}

// New creates an unstarted sidecar supervisor. sender pushes stanzas back to
// WhatsApp; selfJID resolves the logged-in account lazily.
func New(bus *eventbus.Bus, sender StanzaSender, selfJID func() string) *Sidecar {
	return &Sidecar{bus: bus, sender: sender, selfJID: selfJID}
}

// Start launches the sidecar if its WASM assets are present. Missing assets are
// a warning, not a fatal error: text messaging keeps working without calls.
func (s *Sidecar) Start() {
	dir := appcfg.SidecarDir()
	if err := checkAssets(dir); err != nil {
		log.Printf("⚠️ [Sidecar] %v — calling features disabled", err)
		log.Printf("💡 [Sidecar] Run 'cd %s && npm run fetch-wasm' to enable calls", dir)
		return
	}

	cmd := buildCommand(dir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("❌ [Sidecar] stdin pipe: %v", err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("❌ [Sidecar] stdout pipe: %v", err)
		return
	}
	attachStderr(cmd)

	if err := cmd.Start(); err != nil {
		log.Printf("❌ [Sidecar] start: %v", err)
		return
	}

	s.mu.Lock()
	s.cmd, s.stdin, s.closed = cmd, stdin, false
	s.mu.Unlock()

	log.Printf("🚀 [Sidecar] Started (PID: %d)", cmd.Process.Pid)
	go s.readLoop(stdout)
	go s.waitLoop(cmd)
	go s.handshake()
}

// Stop terminates the sidecar process.
func (s *Sidecar) Stop() {
	s.mu.Lock()
	if s.closed || s.cmd == nil || s.cmd.Process == nil {
		s.closed = true
		s.mu.Unlock()
		return
	}
	s.closed = true
	cmd := s.cmd
	s.mu.Unlock()

	log.Println("🛑 [Sidecar] Stopping...")
	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()
	log.Println("✅ [Sidecar] Stopped")
}

// Send writes one JSON message to the sidecar.
func (s *Sidecar) Send(msg map[string]any) {
	s.mu.Lock()
	stdin := s.stdin
	s.mu.Unlock()

	if stdin == nil {
		log.Println("⚠️ [Sidecar] Not connected")
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	if _, err := stdin.Write(append(data, '\n')); err != nil {
		log.Printf("❌ [Sidecar] write: %v", err)
	}
}

// checkAssets verifies the sidecar project and its cached call WASM exist.
func checkAssets(dir string) error {
	loader := filepath.Join(dir, appcfg.SidecarLoaderFile)
	if _, err := os.Stat(loader); err != nil {
		return &missingAssetError{what: "WASM loader", path: loader}
	}

	cache := appcfg.SidecarWAMCache()
	entries, err := os.ReadDir(cache)
	if err != nil {
		return &missingAssetError{what: "WASM cache", path: cache}
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".wasm") {
			return nil
		}
	}
	return &missingAssetError{what: "WASM binary", path: cache}
}

// buildCommand prepares the Node process definition.
func buildCommand(dir string) *exec.Cmd {
	cmd := exec.Command("node", appcfg.SidecarEntry)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	return cmd
}

// attachStderr redirects sidecar stderr to its own log file so the bridge's
// own log output stays readable.
func attachStderr(cmd *exec.Cmd) {
	path := appcfg.SidecarLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	cmd.Stderr = f
}

// readLoop consumes sidecar stdout until the pipe closes.
func (s *Sidecar) readLoop(stdout io.Reader) {
	dec := json.NewDecoder(stdout)
	for {
		var msg map[string]any
		if err := dec.Decode(&msg); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("❌ [Sidecar] decode: %v", err)
			}
			return
		}
		s.handleMessage(msg)
	}
}

// waitLoop reports unexpected sidecar exits.
func (s *Sidecar) waitLoop(cmd *exec.Cmd) {
	log.Printf("🛑 [Sidecar] Process exited: %v", cmd.Wait())

	s.mu.Lock()
	s.ready = false
	s.mu.Unlock()
}

// handshake identifies the account to the sidecar shortly after startup.
func (s *Sidecar) handshake() {
	time.Sleep(initDelay)

	for attempt := 0; attempt < selfJIDRetryLimit; attempt++ {
		self, phone, lid := jidTriple(s.selfJID())
		if self != "" {
			s.Send(map[string]any{
				"type":     "init",
				"selfJid":  self,
				"phoneJid": phone,
				"lidJid":   lid,
			})
			return
		}
		time.Sleep(selfJIDRetryDelay)
	}
	log.Println("⚠️ [Sidecar] Could not determine own JID, skipping init")
}
