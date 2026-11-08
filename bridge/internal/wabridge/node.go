// whatsapp-bridge-project/bridge-go/internal/wabridge/node.go

package wabridge

import (
	"context"
	"fmt"
	"reflect"

	waBinary "go.mau.fi/whatsmeow/binary"
)

// SendNode transmits a raw XML stanza. whatsmeow keeps sendNode unexported, so
// it is reached through reflection; the VoIP sidecar depends on this.
func (s *Service) SendNode(node *waBinary.Node) error {
	if node == nil {
		return fmt.Errorf("nil stanza")
	}

	method := reflect.ValueOf(s.client).MethodByName("sendNode")
	if !method.IsValid() {
		return fmt.Errorf("sendNode unavailable on this whatsmeow build")
	}

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(*node),
	})
	if len(results) > 0 && !results[0].IsNil() {
		return fmt.Errorf("%v", results[0].Interface())
	}
	return nil
}
