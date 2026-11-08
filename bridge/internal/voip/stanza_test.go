// whatsapp-bridge-project/bridge-go/internal/voip/stanza_test.go

package voip

import (
	"testing"

	waBinary "go.mau.fi/whatsmeow/binary"
)

func TestMapToNodeRejectsMissingTag(t *testing.T) {
	if got := MapToNode(map[string]any{"attrs": map[string]any{"a": "b"}}); got != nil {
		t.Errorf("MapToNode without tag = %v, want nil", got)
	}
}

func TestMapToNodeBuildsTagAttrs(t *testing.T) {
	node := MapToNode(map[string]any{
		"tag":   "iq",
		"attrs": map[string]any{"id": "1", "to": "905551234567@s.whatsapp.net"},
	})

	if node == nil || node.Tag != "iq" {
		t.Fatalf("unexpected node %+v", node)
	}
	if node.Attrs["id"] != "1" {
		t.Errorf("attrs[id] = %v, want 1", node.Attrs["id"])
	}
}

func TestMapToNodeRecursesIntoContent(t *testing.T) {
	node := MapToNode(map[string]any{
		"tag": "stanza",
		"content": []any{
			map[string]any{"tag": "child1"},
			"not-a-map",
			map[string]any{"tag": "child2"},
		},
	})

	children, ok := node.Content.([]*waBinary.Node)
	if !ok {
		t.Fatalf("content is %T, want []*waBinary.Node", node.Content)
	}
	if len(children) != 2 {
		t.Errorf("children = %d, want 2 (non-map entries dropped)", len(children))
	}
	if children[0].Tag != "child1" || children[1].Tag != "child2" {
		t.Errorf("child order/tags wrong: %q, %q", children[0].Tag, children[1].Tag)
	}
}

func TestNodeToMapRoundTrip(t *testing.T) {
	original := map[string]any{
		"tag":     "message",
		"attrs":   map[string]any{"id": "X"},
		"content": []any{map[string]any{"tag": "body", "attrs": map[string]any{"v": "1"}}},
	}

	back := NodeToMap(MapToNode(original))

	if back["tag"] != "message" {
		t.Errorf("tag = %v, want message", back["tag"])
	}
	attrs, ok := back["attrs"].(map[string]any)
	if !ok || attrs["id"] != "X" {
		t.Errorf("attrs = %v, want map with id=X", back["attrs"])
	}
	content, ok := back["content"].([]any)
	if !ok || len(content) != 1 {
		t.Fatalf("content = %v, want 1 child", back["content"])
	}
	child := content[0].(map[string]any)
	if child["tag"] != "body" {
		t.Errorf("child tag = %v, want body", child["tag"])
	}
}

func TestNodeToMapNilIsNil(t *testing.T) {
	if got := NodeToMap(nil); got != nil {
		t.Errorf("NodeToMap(nil) = %v, want nil", got)
	}
}

func TestJidTripleFromPhoneJID(t *testing.T) {
	self, phone, lid := jidTriple("905550000005@s.whatsapp.net")
	if self != "905550000005@s.whatsapp.net" {
		t.Errorf("self = %s", self)
	}
	if phone != self {
		t.Errorf("phone = %s, want %s", phone, self)
	}
	if lid != "905550000005@lid" {
		t.Errorf("lid = %s, want 905550000005@lid", lid)
	}
}

func TestJidTripleFromLID(t *testing.T) {
	self, phone, lid := jidTriple("905550000005@lid")
	if phone != "905550000005@s.whatsapp.net" {
		t.Errorf("phone = %s, want @s.whatsapp.net", phone)
	}
	if lid != self {
		t.Errorf("lid = %s, want %s", lid, self)
	}
}
