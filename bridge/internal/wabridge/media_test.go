// whatsapp-bridge-project/bridge-go/internal/wabridge/media_test.go

package wabridge

import (
	"testing"

	"go.mau.fi/whatsmeow"
)

func TestDescribeMediaForEachKind(t *testing.T) {
	up := &whatsmeow.UploadResponse{
		URL:        "https://mmg.whatsapp.net/x",
		DirectPath: "/v/t62",
		MediaKey:   []byte{1, 2, 3},
		FileLength: 4096,
	}

	cases := []struct {
		name     string
		req      MediaRequest
		wantKind string
		wantText string
	}{
		{"image with caption", MediaRequest{MimeType: "image/jpeg", Caption: "look"}, "image", "look"},
		{"video with caption", MediaRequest{MimeType: "video/mp4", Caption: "clip"}, "video", "clip"},
		{"audio has no text", MediaRequest{MimeType: "audio/ogg"}, "audio", ""},
		{"document prefers caption", MediaRequest{MimeType: "application/pdf", Caption: "report", FileName: "f.pdf"}, "document", "report"},
		{"document falls back to filename", MediaRequest{MimeType: "application/pdf", FileName: "f.pdf"}, "document", "f.pdf"},
		{"document with neither", MediaRequest{MimeType: "application/pdf"}, "document", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, text, media := describeMedia(tc.req, up)

			if kind != tc.wantKind {
				t.Errorf("kind = %s, want %s", kind, tc.wantKind)
			}
			if text != tc.wantText {
				t.Errorf("text = %q, want %q", text, tc.wantText)
			}
			if media == nil {
				t.Fatal("media metadata is nil")
			}
			if media.DirectPath != up.DirectPath {
				t.Errorf("direct_path = %s, want %s", media.DirectPath, up.DirectPath)
			}
			if media.MediaKey == "" {
				t.Error("media_key should be base64-encoded, not empty")
			}
			if media.FileLength != up.FileLength {
				t.Errorf("file_length = %d, want %d", media.FileLength, up.FileLength)
			}
			if media.MimeType != tc.req.MimeType {
				t.Errorf("mime_type = %s, want %s", media.MimeType, tc.req.MimeType)
			}
		})
	}
}
