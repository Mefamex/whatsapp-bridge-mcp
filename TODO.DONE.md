# WhatsApp Bridge & MCP — Doğrulanmış İş Arşivi (TODO.DONE.md)

> **Kural:** Ters kronolojik sıralama (en yeni en üstte). Her tamamlanan işin altında çalıştırılan doğrulama komutu ve test kanıtı yer alır.

---

### [2026-11-08 04:11] Proje İskeleti, PII Temizliği ve MCP Doğrulama Zinciri

- **Açıklama:** WhatsApp Bridge Go çekirdeği ile Python FastMCP sunucusu `whatsapp-bridge-mcp` reposunda bağımsız bir depoda toplandı. Kaynak kod ve testlerdeki kişisel telefon ve JID bilgileri temizlenerek dummy test değerleriyle değiştirildi. Marmelab self-documenting Makefile yazıldı ve doğrulama zinciri (`make verify`) kuruldu.
- **Doğrulama Komutu:** `make verify`
- **Çıktı / Kanıt:**
  ```text
  ==> Go vet analizi yapılıyor...
  cd bridge && go vet ./...
  ==> Python sözdizimi denetleniyor...
  .venv/bin/python -m py_compile mcp_server/*.py mcp_server/**/*.py
  ==> Go birim testleri koşturuluyor...
  cd bridge && go test ./...
  ok  	whatsapp-bridge/internal/appcfg	(cached)
  ok  	whatsapp-bridge/internal/eventbus	(cached)
  ok  	whatsapp-bridge/internal/httpapi	(cached)
  ok  	whatsapp-bridge/internal/msgstore	(cached)
  ok  	whatsapp-bridge/internal/voip	(cached)
  ok  	whatsapp-bridge/internal/wabridge	(cached)
  ==> Python MCP testleri koşturuluyor...
  .venv/bin/python -m unittest discover -s tests -p 'test_*.py' -v
  Ran 5 tests in 0.001s - OK
  ==> TÜM DOĞRULAMALAR BAŞARIYLA GEÇTİ (ALL_OK).
  ```
