# WhatsApp Bridge & MCP — Yapılacaklar (TODO.md)

<!--
  BAĞLAM ÇAPASI (CONTEXT ANCHOR):
  Nihai Amaç: WhatsApp Web protokolünü Whatsmeow Go kütüphanesiyle hafif bir yerel daemon olarak çalıştırmak ve tüm yapay zekâ kodlama ajanlarına (Claude, Cursor, OpenCode, Antigravity) tek komutla Model Context Protocol (MCP) üzerinden WhatsApp araçları sağlamak.
  Mevcut Durum: Go çekirdek köprüsü ve Python FastMCP sunucusu 'whatsapp-bridge-mcp' reposunda birleştirildi; PII verilerinden arındırıldı ve birim testlerle doğrulandı.
  Kırmızı Çizgiler:
    - whatsmeow.db (oturum SQLite dosyası) ASLA git deposuna eklenemez.
    - Kod içinde gerçek telefon numarası, kişisel JID veya secret barındırılamaz.
    - Dağıtım öncesi 'make verify' (lint + test) geçmek zorundadır.
-->

---

## 1. Aktif Sprint (Current Sprint)

- [ ] **Docker Desteği:** Go ikili dosyası ve Python MCP sunucusunu tek bir `Dockerfile` veya compose ile izole çalıştırma seçeneği.
- [ ] **Gelişmiş Medya Alma:** Gelen ses ve sesli mesajları (OGG/Opus) MCP istemcisi üzerinden base64/dosya olarak otomatik dışa aktarma.
- [ ] **Homebrew / AUR Paketleme Kılavuzu:** Arch Linux ve macOS kullanıcıları için hızlı kurulum belgeleri.

---

## 2. Gelecek Planları (Backlog)

- [ ] **Çoklu Oturum (Multi-Session):** Birden fazla WhatsApp numarasını aynı köprüde farklı port veya prefix ile yönetme.
- [ ] **Web Dashboard:** QR kod okutma, oturum durumu izleme ve anlık bağlantı metriği için gömülü minimal HTML paneli.
- [ ] **Webhook İmzası (HMAC-SHA256):** Gelen webhook isteklerinin doğrulanması için opsiyonel HMAC imzalama desteği.
