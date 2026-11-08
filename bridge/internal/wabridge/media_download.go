// whatsapp-bridge-project/bridge-go/internal/wabridge/media_download.go

package wabridge

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"go.mau.fi/whatsmeow"

	"whatsapp-bridge/internal/msgstore"
)

// ErrMediaIncomplete means the stored message lacks decryption material.
var ErrMediaIncomplete = errors.New("missing media encryption keys or direct path")

// DownloadMedia fetches and decrypts an attachment described by a buffered
// message's media metadata.
func (s *Service) DownloadMedia(ctx context.Context, m *msgstore.Media) ([]byte, error) {
	if m == nil || m.DirectPath == "" || m.MediaKey == "" ||
		m.FileEncSHA256 == "" || m.FileSHA256 == "" {
		return nil, ErrMediaIncomplete
	}

	mediaKey, err := base64.StdEncoding.DecodeString(m.MediaKey)
	if err != nil {
		return nil, ErrMediaIncomplete
	}
	encHash, err := base64.StdEncoding.DecodeString(m.FileEncSHA256)
	if err != nil {
		return nil, ErrMediaIncomplete
	}
	fileHash, err := base64.StdEncoding.DecodeString(m.FileSHA256)
	if err != nil {
		return nil, ErrMediaIncomplete
	}

	return s.client.DownloadMediaWithPath(
		ctx,
		m.DirectPath,
		encHash,
		fileHash,
		mediaKey,
		downloadMediaType(m.MimeType),
		"image", // MMS type; accepted for every attachment kind
		false,
	)
}

// downloadMediaType maps a MIME type to a whatsmeow media category.
func downloadMediaType(mime string) whatsmeow.MediaType {
	switch {
	case strings.HasPrefix(mime, "video/"):
		return whatsmeow.MediaVideo
	case strings.HasPrefix(mime, "audio/"):
		return whatsmeow.MediaAudio
	case strings.HasPrefix(mime, "image/"):
		return whatsmeow.MediaImage
	default:
		return whatsmeow.MediaDocument
	}
}
