// whatsapp-bridge-project/bridge-go/internal/wabridge/login.go

package wabridge

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
)

// loginTimeout bounds how long Connect waits for a usable session.
const loginTimeout = 120 * time.Second

// Connect brings the socket up, pairing via QR code only when no paired
// session is stored.
func (s *Service) Connect() error {
	if s.IsLoggedIn() {
		log.Println("🔄 Mevcut oturum kullanılıyor...")
		if err := s.client.Connect(); err != nil {
			return fmt.Errorf("connect failed: %w", err)
		}
	} else if err := s.pairWithQR(); err != nil {
		return err
	}
	return s.waitForConnection()
}

// waitForConnection blocks until the socket is authenticated.
func (s *Service) waitForConnection() error {
	deadline := time.Now().Add(loginTimeout)
	for time.Now().Before(deadline) {
		if s.client.IsLoggedIn() {
			log.Println("✅ Login başarılı")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("login timeout")
}

// pairWithQR prints pairing codes until WhatsApp confirms the link.
func (s *Service) pairWithQR() error {
	qrChan, err := s.client.GetQRChannel(context.Background())
	if err != nil {
		return fmt.Errorf("qr channel: %w", err)
	}
	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	log.Println("\n═══════════════════════════════════════════")
	log.Println("📱 QR KOD - TELEFONDAN TARAYIN")
	log.Println("═══════════════════════════════════════════")
	log.Println("WhatsApp > Bağlı Cihazlar > Cihaz Bağla")
	log.Println("═══════════════════════════════════════════")

	for item := range qrChan {
		switch {
		case item.Error != nil:
			return fmt.Errorf("qr error: %w", item.Error)
		case item.Event == "success":
			log.Println("✅ QR kod onaylandı!")
			return nil
		case item.Code != "":
			printQRCode(item.Code)
		default:
			log.Printf("[qr] %s", item.Event)
		}
	}
	return fmt.Errorf("qr pairing aborted")
}

// printQRCode renders a pairing code to stdout.
func printQRCode(code string) {
	fmt.Println()
	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)
	fmt.Println()
}
