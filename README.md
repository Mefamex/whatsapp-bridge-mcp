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

### Genel Ekosistem Mimarisi (4 Katmanlı Hibrit Model)

```text
                               ┌──────────────────────────────────────────────┐
                               │            WHATSAPP (Mobil / Web)            │
                               └──────────────────────┬───────────────────────┘
                                                      │ WebSocket
                                                      ▼
                       ┌──────────────────────────────────────────────────────────────┐
                       │        KATMAN 1: whatsapp-bridge (Whatsmeow / Go)            │
                       │     - REST API & Gerçek Zamanlı SSE Dinleyici (:31415)       │
                       └──────────────┬───────────────────────────────┬───────────────┘
                                      │                               │
        [İki Yönlü MCP Araçları]      │                               │ [Gelen Olay Akışı (SSE)]
                                      ▼                               ▼
       ┌────────────────────────────────────────────┐  ┌─────────────────────────────────────────────┐
       │     KATMAN 2: whatsapp-mcp (Agent-to-WA)   │  │  KATMAN 3: whatsapp-tetikleyici (Dispatcher)│
       │  - FastMCP STDIO Sunucusu                  │  │  - Whitelist & RBAC (Telefon & LID)         │
       │  - Ajanların WhatsApp'ı Kullanması         │  │  - VRAM Tahliyeli Multimodal (Whisper/Qwen) │
       │  - İnsan Onayı (whatsapp_ask_confirm)      │  └──────────────┬──────────────────────────────┘
       └────────────────────────────────────────────┘                 │
                                                      ┌───────────────┴───────────────┐
                                                      ▼                               ▼
                                      ┌──────────────────────────────┐ ┌─────────────────────────────┐
                                      │ KATMAN 4A: whatsapp-to-cmd   │ │ KATMAN 4B: whatsapp-to-agent│
                                      │   (Doğrudan Sistem Kumandası)│ │     (Ajan & Oturum Tüneli)  │
                                      │ - /cmd & /cmdhub (Bash Tüneli│ │ - /agent agy|claude|opencode│
                                      │ - /camera, /ekran, /durum    │ │ - /session go & /reply      │
                                      └──────────────────────────────┘ └─────────────────────────────┘
```

### Köprü ve MCP Veri Akışı (Bridge & MCP Detail)

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

### Yükseltmeden Sonra Hedef Mimari (Headless Daemon + Birleşik CLI + Gömülü Web Dashboard)

> **Kritik Yükseltme Hedefi:** Ağır masaüstü pencereleri (Electron/Qt) yerine sistem yükünü ~30 MB RAM seviyesinde tutan, hem headless sunucularda (VPS, Raspberry Pi) hem de yerel masaüstünde sıfır ek yükle çalışan; WhatsApp ("ChatOps"), Birleşik CLI ve yerel Gömülü Web Dashboard'u birleştiren hibrit mimari.

```text
                               ┌──────────────────────────────────────────────────────────────┐
                               │               Kullanıcı & Yönetim Arayüzleri                 │
                               │  ┌──────────────────────┐    ┌─────────────────────────────┐ │
                               │  │ WhatsApp ("ChatOps") │    │ Yerel Gömülü Web Dashboard  │ │
                               │  │ /cmd, /camera, /agent│    │ (http://127.0.0.1:31415)    │ │
                               │  │ Anket, Emoji Tepkiler│    │ [Büyük QR Kod, Metrik, Log] │ │
                               │  └──────────┬───────────┘    └──────────────┬──────────────┘ │
                               │             │                               │                │
                               │             │                ┌──────────────┴──────────────┐ │
                               │             │                │    Birleşik Yönetim CLI     │ │
                               │             │                │   `tetikleyici status/qr`   │ │
                               │             │                └──────────────┬──────────────┘ │
                               └─────────────┼───────────────────────────────┼────────────────┘
                                             │ WhatsApp Web                  │ Yerel HTTP /
                                             │ Soketi (Multi-Device)         │ REST & SSE
                                             ▼                               ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│ 1. KÖPRÜ: whatsapp-bridge-mcp (Go Headless Daemon :31415)                                   │
│  - Whatsmeow Çekirdeği (SQLite WAL whatsmeow.db, ~30 MB RAM)                                │
│  - Gömülü Web Sunucusu (go:embed Web Dashboard & QR Motoru)                                 │
│  - HTTP REST API & Gerçek Zamanlı SSE Yayıncısı (/events)                                   │
│  - FastMCP STDIO Katmanı (AI Ajanları için whatsapp_* Araç Seti)                            │
└─────────────────────────────────────────────┬───────────────────────────────────────────────┘
                                              │ Gerçek Zamanlı Olay Akışı (SSE)
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│ 2. DAĞITICI: whatsapp-tetikleyici (Go Headless Daemon)                                      │
│  - Beyaz Liste & RBAC Yetki Denetimi (Telefon & LID Doğrulama)                              │
│  - Catch-up Filtresi (Eski/Bayat Mesajların Açılışta Önlenmesi)                             │
│  - Tünel & Durum Makinesi (/cmdhub Bash Tüneli & /session)                                  │
│  - Ajan Adaptör Mimarisi (Ortak AgentDriver Arayüzü & Sabit Session ID)                     │
└───────┬─────────────────────────────┬───────────────────────────────┬───────────────────────┘
        │                             │                               │
        ▼ [Sistem Komutları]          ▼ [Multimodal Hatlar]           ▼ [AI Ajan Sürücüleri]
 ┌──────────────┐              ┌─────────────────────────────┐ ┌─────────────────────────────┐
 │ Doğrudan CLI │              │   Yerel Çıkarım Boruları    │ │ Sabit Oturumlu AI Adaptörler│
 │ /cmd, /camera│              │   - Whisper (Ses -> Metin)  │ │ - OpenCode Driver           │
 │ /ekran, /ping│              │   - Ollama V-LLM (Görsel)   │ │ - AGY Driver (Antigravity)  │
 │ /zamanla     │              │   (İş bitince VRAM tahliye) │ │ - Claude Driver (Claude CLI)│
 └──────┬───────┘              └──────────────┬──────────────┘ └──────────────┬──────────────┘
        │                                     │                               │
        └─────────────────────────────────────┼───────────────────────────────┘
                                              │
                                              ▼
 ┌────────────────────────────────────────────────────────────────────────────────────────────┐
 │ Formatlanmış Yanıt & Geri Bildirim                                                         │
 │ (POST /send, POST /reaction [⏳->✅], POST /poll-vote -> WhatsApp Köprüsü)                │
 └────────────────────────────────────────────────────────────────────────────────────────────┘
```

- **Masaüstü GUI Bağımlılığı Yok:** X11/Wayland gerektirmez, doğrudan Linux VPS veya Raspberry Pi üzerinde çalışabilir.
- **Sıfır Ekstra Bellek:** Gömülü Web UI doğrudan Go binary içine gömülüdür (`go:embed`), Node.js/Electron gibi fazladan RAM harcamaz.
- **Tek Noktadan Yönetim:** Tüm sistem `tetikleyici status`, `tetikleyici qr`, `tetikleyici logs` gibi birleşik CLI komutlarıyla terminalden kolayca idare edilir.

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
