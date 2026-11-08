// whatsapp-bridge-project/bridge-go/internal/wabridge/service.go

// Package wabridge wraps the whatsmeow client: session handling, sending,
// contacts, history sync and event translation. It is the only package that
// talks to whatsmeow directly.
package wabridge

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	// Registers the "sqlite3" driver with database/sql for the session store.
	_ "github.com/mattn/go-sqlite3"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/msgstore"
	"whatsapp-bridge/internal/voip"
)

// Service owns the WhatsApp client and its session store.
type Service struct {
	client    *whatsmeow.Client
	container *sqlstore.Container
	messages  *msgstore.Store
	bus       *eventbus.Bus

	connectMu     sync.Mutex
	onConnected   func()
	loggedOutHard bool

	// sidecar is the VoIP sidecar for call handling. May be nil if
	// WASM assets are missing.
	sidecar *voip.Sidecar
}

// New opens the session database and builds the client. The returned service
// must be Closed to release the database handle.
func New(bus *eventbus.Bus, messages *msgstore.Store) (*Service, error) {
	dsn, err := appcfg.DBDSN()
	if err != nil {
		return nil, err
	}

	logger := waLog.Stdout("Client", "INFO", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", dsn, logger)
	if err != nil {
		return nil, err
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		container.Close()
		return nil, err
	}

	return &Service{
		client:    whatsmeow.NewClient(device, logger),
		container: container,
		messages:  messages,
		bus:       bus,
	}, nil
}

// Messages exposes the shared in-memory message store.
func (s *Service) Messages() *msgstore.Store { return s.messages }

// OnConnected registers a callback fired once the socket comes up.
func (s *Service) OnConnected(fn func()) {
	s.connectMu.Lock()
	defer s.connectMu.Unlock()
	s.onConnected = fn
}

// IsLoggedIn reports whether a paired session exists, regardless of whether
// the socket is currently up. Use Connected for live readiness.
func (s *Service) IsLoggedIn() bool {
	return s.client.Store.ID != nil && !s.client.Store.ID.IsEmpty()
}

// Connected reports whether the socket is authenticated right now.
func (s *Service) Connected() bool { return s.client.IsLoggedIn() }

// SelfJID returns the account's own JID as a string.
func (s *Service) SelfJID() string {
	if id := s.client.Store.ID; id != nil {
		return id.String()
	}
	return ""
}

// SetHandler registers the raw event entry point.
func (s *Service) SetHandler(fn func(any)) {
	s.client.AddEventHandler(fn)
}

// Close disconnects the client and closes the session database.
func (s *Service) Close() {
	s.client.Disconnect()
	if err := s.container.Close(); err != nil {
		log.Printf("⚠️ [BRIDGE] Session store close error: %v", err)
	}
	if s.sidecar != nil {
		s.sidecar.Stop()
	}
}

// SetSidecar sets the VoIP sidecar for call handling.
func (s *Service) SetSidecar(sc *voip.Sidecar) {
	s.sidecar = sc
}

// SendCallStanza sends a call stanza via the sidecar.
// Expects a call stanza node (call/offer/transport/terminate).
func (s *Service) SendCallStanza(node *waBinary.Node) error {
	if s.sidecar == nil {
		return fmt.Errorf("sidecar not available")
	}
	s.sidecar.Send(map[string]any{
		"type":   "call_stanza_out",
		"stanza": s.nodeToMap(node),
	})
	return nil
}

// nodeToMap converts a binary node to a map for IPC.
func (s *Service) nodeToMap(node *waBinary.Node) map[string]any {
	if node == nil {
		return nil
	}
	m := map[string]any{
		"tag": node.Tag,
	}
	if len(node.Attrs) > 0 {
		attrs := make(map[string]any)
		for k, v := range node.Attrs {
			attrs[k] = v
		}
		m["attrs"] = attrs
	}
	if nodes, ok := node.Content.([]*waBinary.Node); ok {
		content := make([]any, len(nodes))
		for i, child := range nodes {
			content[i] = s.nodeToMap(child)
		}
		m["content"] = content
	}
	return m
}

// fireConnected runs the registered connect hook once.
func (s *Service) fireConnected() {
	s.connectMu.Lock()
	fn := s.onConnected
	s.connectMu.Unlock()
	if fn != nil {
		fn()
	}
}
