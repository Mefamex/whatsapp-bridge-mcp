# WhatsApp Bridge & MCP — Soru ve Karar Kuyruğu (ASK.md)

> **Kural:** "Tahminle koda girilmez." Çözüm bekleyen mimari belirsizlikler ve alınan otonom kararlar bu dosyada kayıt altına alınır.

---

## 1. Açık Sorular (Cevap Bekleyenler)

*Şu an açık bir mimari soru bulunmamaktadır.*

---

## 2. Kendi Kararım (Otonom Alınan Kararlar)

### [2026-11-08] MCP Klasör Adı: `mcp_server/`
- **Seçenek A:** Klasör adını `mcp/` yapmak.
- **Seçenek B:** Klasör adını `mcp_server/` yapmak.
- **Karar:** Seçenek B (`mcp_server/`).
- **Gerekçe:** Resmi `mcp` Python kütüphanesi (`from mcp.types import ...`) ile yerel proje klasörü aynı isimde olduğunda Python modül gölgelemesi (shadowing) yaşanmakta ve standart kütüphane importları çökmektedir. İsim `mcp_server` yapılarak bu çakışma kesin olarak önlenmiştir.
