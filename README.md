# whatsapp-bridge-mcp

|              |                                                               |
| ------------ | ------------------------------------------------------------- |
| **Author**   | [Mehmet Akif Akkoç (Mefamex)](https://github.com/Mefamex)     |
| **Website**  | [https://github.com/Mefamex/whatsapp-bridge-mcp](https://github.com/Mefamex/whatsapp-bridge-mcp) |
| **License**  | MIT                                                           |
| **Created**  | 2026-11-08                                                    |
| **Updated**  | 2026-11-08                                                    |
| **Stack**    | Go 1.22+, Python 3.10+, Whatsmeow, FastMCP, SQLite            |
| **Platform** | Linux, macOS, Windows (WSL)                                   |
| **Status**   | Production Ready                                              |

> WhatsApp Web protokolünü Whatsmeow Go kütüphanesiyle hafif bir yerel arka plan servisi (`systemd`/daemon) olarak çalıştıran; Claude, Cursor, OpenCode ve Antigravity gibi AI kodlama ajanlarına Model Context Protocol (MCP) üzerinden iki yönlü WhatsApp yetenekleri kazandıran açık kaynaklı köprü ve araç seti.

---

## 1. Mimari ve Görsel Önizleme (Architecture & Preview)

```text
┌───────────────────────┐         ┌────────────────────────┐         ┌─────────────────────────┐
│   WhatsApp Sunucusu   │ <=====> │  WhatsApp Bridge (Go)  │ <=====> │ WhatsApp MCP (Python)   │
│   (Multi-Device Web)  │  Proto  │  - Port :31415         │  HTTP   │ - FastMCP STDIO         │
└───────────────────────┘         │  - REST API & SSE      │  REST   └────────────┬────────────┘
                                  │  - whatsmeow.db (WAL)  │                      │
                                  └────────────────────────┘                      │ MCP Protokolü
                                                                                  ▼
                                                                     ┌─────────────────────────┐
                                                                     │     AI Kodlama Ajanı    │
                                                                     │  (Claude, Cursor,       │
                                                                     │   OpenCode, Antigravity)│
                                                                     └─────────────────────────┘
```

- **Hafif Bellek İzi:** Go çekirdeği yalnızca ~30 MB RAM tüketir, Node.js veya Chromium gerektirmez.
- **İki Katmanlı Güven:** MCP istemcisi veya yapay zekâ oturumu çökse bile WhatsApp oturumu Go servisinde kesintisiz devam eder.

---

## 2. Belge Haritası (Repository Doc Map)

| Belge | Amacı ve Sorumluluğu |
| :--- | :--- |
| [`README.md`](./README.md) | Proje vitrini, mimari şema, kurulum ve MCP entegrasyon kılavuzu |
| [`TODO.md`](./TODO.md) | Aktif sprint, bağlam çapası ve gelecek geliştirme hedefleri |
| [`TODO.DONE.md`](./TODO.DONE.md) | Doğrulanmış iş arşivi (ters kronolojik, test çıktılarıyla) |
| [`ASK.md`](./ASK.md) | Açık mimari sorular ve alınan otonom teknik kararlar |
| [`NOTES.md`](./NOTES.md) | Dahili teknik bellek, protokol detayları ve sistem sınırları |
| [`config.yaml`](./config.yaml) | Kaskad yapılandırma ayarları ve çevre değişkeni şablonu |
| [`Makefile`](./Makefile) | Tek komutla bağımlılık, derleme, test ve temizlik yönetimi |

---

## 3. Temel Yetenekler ve Özellikler (Key Features)

* **Resmi Whatsmeow Çekirdeği:** Bağımsız Go ikili dosyası ile Chromium veya Puppeteer olmadan doğrudan WhatsApp Web soketine bağlanır.
* **Standart Model Context Protocol (MCP):** Ajanların araç çağırma (tool call) standartlarıyla birebir uyumlu FastMCP Python sunucusu.
* **İki Yönlü Mesajlaşma:** Metin mesajları gönderme (`whatsapp_send`), okuma (`whatsapp_get_messages`), okundu işaretleme (`whatsapp_mark_read`).
* **Medya ve Anket Desteği:** Resim/dosya gönderimi (`whatsapp_send_media`) ve oylama/anket başlatma (`whatsapp_send_poll`).
* **İnteraktif Ajan Onay Mekanizması:** Ajanların tehlikeli bir işlem öncesi WhatsApp'tan doğrudan soru sorup onay alabilmesi (`whatsapp_ask_confirm`, `whatsapp_ask_quick`).
* **Olay Akışı (SSE):** Gelen her mesaj ve durum güncellemesini gerçek zamanlı Server-Sent Events akışında (`GET /events`) yayınlar.

---

## 4. Teknoloji Yığını ve Gereksinimler (Tech Stack & Prerequisites)

* **Go:** 1.22 veya daha yenisi
* **Python:** 3.10 veya daha yenisi (ve modern paket yöneticisi `uv`)
* **SQLite:** 3.x (Whatsmeow oturum anahtarları için yerleşik)
* **İşletim Sistemi:** Linux, macOS veya WSL

---

## 5. Kurulum ve Hızlı Başlangıç (Installation & Quick Start)

### 1. Depoyu Klonlayın
```bash
git clone https://github.com/Mefamex/whatsapp-bridge-mcp.git
cd whatsapp-bridge-mcp
```

### 2. Bağımlılıkları Kurun
```bash
make install
```

### 3. WhatsApp Bridge'i Derleyin ve Başlatın
```bash
# İlk çalıştırmada terminalde QR kod belirir; WhatsApp uygulamanızdan okutun
make run
```

### 4. MCP Sunucusunu Ajanınıza Ekleyin
`claude_desktop_config.json`, `opencode.json` veya Antigravity MCP ayarlarına şu bloğu ekleyin:

```json
{
  "mcpServers": {
    "whatsapp": {
      "command": "/proje/yolu/whatsapp-bridge-mcp/.venv/bin/python",
      "args": ["-m", "mcp_server"],
      "cwd": "/proje/yolu/whatsapp-bridge-mcp"
    }
  }
}
```

---

## 6. Yapılandırma (Configuration Guide)

Öncelik Sırası: `Ortam Değişkenleri (ENV) > config.yaml > Dahili Varsayılanlar`

| Parametre | Ortam Değişkeni (ENV) | Varsayılan | Açıklama |
| :--- | :--- | :--- | :--- |
| `bridge.host` | `WHATSAPP_API_HOST` | `127.0.0.1` | Bridge REST dinleme adresi |
| `bridge.port` | `WHATSAPP_API_PORT` | `31415` | Bridge REST ve SSE portu |
| `bridge.db_path` | `WHATSAPP_DB_PATH` | `whatsmeow.db` | Oturum veritabanı dosyası |
| `mcp.api_base` | `WHATSAPP_API_BASE` | `http://127.0.0.1:31415` | MCP istemcisinin köprü adresi |

---

## 7. Kullanım ve Komutlar (Usage & CLI Reference)

Tüm yaşam döngüsü komutları self-documenting `Makefile` üzerinden yönetilir:

```bash
make            # Kullanılabilir tüm komutları listeler
make build      # Go bridge ikilisini bin/whatsapp-bridge olarak derler
make run        # Bridge arka plan servisini başlatır
make run-mcp    # MCP STDIO sunucusunu manuel test için çalıştırır
make test       # Go ve Python testlerini koşturur
make verify     # Commit öncesi lint ve test kontrollerini tamamlar
make clean      # Geçici dosyaları ve önbellekleri temizler
```

---

## 8. Güvenlik, Gizlilik ve Kırmızı Çizgiler (Safety & Privacy)

1. **Sıfır Secret İlkesi:** WhatsApp oturum anahtarlarını barındıran `whatsmeow.db` dosyası ASLA Git'e eklenemez (`.gitignore` ile korunmaktadır).
2. **Yerel Sınır:** Bridge API varsayılan olarak `127.0.0.1` dinler; dış internete şifresiz açılmamalıdır.
3. **Kişisel Veri (PII) Hijyeni:** Kod tabanında, testlerde ve loglarda kişisel telefon numarası veya JID yer alamaz.

---

## 9. Lisans ve Katkı (License)

Bu proje [MIT Lisansı](./LICENSE) altında açık kaynak olarak sunulmuştur.  
Geliştirici: **Mehmet Akif Akkoç (Mefamex)**
