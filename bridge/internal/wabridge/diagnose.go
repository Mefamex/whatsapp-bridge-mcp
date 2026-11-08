// whatsapp-bridge-project/bridge-go/internal/wabridge/diagnose.go

package wabridge

import (
	"reflect"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
)

// maxReportedFields caps the diagnostic so one malformed message cannot print
// 113 field names into the journal.
const maxReportedFields = 4

// describeUnsupported explains a message we have no branch for. Naming the
// proto field is the point: without it the bridge logs an identical
// "[unsupported]" for a reaction, a revoke and a poll, and nothing on the
// outside tells them apart.
func describeUnsupported(m *waE2E.Message) (kind, text string) {
	if m == nil {
		return "unknown", "[unsupported: nilPayload]"
	}

	names := renderMessage(m)
	if names == "" {
		return "unknown", "[unsupported: noPayload]"
	}
	return "unknown", "[unsupported: " + names + "]"
}

// renderMessage lists the populated proto fields of a message.
//
// Reflection is deliberate: waE2E.Message has 113 fields and gains more every
// few months, so a hand-written list would rot silently. Generated protobuf
// structs also carry unexported bookkeeping (state, sizeCache, unknownFields),
// which IsExported skips. Scalar fields need a non-zero check too — a
// reflection-only listing would miss them entirely.
func renderMessage(m *waE2E.Message) string {
	v := reflect.ValueOf(m).Elem()
	t := v.Type()

	found := make([]string, 0, maxReportedFields)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		f := v.Field(i)
		switch f.Kind() {
		case reflect.Ptr, reflect.Interface:
			if f.IsNil() {
				continue
			}
		case reflect.Bool:
			if !f.Bool() {
				continue
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() == 0 {
				continue
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if f.Uint() == 0 {
				continue
			}
		case reflect.String:
			if f.String() == "" {
				continue
			}
		default:
			continue
		}

		found = append(found, field.Name)
		if len(found) == maxReportedFields {
			found = append(found, "…")
			return strings.Join(found, ", ")
		}
	}

	return strings.Join(found, ", ")
}
