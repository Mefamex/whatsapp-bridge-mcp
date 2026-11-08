// whatsapp-bridge-project/bridge-go/internal/wabridge/media_send.go

package wabridge

import (
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// buildMediaMessage wraps an upload result in the matching waE2E message type.
func buildMediaMessage(req MediaRequest, up *whatsmeow.UploadResponse) *waE2E.Message {
	caption := proto.String(req.Caption)
	mimetype := proto.String(req.MimeType)

	switch {
	case strings.HasPrefix(req.MimeType, "image/"):
		return &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(up.URL),
			DirectPath:    proto.String(up.DirectPath),
			MediaKey:      up.MediaKey,
			FileEncSHA256: up.FileEncSHA256,
			FileSHA256:    up.FileSHA256,
			FileLength:    proto.Uint64(up.FileLength),
			Mimetype:      mimetype,
			Caption:       caption,
		}}

	case strings.HasPrefix(req.MimeType, "video/"):
		return &waE2E.Message{VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(up.URL),
			DirectPath:    proto.String(up.DirectPath),
			MediaKey:      up.MediaKey,
			FileEncSHA256: up.FileEncSHA256,
			FileSHA256:    up.FileSHA256,
			FileLength:    proto.Uint64(up.FileLength),
			Mimetype:      mimetype,
			Caption:       caption,
		}}

	case strings.HasPrefix(req.MimeType, "audio/"):
		return &waE2E.Message{AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(up.URL),
			DirectPath:    proto.String(up.DirectPath),
			MediaKey:      up.MediaKey,
			FileEncSHA256: up.FileEncSHA256,
			FileSHA256:    up.FileSHA256,
			FileLength:    proto.Uint64(up.FileLength),
			Mimetype:      mimetype,
		}}

	default:
		fileName := req.FileName
		if fileName == "" {
			fileName = "file"
		}
		return &waE2E.Message{DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(up.URL),
			DirectPath:    proto.String(up.DirectPath),
			MediaKey:      up.MediaKey,
			FileEncSHA256: up.FileEncSHA256,
			FileSHA256:    up.FileSHA256,
			FileLength:    proto.Uint64(up.FileLength),
			Mimetype:      mimetype,
			FileName:      proto.String(fileName),
			Title:         caption,
		}}
	}
}
