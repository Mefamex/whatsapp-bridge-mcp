// whatsapp-bridge-project/bridge-go/internal/wabridge/dispatch_test.go

package wabridge

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/msgstore"
)

// testRecorder captures published bus events via a bus relay.
type testRecorder struct {
	events []recordedEvent
}

type recordedEvent struct {
	eventType string
	data      map[string]any
}

// newTestService builds a Service wired to a real event bus whose relay feeds
// the recorder — closes the loop Dispatch → bus without needing the HTTP layer.
func newTestService() (*Service, *testRecorder) {
	bus := eventbus.New()
	rec := &testRecorder{}
	bus.AddRelay(func(eventType string, data any) {
		rec.events = append(rec.events, recordedEvent{eventType, data.(map[string]any)})
	})
	svc := &Service{bus: bus}
	return svc, rec
}

func TestHandleCall_Offer(t *testing.T) {
	svc, rec := newTestService()

	offer := &events.CallOffer{
		From: types.JID{User: "905534410940", Server: "s.whatsapp.net"},
		Data: nil,
	}

	svc.handleCall(offer)

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rec.events))
	}
	ev := rec.events[0]
	if ev.eventType != "call" {
		t.Errorf("eventType = %q, want %q", ev.eventType, "call")
	}
	if ev.data["from"] != "905534410940@s.whatsapp.net" {
		t.Errorf("from = %v", ev.data["from"])
	}
	if ev.data["kind"] != "offer" {
		t.Errorf("kind = %v", ev.data["kind"])
	}
}

func TestHandleCall_Terminate(t *testing.T) {
	svc, rec := newTestService()

	term := &events.CallTerminate{
		From: types.JID{User: "12345", Server: "s.whatsapp.net"},
		Data: nil,
	}

	svc.handleCall(term)

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rec.events))
	}
	if rec.events[0].data["kind"] != "terminate" {
		t.Errorf("kind = %v", rec.events[0].data["kind"])
	}
}

func TestHandleCall_NonCallEvent(t *testing.T) {
	svc, rec := newTestService()

	// A regular message event should not produce a call event
	msg := &events.Message{
		Info: types.MessageInfo{
			ID:   "MSG123",
			Chat: types.JID{User: "905534410940", Server: "s.whatsapp.net"},
		},
	}

	svc.handleCall(msg)

	if len(rec.events) != 0 {
		t.Fatalf("expected 0 events for non-call, got %d", len(rec.events))
	}
}

func TestHandleCall_AllCallKinds(t *testing.T) {
	// Verify all call event types are recognized
	callEvents := []struct {
		name string
		evt  any
	}{
		{"offer", &events.CallOffer{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"preaccept", &events.CallPreAccept{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"accept", &events.CallAccept{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"transport", &events.CallTransport{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"relaylatency", &events.CallRelayLatency{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"offer_notice", &events.CallOfferNotice{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"reject", &events.CallReject{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
		{"terminate", &events.CallTerminate{From: types.JID{User: "1", Server: "s.whatsapp.net"}}},
	}

	for _, tc := range callEvents {
		t.Run(tc.name, func(t *testing.T) {
			svc, rec := newTestService()
			svc.handleCall(tc.evt)
			if len(rec.events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(rec.events))
			}
			if rec.events[0].eventType != "call" {
				t.Errorf("eventType = %q, want %q", rec.events[0].eventType, "call")
			}
		})
	}
}

func TestHandleCall_UnknownCallEvent(t *testing.T) {
	// UnknownCallEvent with a from-bearing node → published as kind "unknown"
	svc, rec := newTestService()
	node := &waBinary.Node{
		Tag:   "call",
		Attrs: waBinary.Attrs{"from": "905534410940@s.whatsapp.net"},
	}

	svc.handleCall(&events.UnknownCallEvent{Node: node})

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rec.events))
	}
	if rec.events[0].data["kind"] != "unknown" {
		t.Errorf("kind = %v", rec.events[0].data["kind"])
	}
	if rec.events[0].data["from"] != "905534410940@s.whatsapp.net" {
		t.Errorf("from = %v", rec.events[0].data["from"])
	}
}

func TestHandleCall_UnknownCallEventNilNodeIgnored(t *testing.T) {
	// Nil node carries no info → silently ignored (matches voip package).
	svc, rec := newTestService()

	svc.handleCall(&events.UnknownCallEvent{})

	if len(rec.events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(rec.events))
	}
}

func TestCallEventOf_ExtractsCallID(t *testing.T) {
	from, kind, node, ok := callEventOf(&events.CallOffer{
		From: types.JID{User: "905534410940", Server: "s.whatsapp.net"},
		Data: nil,
	})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if from != "905534410940@s.whatsapp.net" {
		t.Errorf("from = %q", from)
	}
	if kind != "offer" {
		t.Errorf("kind = %q", kind)
	}
	_ = node
}

func TestUnknownCallFrom_NilNode(t *testing.T) {
	result := unknownCallFrom(nil)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestDispatch_CallEventGoesToBus(t *testing.T) {
	// Integration: Dispatch routes call events to handleCall
	svc, rec := newTestService()

	// Simulate a call offer arriving through Dispatch
	svc.Dispatch(&events.CallOffer{
		From: types.JID{User: "905534410940", Server: "s.whatsapp.net"},
		Data: nil,
	})

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event from Dispatch, got %d", len(rec.events))
	}
	if rec.events[0].eventType != "call" {
		t.Errorf("eventType = %q, want %q", rec.events[0].eventType, "call")
	}
}

func TestDispatch_NonCallEventNoCall(t *testing.T) {
	// A Connected event should not produce a call event
	svc, rec := newTestService()

	svc.Dispatch(&events.Connected{})

	for _, ev := range rec.events {
		if ev.eventType == "call" {
			t.Error("Connected event should not produce a call event")
		}
	}
}

func TestLidTelefonuNilClientGuvenli(t *testing.T) {
	// İstemci/store yokken çökmemeli, "" dönmeli (test servisinde client nil).
	svc, _ := newTestService()
	if got := svc.lidTelefonu("47575939964948@lid"); got != "" {
		t.Errorf("lidTelefonu(nil client) = %q, want \"\"", got)
	}
	if got := svc.lidTelefonu("905451483283@s.whatsapp.net"); got != "" {
		t.Errorf("lidTelefonu(non-LID) = %q, want \"\"", got)
	}
	if got := svc.lidTelefonu("bozuk"); got != "" {
		t.Errorf("lidTelefonu(bozuk) = %q, want \"\"", got)
	}
}

func TestTakmaAdVarTekrariOnler(t *testing.T) {
	if !takmaAdVar([]string{"a", "b"}, "b") {
		t.Error("takmaAdVar var olanı bulamadı")
	}
	if takmaAdVar([]string{"a"}, "b") {
		t.Error("takmaAdVar olmayanı var saydı")
	}
	if takmaAdVar(nil, "b") {
		t.Error("takmaAdVar boş listede var saydı")
	}
}

// fakeLIDStore, LIDStore arayüzünün bellek içi karşılığı.
type fakeLIDStore struct {
	mapping map[types.JID]types.JID
}

func (f *fakeLIDStore) PutManyLIDMappings(_ context.Context, mappings []store.LIDMapping) error {
	for _, m := range mappings {
		f.mapping[m.LID] = m.PN
	}
	return nil
}

func (f *fakeLIDStore) PutLIDMapping(_ context.Context, lid, pn types.JID) error {
	f.mapping[lid] = pn
	return nil
}

func (f *fakeLIDStore) GetPNForLID(_ context.Context, lid types.JID) (types.JID, error) {
	if pn, ok := f.mapping[lid]; ok {
		return pn, nil
	}
	return types.JID{}, fmt.Errorf("eşleşme yok")
}

func (f *fakeLIDStore) GetLIDForPN(_ context.Context, pn types.JID) (types.JID, error) {
	for lid, mapped := range f.mapping {
		if mapped == pn {
			return lid, nil
		}
	}
	return types.JID{}, fmt.Errorf("eşleşme yok")
}

func (f *fakeLIDStore) GetManyLIDsForPNs(_ context.Context, pns []types.JID) (map[types.JID]types.JID, error) {
	out := make(map[types.JID]types.JID)
	for _, pn := range pns {
		if lid, err := f.GetLIDForPN(context.Background(), pn); err == nil {
			out[pn] = lid
		}
	}
	return out, nil
}

// testClient, Store.Device'i fake LID deposuyla kurar (sqlite yok).
func testClient(mapping map[types.JID]types.JID) *whatsmeow.Client {
	if mapping == nil {
		mapping = map[types.JID]types.JID{}
	}
	return &whatsmeow.Client{Store: &store.Device{LIDs: &fakeLIDStore{mapping: mapping}}}
}

func TestHandleMessageLidTakmaAdEklenir(t *testing.T) {
	// Eşleşme varsa: gelen LID mesajın TakmaAd'ında telefon görünmeli,
	// From değişmemeli.
	lid := types.JID{User: "47575939964948", Server: types.HiddenUserServer}
	pn := types.JID{User: "905451483283", Server: types.DefaultUserServer}

	svc := &Service{
		bus:      eventbus.New(),
		client:   testClient(map[types.JID]types.JID{lid: pn}),
		messages: msgstore.New(msgstore.DefaultCapacity),
	}

	svc.handleMessage(&events.Message{
		Info: types.MessageInfo{
			ID:        "WAID-LID-1",
			Chat:      lid,
			Sender:    lid,
			Timestamp: time.Now(),
			PushName:  "Mefamex",
		},
		Message: &waE2E.Message{Conversation: proto.String("merhaba")},
	})

	got, ok := svc.messages.ByID("WAID-LID-1")
	if !ok {
		t.Fatal("mesaj store'a yazılmadı")
	}
	if got.From != lid.String() {
		t.Errorf("from = %q, want %q (değişmemeli)", got.From, lid.String())
	}
	if !takmaAdVar(got.TakmaAd, pn.String()) {
		t.Errorf("takmaAd = %v, want içinde %q", got.TakmaAd, pn.String())
	}
}

func TestHandleMessageLidEslesmeYoksaTakmaAdYok(t *testing.T) {
	// Eşleşme yoksa: TakmaAd boş kalmalı, From yine LID.
	lid := types.JID{User: "11111111111111", Server: types.HiddenUserServer}

	svc := &Service{
		bus:      eventbus.New(),
		client:   testClient(nil),
		messages: msgstore.New(msgstore.DefaultCapacity),
	}

	svc.handleMessage(&events.Message{
		Info: types.MessageInfo{
			ID:        "WAID-LID-2",
			Chat:      lid,
			Sender:    lid,
			Timestamp: time.Now(),
		},
		Message: &waE2E.Message{Conversation: proto.String("selam")},
	})

	got, ok := svc.messages.ByID("WAID-LID-2")
	if !ok {
		t.Fatal("mesaj store'a yazılmadı")
	}
	if got.From != lid.String() {
		t.Errorf("from = %q, want %q", got.From, lid.String())
	}
	if len(got.TakmaAd) != 0 {
		t.Errorf("takmaAd = %v, want boş", got.TakmaAd)
	}
}
