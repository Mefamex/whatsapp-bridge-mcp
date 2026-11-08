// whatsapp-bridge-project/bridge-go/internal/voip/stanza.go

package voip

import (
	waBinary "go.mau.fi/whatsmeow/binary"
)

// MapToNode rebuilds a binary XML node from its JSON representation.
func MapToNode(m map[string]any) *waBinary.Node {
	tag, _ := m["tag"].(string)
	if tag == "" {
		return nil
	}

	attrs := make(waBinary.Attrs)
	if raw, ok := m["attrs"].(map[string]any); ok {
		for k, v := range raw {
			attrs[k] = v
		}
	}

	var content []*waBinary.Node
	if raw, ok := m["content"].([]any); ok {
		for _, child := range raw {
			if childMap, ok := child.(map[string]any); ok {
				if node := MapToNode(childMap); node != nil {
					content = append(content, node)
				}
			}
		}
	}

	return &waBinary.Node{Tag: tag, Attrs: attrs, Content: content}
}

// NodeToMap renders a binary XML node as JSON-friendly data.
func NodeToMap(node *waBinary.Node) map[string]any {
	if node == nil {
		return nil
	}

	out := map[string]any{"tag": node.Tag}

	if len(node.Attrs) > 0 {
		attrs := make(map[string]any, len(node.Attrs))
		for k, v := range node.Attrs {
			attrs[k] = v
		}
		out["attrs"] = attrs
	}

	if nodes, ok := node.Content.([]*waBinary.Node); ok {
		content := make([]any, len(nodes))
		for i, child := range nodes {
			content[i] = NodeToMap(child)
		}
		out["content"] = content
	}

	return out
}
