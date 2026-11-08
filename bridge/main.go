// whatsapp-bridge-project/bridge-go/main.go

// Command whatsapp-bridge runs the local WhatsApp bridge: a whatsmeow session
// plus a loopback HTTP/SSE API consumed by the MCP server.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"whatsapp-bridge/internal/appcfg"
	"whatsapp-bridge/internal/eventbus"
	"whatsapp-bridge/internal/httpapi"
	"whatsapp-bridge/internal/msgstore"
	"whatsapp-bridge/internal/voip"
	"whatsapp-bridge/internal/wabridge"
	"whatsapp-bridge/internal/webhook"
)

func main() {
	// Tek kapatma kaynağı: SIGINT/SIGTERM bağlamı. Hem bağlantı döngüsü hem
	// kapanış bunu dinler; iki ayrı sinyal yolu olursa biri unutulur.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	started := time.Now()

	bus := eventbus.New()
	bus.AddRelay(webhook.Relay)

	// Message log: JSONL replay sayesinde restart'ta /messages ve çift
	// gönderim geçmişi siyah kutusu korunur. Log açılamazsa sessizce
	// in-memory'e düşmek YERİNE nedenini loglarız — "Stored 0" vakalarının
	// kökü buydu.
	messages := msgstore.New(msgstore.DefaultCapacity)
	if logPath, perr := appcfg.MessagesLogPath(); perr == nil {
		if opened, oerr := msgstore.Open(msgstore.DefaultCapacity, logPath); oerr == nil {
			messages = opened
			defer opened.Close()
		} else {
			log.Printf("⚠️ message log açılamadı (%v) — in-memory devam", oerr)
		}
	} else {
		log.Printf("⚠️ message log yolu yok (%v) — in-memory devam", perr)
	}

	wa, err := wabridge.New(bus, messages)
	if err != nil {
		log.Fatalf("WhatsApp session init failed: %v", err)
	}
	defer wa.Close()

	sidecar := voip.New(bus, wa, wa.SelfJID)
	wa.SetSidecar(sidecar) // Wire sidecar for call handling
	wa.SetHandler(newRouter(wa, sidecar).handle)
	wa.OnConnected(sidecar.Start)
	defer sidecar.Stop()

	api := httpapi.New(wa, messages, bus, started)
	srv := &http.Server{
		Addr:        appcfg.ListenAddr(),
		Handler:     api.Router(),
		ReadTimeout: 30 * time.Second,

		// WriteTimeout must stay off. It is a deadline for the *whole*
		// response, so it also caps streaming endpoints: /events is a
		// long-lived SSE connection that is idle most of the time, and a
		// 30s write deadline tore it down every 30 seconds. Clients then
		// spent the reconnect backoff waiting for the next message instead
		// of reacting to it, which showed up as 0-30s trigger latency.
		//
		// The listener is loopback-only, so the slow-body exposure a write
		// deadline would protect against does not apply here.
		WriteTimeout: 0,
	}

	go serveAPI(srv)
	go connectWhatsApp(ctx, wa)

	<-ctx.Done()
	log.Printf("kapatma sinyali alındı (%v)", ctx.Err())

	shutdown(srv)
}

// serveAPI runs the HTTP listener in the background.
func serveAPI(srv *http.Server) {
	log.Printf("🌐 [API] HTTP server starting on http://%s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// connectWhatsApp brings the WhatsApp session up without blocking startup.
// connectWhatsApp, WhatsApp soketine bağlanır ve BAŞARISIZ OLursa tekrar dener.
//
// Neden tekrar denemek zorunda: systemd servisi boot'ta ağ hazır olmadan
// başlar. 2026-10-05 rebootunda ölçüldü:
//
//	00:44:37 ❌ Connection failed: … lookup web.whatsapp.com: Temporary failure
//	in name resolution
//
// ve köprü bir daha DENEMEDİ — API yanıt veriyor, `connected:false` diyor,
// mesaj gelmiyor. Her açılışta böyle olur ve tek çözüm "sudo systemctl restart".
//
// whatsmeow başarılı bağlantıdan sonra KENDİ yeniden bağlanır; bu yüzden
// döngü yalnız başarısızlıkta çalışır, bağlantı kurulunca çıkar.
func connectWhatsApp(ctx context.Context, wa *wabridge.Service) {
	backoff := 2 * time.Second

	for ctx.Err() == nil {
		err := wa.Connect()
		if err == nil {
			return
		}

		// Ağ henüz yoksa (boot sırası) hata DNS kaynaklıdır; ayrıca belirt.
		// Gerçek oturum/session hatası ise sürekli tekrar denemek anlamsız
		// olur — günlüğe net yazılır, kullanıcı kararını verir.
		if ctx.Err() != nil {
			return
		}

		log.Printf("❌ Bağlantı kurulamadı: %v — %.0fs sonra tekrar denenecek (ağ henüz hazır olmayabilir)",
			err, backoff.Seconds())

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		// 2s → 4s → 8s → 16s → 30s (tavan). Sinir koymazsak ağ 30 dakika
		// gelmediğinde saniyede bir log şişer.
		if backoff < 30*time.Second {
			backoff *= 2
		}
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
	}
}

// shutdown drains in-flight HTTP requests before exit.
func shutdown(srv *http.Server) {
	log.Println("🛑 Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
