// whatsapp-bridge-project/bridge-go/internal/eventbus/bus_test.go

package eventbus

import "testing"

func TestPublishReachesSubscriber(t *testing.T) {
	b := New()
	ch, cancel := b.Subscribe(4)
	defer cancel()

	b.Publish("message", map[string]any{"text": "hi"})

	select {
	case frame := <-ch:
		if got := string(frame); len(got) == 0 || got[:6] != "data: " {
			t.Fatalf("frame = %q, want SSE data frame", got)
		}
	default:
		t.Fatal("expected a frame on the subscriber channel")
	}
}

func TestSubscribeCancelStopsDelivery(t *testing.T) {
	b := New()
	ch, cancel := b.Subscribe(4)
	cancel()

	b.Publish("message", nil)

	if _, open := <-ch; open {
		t.Error("channel should be closed after cancel")
	}
}

func TestSlowSubscriberDoesNotBlockPublish(t *testing.T) {
	b := New()
	_, cancel := b.Subscribe(1)
	defer cancel()

	done := make(chan struct{})
	go func() {
		// Far more events than the buffer can hold.
		for i := 0; i < 50; i++ {
			b.Publish("message", i)
		}
		close(done)
	}()

	<-done // must not deadlock
}

func TestRelayReceivesStructuredPayload(t *testing.T) {
	b := New()

	var gotType string
	var gotData any
	b.AddRelay(func(eventType string, data any) {
		gotType, gotData = eventType, data
	})

	payload := map[string]any{"id": "abc"}
	b.Publish("receipt", payload)

	if gotType != "receipt" {
		t.Errorf("event type = %q, want receipt", gotType)
	}
	if gotData == nil {
		t.Error("relay received nil data")
	}
}

func TestPublishWithNoSubscribersDoesNotPanic(t *testing.T) {
	New().Publish("connected", map[string]any{})
}
