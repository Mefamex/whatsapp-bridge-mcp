# WhatsApp Bridge & MCP — Dahili Teknik Bellek (NOTES.md)

> **Amaç:** Mimari kararlar, teknik detaylar, elenen alternatifler ve sistem kısıtlarının kalıcı teknik kaydıdır.

---

## 1. Mimari Tasarım ve Protokol Ayrımı

- **Neden Go + Whatsmeow?**
  WhatsApp Web protokolü karmaşık protobuf, websocket ve ikili sinyalleşme katmanlarına sahiptir. Node.js tabanlı `whatsapp-web.js` veya `baileys` sürekli kırılgan oturum kopmaları ve yüksek bellek tüketimi yaşatırken, Go tabanlı `whatsmeow` tek bir ikili dosyada sıfır runtime bağımlılığıyla çalışır, ~30 MB RAM tüketir ve SQLite WAL modunda oturumu aylarca korur.

- **Neden REST + SSE ve Bağımsız MCP?**
  Doğrudan MCP sunucusunu Go içine gömmek yerine, Go çekirdeğini standart bir HTTP REST (`:31415`) ve Server-Sent Events (SSE) yayıncısı yaptık. Bu sayede:
  1. Yalnızca AI ajanları değil, harici betikler (cURL, bash, Python) veya harici bot motorları da köprüye bağlanabilir.
  2. Python tabanlı FastMCP sunucusu sadece bir arayüz istemcisi olarak çalışır; çökse bile Go WhatsApp oturumu kopmaz.

---

## 2. PII ve Veri Güvenliği Standartları

- `whatsmeow.db`, WhatsApp oturum anahtarlarını, önbellek mesajlarını ve temas listesini tutan hassas bir veritabanıdır. `.gitignore` içine sıkı şekilde eklenmiştir.
- Test dosyalarında ve dokümantasyonda asla gerçek telefon numarası yer alamaz; E.164 uyumlu `905551234567` veya RFC örnekleri kullanılır.

---

## 3. Sistem Limitleri ve Deneyimler

- **LID ve Telefon JID Eşlemesi:** WhatsApp çoklu cihaz mimarisinde kişisel kimlikler `@lid` veya `@s.whatsapp.net` olarak gelebilir. Köprünün `directory.go` ve `resolveTarget` modülleri bu kimlikleri otomatik olarak standart telefon JID'sine çözümler.
- **SSE Yeniden Bağlantı:** İstemciler koptuğunda köprü SSE akışında mesaj kaybetmez; bellek içi `msgstore` son mesajları saklar.
