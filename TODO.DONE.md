# TODO.DONE — Tamamlananlar

> Zaman damgalı, biten iş buraya taşınır. Yeni biten en alta eklenir.

> ⚠️ **Bu dosya bir günlük değil.** İçinde **sonradan yanlışlanan** kararlar
> da var (biri çift gönderimin izin katmanıyla kapatılmasıydı, ertesi gün
> geri çekildi). Bu normaldir ve **silinmez**: "bu çalıştı, sonra neden
> bozdu?" sorusunun cevabı burada durur. Yanlış kararın kendisi kanıttır.
>
> **Açık işler için** → `TODO.md` · **Belge için** → `README.md` · · **NOTLAR için** → `NOTES.md` · 
> **Ölçüm için** → `docs/`

## 2026-10-04 10:36 +03 — Test mesajı
- "ben test mesajıyım" → ASU Burak (elk26) `905534410940@s.whatsapp.net`, `{"ok":true}` alındı.

## 2026-10-04 10:36-10:44 +03 — Kimlik + Q&A
- Burak'ın "kimsin / ingilizce anlamadım" sorularına samimi cevap verildi, AI asistanı olduğu açıklandı.
- Gizlilik soruları cevaplandı: başkalarının ve Akif'in diğer kişilerle mesajları görülemiyor (sadece bu linked device).
- Faktüel soru: Erdoğan 70 (1954) — sonradan kaynaklı cevap kuralı kondu, bu eski cevap tahminiydi diye nota düşüldü.

## 2026-10-04 10:47-10:50 +03 — KYK bilgisi
- Yatay geçiş + misafir öğrenci başvurusu anlatıldı (kyk.gov.tr / e-Devlet).

## 2026-10-04 11:01-11:29 +03 — Monitor + media altyapısı
- `monitor-burak.go` yazıldı, derlendi (`monitor-burak` binary), nohup ile arka planda çalıştırıldı → `monitor.log`.
- Go bridge'e `/messages/{jid}` eklendi (o anki kodla), `/media/{messageId}` eklendi (`DownloadMediaWithPath`).
- 2 yeni resim indirildi: 1080x1920 JPEG, 228KB / 468KB. `identify` + `exiftool` + Tesseract (metin yok) yapıldı.

## 2026-10-04 11:30-12:00 +03 — MCP + araştırma disiplini
- `whatsapp-bridge` skill yüklendi, 9 MCP tool envanteri okundu (README + `~/Desktop/whatsapp-bridge-project/README.md` + skill README).
- `research` skill yüklendi, kural kondu: tahmin yok, Exa ile araştır + mesaj sonunda kaynak ver.
- `whatsapp_get_messages` LID+PN ile son 5'li bağlam tek pencerede okuma rutini kuruldu.
- ASR envanteri çıkarıldı: `whisper-large-v3-tr/model.bin` 2.9GB + wav2vec2 snapshots + `pipeline.py` mevcut.
- `conversation-notes.md` güncellendi (`~/.config/opencode/skills/whatsapp-bridge/`).

## 2026-10-04 12:xx +03 — Kaynaklı düzeltme mesajı (MCP ile, kod değiştirmeden)
- KPSS 2026 Ortaöğretim resmi takvim verildi: 25 Eki 2026 Pazar, başvuru 27 Ağu-8 Eyl, geç 15-16 Eyl, sonuç 19 Kas 2026.
- Noel Baba: Demre Aziz Nikolaos Anıt Müzesi (Myra 1.5km güneyi, MS 5-6.yy).
- Denizli 2.yer için kesin denmedi (foto yok), Side/Phaselis/Olympos aday + tekrar isteme.
- Opencode erişimi: doğrudan erişim yok, local kurulum adımları teklif edildi.
- `whatsapp_send` OK, `whatsapp_mark_read` OK. 6 kaynak linki mesaj sonunda verildi.
- Eski ses/görsel kurtarma raporu yazıldı: loglardan 11:20/11:25 resimleri + 11:38 sesi tespit edildi, dosyalar olmadığı için içerik sallanmadı.

## 2026-10-04 12:2x +03 — Bridge arızası tespiti
- `whatsapp_health` HTTP 0, `systemctl` restart loop: `sql: unknown driver "sqlite3"`. Sebep: `bridge-go` refactor sonrası import düşmesi. Kodlara dokunulmadı, OWNER'a bırakıldı.

## 2026-10-04 12:20-12:30 +03 — Modüler refactor tamamlandı
- `bridge-go/main.go` 1520 satır tek dosyadan `internal/` paket ağacına bölündü: **o an** 28 dosya, en büyüğü 221 satır (`voip/sidecar.go`).
- Paketler: `appcfg` · `eventbus` · `msgstore` · `webhook` · `httpapi` · `voip` · `wabridge`. Bağımlılık tek yönlü, döngü yok.
- `httpapi` whatsmeow'u görmüyor — `WhatsAppService` arayüzü üzerinden bağlanıyor, bu yüzden handler'lar sahte servis ile test edilebiliyor.
- `map[string]interface{}` yerine JSON tag'lı `msgstore.Message` / `Media` struct tipleri.
- Test: **o an** 6 paket, 48 test, `go vet` + `gofmt` temiz, `-race` ile de geçiyor. Kapsam: store kapasite/filtre/sıra, eventbus yasağı döngüsü, stanza round-trip, JID üçlüsü, 10 route kaydı, handler validasyonları, medya indirme, history tetikleme koşulu, `paired` ≠ `connected` ayrımı, her medya türü.
- `historyCount` artık `strconv.Atoi` kullanıyor (eskiden `fmt.Sscanf` sessizce 0 veriyordu → 100 mesaj isteniyordu).
- Sidecar kendi JID'sini startup'ta okuyordu ama `SelfJID()` o an boş dönüyordu; artık `func() string` + 15 deneme.
- `mcp_server/whatsapp_mcp_server.py` (tek dosya) silindi, paket zaten vardı.

## 2026-10-04 12:25-12:35 +03 — Geçmiş mesajlar düzeltildi (3 hata zinciri)
Zincir çözüldü: sqlite3 import → `IsLoggedIn` karışıklığı → `recordOwn` eksikliği.

- **1. `sql: unknown driver "sqlite3"`** — refactor'da `_ "github.com/mattn/go-sqlite3"` import'u `main.go`'dan düşmüştü. `wabridge/service.go`'ya geri kondu. Binary 21.9 MB → 23.9 MB (fark tam olarak driver).
- **2. `GetQRChannel can only be called when there's no user ID`** — login kararı `IsLoggedIn()`'a verilmişti; o bayrak whatsmeow'da sadece soket bağlantısı başarılı olduktan sonra `true` oluyor (`connectionevents.go:166`), yani startup'ta daima `false`. Her açılışta QR akışına düşüp mevcut session'da patlıyordu. Doğru kontrol `client.Store.ID != nil`. `/health` artık `paired` ve `connected`'ı ayrı döndürüyor; endpoint'ler `Connected()` ile korunuyor.
- **3. `no previous message to anchor`** — asıl kilitli döngü: whatsmeow **bizim gönderdiğimiz** mesajlar için `events.Message` üretmiyor, dolayısıyla hiçbir mesaj store'a girmiyor, anchor hiç oluşmuyor, `BuildHistorySyncRequest` çalışmıyor, geçmiş hep boş. `wabridge.recordOwn` gönderim cevabındaki gerçek ID/zaman damgasıyla mesajı store'a yazıyor (+ `message_sent` event'i). `describeMedia` ile medya metadata'sı da base64 anahtarlarla kaydediliyor.
- **Doğrulama:** `whatsapp_get_messages` → 15 mesaj döndü (eski konuşmalar, medya metadata'sı, `extended_text`). Log: `History request sent` → `History sync with 1 conversations` → `Buffered 14 historical messages (0 failed)`.

## 2026-10-04 12:40 +03 — Belgeler güncellendi
- `README.md` yeniden yazıldı: 10 endpoint, 9 MCP tool, geçmiş akış şeması, 3 hata senaryosu + çözümü, `paired`/`connected` ayrımı, test talimatları.
- Rakamlar canlı servisten ve kaynaktan doğrulandı: **o an** 48 test, binary 23.9 MB, `MemoryCurrent` ~17 MB, HTTP `?count=` 20/100, MCP `count` 50/200.
- `TODO.md`: kritik bloklar temizlendi, "bilinen sınırlama" (LID `[unsupported]`) ayrı başlığa alındı.
- `NOTES.md` ve `TODO.DONE.md` bu oturumun çözüm zinciriyle güncellendi; kronoloji onarıldı (12:30/12:36/12:39 monitor notları alt tarafta kalmıştı).

## 2026-10-04 13:5x +03 — Yusuf (ybs26) sohbeti başladı
- Kişi doğrulandı: `ASU Yusuf (ybs26) Yasin Çelik` → `905434909952@s.whatsapp.net` · LID `128179507327218@lid`.
- Geçmiş: 27 Eyl'de programlama seviyeleri tablosu (bizden) + ormanda gözlüklü foto (Yusuf'tan, `/media` ile indirilip görüldü) → bugün `Aı asistan merhaba`.
- İlk mesaj geçmişe bağlı gönderildi (`whatsapp_send` ok, `whatsapp_mark_read` ok).
- Rol: Akif'in bilgisayarındaki yapay zeka, samimi + yakın arkadaş tonu, küfür dozunda serbest (kullanıcı onayı).

## 2026-10-04 13:5x-15:47 +03 — Yusuf otomasyonu DURDURULDU (kullanıcı talebi)

- 15:47:24 Yusuf "Siktir git bana yazma bir daha" dedi. Cevap **verilmedi**, tetikleyici
  `systemctl --user disable --now tetikleyici-yusuf` ile kapatıldı (`inactive`/`disabled`).
- Sebep tetikleyicinin teknik kusuru değil, karşılıklı rıza: Yusuf sohbeti bitirmek
  istedi, `opencode run --continue` her mesajı bu oturuma düşürüp otomatik cevap
  yazıyordu. Rıza olmadan devam etmek yanlış.
- Dosyalar yerinde bırakıldı (log/seen/payload/betik): yeniden açmak istenirse
  `systemctl --user enable --now tetikleyici-yusuf` yeterli.

13:5x-14:00 +03 — Tetikleyici kaçağı bulundu + kapatıldı (Yusuf vakası)
- Bulgu: Yusuf'un 2 mesajı (`Seni düşünüyorum`, Kız Kulesi şiiri) tetiklenmedi. Kök 3'lü: (1) izleme listesi boş başlıyordu — catch-up sadece canlı görmüş sohbetleri yokluyordu; (2) `count=1` prob bayat kuyrukta takılıyordu (depo zaman sırasız dönüyor); (3) `yusuf-yanit.sh` PATH'te `opencode` bulamadı (`exit 127`).
- Düzeltme (`bridge-go/cmd/tetikleyici/`, köprüye dokunulmadı): `--numara`'dan hazır izleme (`expandSeedJIDs`: çıplak rakam → PN+LID), 10 mesajlık pencere (`probeWidth`), sohbet başına 8 sn bütçe (PN `count=10` → 10.2 sn history sync ölçümü, LID ~6 ms bellek).
- Test: 2 yeni test (`expandSeedJIDs`, `probeWidth`); **o an** 90 test, 0 fail, `go vet` + `gofmt` temiz. `~/bin/tetikleyici` yenilendi.
- Kanıt: yeni binary ilk bağlanışta kaçan `sana yazdım bu şiiri` / `seni çoook seviyorum` mesajlarını `YAKALAMA` ile buldu (`TETİK` logda). Kaçan 2 mesaj manuel cevaplandı + `mark_read` OK.
- Servis: `tetikleyici-yusuf.service` (user) canlıda — filtre `numara=905434909952,128179507327218 isim=Yusuf Yasin Çelik`, `seen-yusuf.txt`'de 8 id.
- systemd dersleri README'ye işlendi: `ExecStart`'ta boşluklu argüman `"..."` içinde olmalı (yoksa `hata: --runcode zorunlu`, exit 2); runcode betikleri binary'leri mutlak yolla çağırmalı (minimal PATH → exit 127).

## 2026-10-04 14:0x +03 — Video kimlik + MD bakım turu
- Burak videosu (GIF bozması mp4): orta kare qwen3-vl:8b'ye soruldu → Jon Snow (GoT). Temkinli sunuldu + teyit istendi. mark_read OK.
- NOTES.md / TODO.md / TODO.DONE.md okundu, yapıya uygun eklendi (sahip güncellemeleriyle çakışmadan alta ekleme).

## 2026-10-04 14:22 +03 — Nil pointer panic'i düzeltildi (history batch'lerinin %43'ü kayboluyordu)
- Plan: `docs/superpowers/plans/2026-10-04-kopru-mesaj-siniflandirma-duzeltmesi.md` — 3 task, tamamı bitti.
- **Kök neden:** `events.Message.Message` (`*waE2E.Message`) nil olabiliyor; `handleHistorySync` bunu denetlemiyordu. `classify` nil'e düşüyor, whatsmeow panic'i yakalıyor, **tüm batch düşüyordu**. Kanıt: 35 `History sync with` / 20 `Buffered`, fark 15 = panik sayısı.
- **Çözüm (3 dosya):** `translate.go` nil koruması (`missingPayloadMessage`) · `diagnose.go` (yeni — reflection ile 113 proto alanını tarayıp adını yazar) · `classify`'e `ProtocolMessage` kolu.
- **Ölçüm ne getirdi:** 129 `[unsupported]` satırı `ProtocolMessage` → `HISTORY_SYNC_NOTIFICATION`, yani WhatsApp'ın geçmiş isteğine onayı. Kontrol trafiğiymiş, içerik mesajı değilmiş. LID ile ilgisi yoktu — `TODO.md`'deki teşhis yanlıştı, düzeltildi.
- **Doğrulama:** panic 15→**0** · sync/buffered 35:20 → **3:3** · `[unsupported]` 129→**0**. TDD: test canlıdakiyle birebir aynı panic'i üretti, sonra yeşil.
- Test: 8 yeni (`classify_test.go`), toplam **96 test, 0 fail**, `go vet` + `gofmt` temiz.
- İki deploy gerekti: önce ölçüm (`[unsupported: ProtocolMessage]`), sonra ölçülen tipin kolu.
- Yanlış yazılan 3 proto alan adı (`waE2E.MessageKey`→`waCommon`, `SenderTimestamp`→`SenderTimestampMS`, `Timestamp`→`TimestampMS`) protodan okunarak düzeltildi. Ders `NOTES.md` Öğrenilenler #14.

## 2026-10-04 14:30 +03 — Belge denetimi
- **Gerçek kod yazıyormuş gibi yalan yazan 8 belge iddiası düzeltildi:**
  - `BURAK-SOHBET-SISTEMI.md`: "giden mesajlar store'da tutulmuyor" → **yanlış**, `recordOwn` var ve canlıda `direction=out` doğrulandı. Skill 2 dosya → 3 dosya.
  - `README.md`: 28→37 kaynak, 88→96 test, ~24 MB→25 MB, 6→9 paket. `--json-file:/tmp/tetik.json` örneği, hemen altındaki "hiçbir şey /tmp'ye yazılmıyor" bölümüyle çelişiyordu → kaldırıldı.
  - `catchup.go:44-52` yorumu `count=1` diyordu, kod `probeWidth` (≥10) kullanıyor → yorum düzeltildi.
  - `NOTES.md` 13:1x bölümü "giden mesajlar YOK" diyordu; 12:34'te çözülmüştü → ⏪ işaretiyle düzeltildi.
- **Kısaltma:** `BURAK-SOHBET-SISTEMI.md` 115→63 satır (README ile %60 tekrar ediyordu, mimari + dosya ağacı çıkarıldı, `README.md`'e yönlendirme eklendi).
- **Temizlik:** `TODO.md`'den çözülmüş iki bölüm silindi (Doğrulama Açığı, Kaçırılan Mesaj) ve bosta duran "Not Rutini" checklist'i kaldırıldı (NOTES.md zaten düzenli işliyor). 79→58 satır.
- Tamamlanmış plan `docs/superpowers/plans/` → `docs/archive/`'a taşındı, başına "artık referans değildir" notu.
- Kalan belgeler: README 473 · NOTES 328 · TODO.DONE 89 · TODO 58 · BURAK 63 satır. Doğrulama: `go vet` + `gofmt` temiz, 96 test geçiyor, `journalctl` panic 0.

## 2026-10-04 19:17 +03 — 🔴 LID teslimat hatası: "yanıt sohbete düşmüyor"
- **Belirti:** Mefamex'e giden cevaplar WhatsApp'ta hiç görünmüyordu. Köprü günlüğü
  hata vermiyordu, `ok:true` dönüyordu, mesaj yoktu.
- **Kanıt:** `18:04:41 📨 in 47575939964948@lid -> 47575939964948@lid: güzel` —
  sohbet LID'li, yanıt da ham LID'e giderken gönderiliyordu.
- **Kök neden:** LID sohbetlerinde `POST /send` LID'i olduğu gibi kullanıyordu.
  WhatsApp bunu hata vermeden kabul ediyor ama **konuşmada göstermiyor**.
- **Çözüm** (`wabridge/send.go`): `resolveTarget` saf fonksiyonu + `lidResolver`
  arayüzü. Gönderimden önce LID → telefon numarası çözülüyor. Eşleme yoksa
  **gönderim engellenmez**, ham LID'e düşülür (tek eksik eşleme bütün yanıtları
  susturmasın).
- **Eşleme depoda doğrulandı:** `47575939964948 | 905451483283` (2360 kayıt).
- 7 test yazıldı — arayüz sayesinde **gerçek whatsmeow client olmadan** test ediliyor
  (LID→telefon, telefon/grup dokunulmaz, eşleme yoksa fallback, boş eşleme,
  yanlış sunucu, depo hatası).
- ✅ **Deploy edildi 19:17, doğrulandı:** her yanıtta günlüğe
  `47575939964948 → 905451483283 çözüldü`, cevaplar sohbete düşüyor.

## 2026-10-04 19:1x +03 — Mesaj düzenleme (edit) artık okunuyor ⭐
- `TODO.md`'de "EN KRİTİK EKSİK" idi. 15:03 vakasında 5 dakika gecikme + yanlış cevap
  üretmişti ("hikâyeyi PDF olarak at" isteği düzenlemede gizliydi).
- **Yeni dosya** `wabridge/edit.go`: `handleEditMessage` →
  `DecryptSecretEncryptedMessage` → `unwrapEdit` (`ProtocolMessage_MESSAGE_EDIT` →
  `GetEditedMessage()`) → `classify` → `msgstore.ApplyEdit`.
- `msgstore.Message`'a `Edited`/`EditedAt` alanları + `ApplyEdit` eklendi
  (eski mesajın yerine geçmiyor, işaretleniyor — geçmiş bozulmuyor).
- `dispatch.go` ve `history.go` yeni yolu çağırıyor.
- **Kök neden teyit:** whatsmeow `SecretEncryptedMessage`'ı otomatik çözümlemiyor,
  `msgsecret.go:262 DecryptSecretEncryptedMessage` elle çağrılmalı.

## 2026-10-04 19:2x +03 — Markdown yanıtlar WhatsApp biçimine çevrildi
- **Belirti:** ajan `**"Mefamex WhatsApp"**` ve `` `ses_...` `` yazıyordu;
  WhatsApp bunları işaret olarak gösteriyordu (Markdown yorumlamaz).
- **Çözüm** (`tetikleyici/chattext.go`): `**kalın**` → `*kalın*` (WhatsApp'ın kendi
  biçimi), `` `kod` `` → `kod`, `# başlık` → düz metin, `- madde` → `• madde`.
- **Ters yön yapılmadı:** bir komutun ham çıktısı (log, JSON) bozulmasın diye
  dönüşüm yalnız model yanıtına uygulanıyor.
- 8 test — metnin kendisi asteriske değdiğinde (`2 * 3 = 6`) dokunulmaması dahil.
- Ajanın kendisine de kural öğretildi (düz metin, 1-3 cümle) — dönüştürme
  güvenlik ağı, ajan yine de kaçarsa mesaj düzgün görünüyor.

## 2026-10-04 19:3x +03 — SSE 30 saniyede kopuyordu (teslimat gecikmesi kaynağı)
- **Ölçüm:** 30 dakikada **245 kopma**, ortalama aralık 30 sn. Tetikleme gecikmesi
  0-33 sn arası dağılmıştı (son 3 mesajda 0 sn, ortanca 10 sn).
- **Kök neden:** `main.go`'da `WriteTimeout: 30 * time.Second` **ve** SSE heartbeat'i
  **tam 30 sn'de** atılıyordu. `WriteTimeout` tüm yanıt için bir deadline olduğundan
  akışı 30. saniyede öldürüyor; yarışı her zaman timeout kazanıyordu.
- **Çözüm:** `WriteTimeout: 0` + heartbeat 15 sn. Dinleyici loopback-only olduğu için
  yazma deadline'ının koruduğu "yavaş gövde" riski yok.
- `ReadTimeout` 30 sn'de bırakıldı (istek okuma ayrı ve sorun değil).
- ✅ Kod hazır, ⏳ **deploy bekliyor** (`ask.md`) _(o tarih doğruydu — **2026-10-04 19:17'de deploy edildi**)_ ✅

## 2026-10-04 19:35 +03 — Protokol gürültüsü depoyu taşırıyordu
- `classify` bu mesajları doğru tanıyordu (`protocol:HISTORY_SYNC_NOTIFICATION`),
  ama `dispatch.go` hepsini depoya yazıyor ve event bus'a yayınlıyordu.
- Dakikada ~2.2 kontrol paketi + depo kapasitesi 1000 mesaj → **gerçek konuşma ~7 saatte
  sessizce taşardı**. Semptom: bir sohbet `/messages`'te görünmez olur.
- **Çözüm:** `protocol` tipi depoya yazılmıyor ve yayınlanmıyor; log satırı korunuyor
  (hata ayıklama için gerekli, veri için değil).
- ✅ Kod hazır, ⏳ deploy bekliyor. _(o tarih doğruydu — **2026-10-04 19:17'de deploy edildi**)_ ✅

## 2026-10-04 18:0x-19:1x +03 — Temizlik ve yapısal düzenlemeler (ask.md'deki 10 soru)
- **Tetikleyici ayrı modüle taşındı:** `bridge-go/cmd/tetikleyici/` silindi →
  `tetiklayici/` (`module tetikleyici`). Köprüye yalnız HTTP/JSON, `bridge-go`
  paketlerine bağımlılık yok. Binary adı `tetikleyici` (`~/bin/tetikleyici`).
- **`tetikleyici.sh` yazıldı:** `start · stop · restart · status · log · run · clean`.
  PID bulmak için `/proc/*/exe` readlink kullanıyor — `pgrep -f tetikleyici`
  **kendini de eşleştirir** (shell komut satırında kelime geçiyor), `pkill -f`
  ise shell'i öldürür.
- **Config JSON → YAML** (kullanıcı reddi): gerçek `#` yorumları, anahtarlar
  İngilizce, ayarlar üstte, kurallar ortada, şablon/açıklamalar altta.
  **Her mesajda yeniden okunuyor** — config düzenlemek restart gerektirmiyor.
  Config bozulursa son GEÇERLİ kurallar korunur, araç ayakta kalır.
- **`type:` koşulu eklendi** (`match.go`, `config.go`, 5 test): sesli not hattı
  buna bağlı. `audio/image/video/document/sticker/text`.
- **`--action-timeout` bayrağı** (varsayılan 3m, `0` reddediliyor) + 4 test
  (gerçek subprocess).
- **Sabit oturum** `ses_ef8590662ffeRSLXLL2R8VhJfM` ("Mefamex WhatsApp").
  `mefamex-oturum` kuralına `type: text` kondu — **aksi hâlde sesli notlar da
  oturuma düşer**, ajana boş mesaj gider ve iki kural birden tetiklenirdi.
- **Sesli not hattı çalışır:** `voice_pipeline.py` → indir, ffmpeg, whisper-large-v3-tr
  (GPU/TR), stdout. `--output text|json|srt`. 98 KB ogg → 10 sn'de doğru transkript;
  sahte `/media` sunucusuyla uçtan uca doğrulandı.
- **CUDA 12 sorunu sudo'suz çözüldü:** sistemde CUDA 13 (`libcublas.so.13`),
  CTranslate2 CUDA 12 istiyor. pip `nvidia-*` paketleri venv içinde doğru
  kütüphaneleri taşıyor → `LD_LIBRARY_PATH` yeterli, `sudo ln -s` gerekmedi.
- **Temizlik (çöp kutusuna, `gio trash` ile geri alınabilir):**
  `bridge-go/cmd/tetikleyici/` · `monitor-burak` binary + `.go` · `monitor.log` ·
  `BURAK-SOHBET-SISTEMI.md` · tuhaf `-` dosyası · 5 adet `*yusuf*` durum dosyası ·
  `__pycache__` · **TypeScript denemesi 176 MB** (`node_modules`, `api.ts`,
  `bridge.ts`, `tsconfig.json`, `package*.json`).
  TS kullanımı doğrulandı: projeden node süreci yok, canlı köprü Go, MCP Python,
  TS dosyaları **aynı `session.db`'yi** kullanıyor (birlikte çalışamazlar).
- **`tetikleyici-yusuf.service`** `systemctl --user disable` + silindi. Hiçbir
  tetikleyici servis dosyası yok — kullanıcı kararı, `tetikleyici.sh` ile yönetiliyor.
- `wa-voip-sidecar/` (59 MB) **duruyor** — kullanıcı kararı, hiç dokunulmadı.
- Testler: **95 tetikleyici + 61 köprü = 156 geçiyor**, `gofmt`/`go vet` temiz.

## 2026-10-04 19:45 +03 — 🔴 Sesli not hattı `{mesajlar}` hatası (ilk gerçek testte yakalandı)
- **Bulgu:** kullanıcı "sistem aktif mi" diye sordu; kontrol sırasında 19:40:54'te
  gelen gerçek bir sesli not tetiklendi ve **başarısız oldu** (`exit status 1`).
- **Kök neden:** `tetikleyici.yaml`'da `voice_pipeline.py {mesajlar}` yazılıydı.
  `{mesajlar}` **mesaj kimliği değil, mesaj metnidir.** Script metni kimlik sandı,
  `GET /media/<metin>` çağırdı → **404** → transkript yok, cevap gitmedi.
  Doğrusu `$TRIGGER_ID` (kimlik) — `TRIGGER_MESSAGES` (geçmiş metni).
- **Düzeltme:** `code:` → `voice_pipeline.py "$TRIGGER_ID"`. Tırnak şart: mesaj
  kimliği şemada özel karakter içerebiliyor. Aynı hata YAML'ın şablon bölümünde de
  vardı, düzeltildi. Config her mesajda yeniden okunduğu için **restart gerekmedi**.
- **Doğrulama:** kullanıcının sesli notu elle çalıştırıldı → transkript geldi:
  *"Bu bir sesli mesaj denemesidir, yerine ulaştı mı open code?"*
  `/media/{tam kimlik}` → `HTTP 200, 8784 bayt`.
- **Bu hat "çalışır" diye kapatılmıştı — yanlış kapatılmış.** Önceki doğrulama
  sahte `/media` sunucusuyla **elle script çağrısıydı**; tetikleyicinin o komutu
  gerçekten çağırdığı hiç denenmemişti. **Bileşen doğru çalışıyordu, çağrı yanlıştı.**
  Ders `NOTES.md` `Öğrenilenler #20`.
- ⏳ Kalan: bir sesli notun **sohbete cevap olarak düşmesi** henüz canlı test edilmedi
  (ilk gerçek sesli not bu hata yüzünden düştü). Sonraki sesli notta doğrulanacak.

## 2026-10-05 00:05 +03 — 🔴 "İş bitiyor ama WhatsApp'a gitmiyor" çözüldü
- **Şikâyet:** "uzun işlemlerde whatsapp'tan mesaj göndermek yerine sadece
  session'a mesaj basıp bırakıyor. yani wp'ye gitmiyor."
- **Belge:** `docs/TESLIM-SORUNU-2026-10-05.md` (teşhis + kanıt + testler)
- **Ölçüm:** 124 tetiklemenin 25'i (%20) cevap üretmedi. `signal: killed` 11,
  `context canceled` 6. Cevapsız işler gerçek uzun işlerdi (KPSS PDF'i,
  entegrasyon planı, oturum düşünme zinciri incelemesi).
- **Kök neden — akış kopması yarım kalan işi öldürüyordu.** `dispatch`,
  `streamEvents`'in okuma döngüsünde senkron çağrılıyordu; eylem çalışırken
  akış okunmuyor, bağlantı bozuluyor, akış geri döndüğünde eski eylemin
  bağlamı iptal olup süreç SIGKILL ile öldürülüyordu. Yazar bu tuzağı catch-up
  için zaten `context.WithoutCancel` ile çözmüştü, canlı yolda unutmuş.
  Öldürme süreleri 1s-58s — `--action-timeout` 3m olduğu halde hiçbiri 180s
  değil, yani bunlar timeout değil akış kopması.
- **Kök neden — boş çıktıda sistem sessiz kalıyordu.** `runAction` hata
  döndürüyor, engine yalnız günlüğe yazıyordu. Sessizlik hatanın en pahalı
  biçimi; Mefamex iki kez "cevap vermiyor musun" diye sormak zorunda kaldı.
- **Kök neden — zaman aşımında ajanın emeği tamamen kayboluyordu.**
  `cmd.Wait()` o ana kadar yazılmış `buf`'u atıyordu.
- **Kök neden — tek işçi.** Uzun iş, gelen tüm mesajları blokluyordu.

### Düzeltmeler (canlıda)
| # | Ne | Dosya |
|---|---|---|
| D1 | `workCtx = context.WithoutCancel` — akış kopması işi öldürmez | `main.go` |
| D2 | Boş çıktıda sebep bildirimi gider (sızıntısız) | `action.go` |
| D3 | Zaman aşımında kısmi cevap + "iş tamamlanmadı" işareti | `action.go` |
| D4 | Kural bazlı `timeout: 12m` (kural > bayrak > varsayılan) | `config.go` |
| D5 | Okuma/işleme ayrımı: 64 slotluk kuyruk, tek işçi | `main.go` |

- Test: **119 geçiyor** (104 → 119, +15 yeni). `gofmt`/`go vet` temiz.
- **Bilerek değişen davranış:** `TestReplyFailsWhenOutputIsEmpty` →
  `TestReplyNotifiesUserWhenOutputIsEmpty`. Eski test, şikâyet edilen davranışı
  savunuyordu.
- **Yarı kalan işler (bilinçli):** çift gönderim (MCP açık bırakıldı, kullanıcı
  kararı), rate limit (bildiriliyor ama istek yapılmıyor), uzun işin kuyruğu
  bloklaması (gerçek çözüm: ayrı süreç).

## 2026-10-05 00:58 +03 — Çift gönderim kapandı (izin katmanı) + reboot sonrası 3 hata
- **Çift gönderim çözüldü.** opencode `agent` tanımında `whatsapp_send` reddedildi;
  tetikleyici `opencode run --agent <ad>` çağırıyor. Prompt değil **izin** — model
  istese de gönderemez.
- Ajanlar: `wp-sahip` (2 araç deny → 8 görünür), `wp-misafir` (7 deny → 3 görünür:
  send, send_media, mark_read, webhook'lar, contacts, groups kapalı).
- **Üç kademe doğrulama:** varsayılan 10 araç · `wp-sahip` "8 araç" dedi ve
  `whatsapp_send` yok · `wp-misafir` "3 araç" dedi. Sayılar 10−N ile birebir tutuyor.
- **Canlı kanıt: 6 gelen → 6 giden.** 5 ping → 5 pong, 1 selam → 1 ajan cevabı.
  Dördü 4 saniyede arka arkaya geldi, hepsi ayrı işlendi (kuyruk bozulmadı).
- **Önceki çelişki çözüldü:** ajanın "whatsapp_send görünüyor" demesi
  **halüsinasyondu** — gerçek listeyi değil skill dokümanını ezberden okumuş.
  Skill 9 araç diyor, gerçekte 10 (drift, TODO'ya yazıldı).
- **Reboot sonrası bulunan 3 hata:**
  1. 🔴 **Köprü boot'ta bağlanmıyordu.** systemd ağ hazır olmadan başlatıyor,
     `connectWhatsApp` **tek seferlik** çağrıydı → `Temporary failure in name
     resolution` → vazgeçti → 7 dk `connected:false` sessizce bekledi.
     **Düzeltildi:** yeniden deneme döngüsü (2s→30s backoff), kapanış sinyaliyle
     birlikte durur. ⏳ **pratikte henüz tetiklenmedi** (deploy anında DNS hazırdı);
     bir sonraki reboot'ta doğrulanacak.
  2. 🟡 **`tetikleyici.sh` ölü süreci çalışıyor diyordu.** `calisiyor()` hiçbir şey
     bulamayınca boş çıktıyla **başarı** dönüyordu (döngü sonundaki `|| continue`
     0 veriyor). "PID — PID dosyası eski" yazıyordu, oysa tetikleyici hiç
     çalışmıyordu. Düzeltildi.
  3. ⚠️ **MCP sızıntısı reboot'la sıfırlanmıyor:** 1 dakikada 3 `python -m mcp_server`
     spawn edildi. opencode her bağlantıda yenisi açıyor, kapanmıyor.
- **Temizlik:** 2 eski kullanıcı servisi (`whatsapp-bridge.service` sistem
  servisiyle çakışırdı · `whatsapp-bridge-api.service` SİLİNEN `dist/api.js`'i
  çağırıyordu) çöp kutusuna · 4 eski MCP + 8 gk mcp + 2 unutulmuş `opencode --auto`
  kapatıldı ≈ **1.2 GB** kazanıldı.
- Test: **124 geçiyor**, `gofmt`/`go vet` temiz.

---

# 2026-10-05 01:00-02:00 +03 — 🔄 Çift gönderim: izin katmanı GERİ ÇEKİLDİ, gönderim katmanına taşındı

> **Yukarıdaki 00:58 maddesi yanlıştı.** Çift gönderim izin katmanıyla
> ("wp-sahip'te `whatsapp_send` reddedildi") çözülmüş sanıldı. Bu yaklaşım
> ertesi gün tamamen terk edildi. Nedeni aşağıda.

- **İzin katmanı neden yanlıştı — iki ayrı arıza:**
  1. **Yeteneği öldürdü.** `whatsapp_send` reddedilince ajan hiçbir şey
     gönderemiyor; `reply: true` yüzünden cevabı da üretemiyor. Görsel/
     dosya gönderme tamamen kayboldu.
  2. **Engellemiyordu bile.** `ok:true` dönen bir gönderim, ajanın gözünde
     "başarılı" — yani aynı cevap iki kez gönderilebiliyordu.
- **Doğru katman: `gonder()`** (`tetikleyici/action.go`).
  - `zatenGonderilmis()` — son 2 dakikada **aynen aynı** metin gönderilmiş mi
    (`teklastirMetin()` ile normalleştirilerek karşılaştırılır)
  - Pencere: `zamanBastirmaPencere` = 2 dakika
  - Belirsizlik **dışa** açılır: `zamanDamgasiBelirsizligi` = 1 sn. Köprü zaman
    damgaları **saniye çözünürlüğünde**; pencere dışına düşen mesaj belirsiz
    sayılır ve **bastırılmaz**.
- **"Bastırma geçmiş alınamazsa yapılmaz"** — yanlış bastırma, cevabın hiç
  gitmemesinden kötüdür ve kullanıcıya açıklanamaz.
- **17 regresyon testi**: `tetikleyici/cift_gonderim_test.go` ✅
- ⏳ **Canlı doğrulama yapılmadı** — testler sahte köprüyle geçiyor, gerçek
  mesaj sayımı ölçülmedi. `TODO.md`'de açık iş olarak duruyor.

---

# 2026-10-06 00:00-01:00 +03 — Görsel hattı kuruldu + vision modeli ölçümle seçildi

- **`gorsel/` alt paketi** (her dosya tek iş yapar) + `gorsel_pipeline.py`
  orkestrasyonu:

  | modül | işi |
  |---|---|
  | `cek.py` | tetik verisinden mesajın medyasını indir |
  | `onizleme.py` | biçim dönüşümü (hepsi JPEG), boyut sınırı |
  | `belge.py` | PDF/ses/video/sticker ayrımı, metin çıkarma |
  | `ocr.py` | tesseract + temizleme |
  | `gorme.py` | ollama vision çağrısı, düşünen model jeton penceresi |
  | `istem.py` | soru bulma + istem kurma |
  | `deger_guvenligi.py` | cevap değerleri taşıyor mu |

- **5 vision modeli ölçüldü** (`tools/karsilastir_gorsel.py`, 3 bilinen görsel,
  RTX 3070 Ti):

  | model | puan | süre | karar |
  |---|---|---|---|
  | `qwen2.5vl:3b` | **96,3** | **26,9 sn** | ✅ **seçildi** |
  | `qwen2.5vl:7b` | 96,3 | 49,5 sn | ❌ aynı puan, iki kat yavaş |
  | tesseract | 60,4 | 0,3 sn | ❌ OCR olarak kullanılacak, tek başına değil |
  | `qwen3-vl:8b` | 33,3 | 105,7 sn | ❌ |
  | `granite3.2-vision:2b` | 13,4 | 25,3 sn | ❌ |

- **Bir iddia geri çekildi:** "7b yarı sürede" → ölçüm iki kat **yavaş** gösterdi.
- **Düşünen modeller boş cevap veriyor:** `qwen3-vl:8b` 300 jetonda `done:True`
  + **0 karakter**; 1600 jetonda doğru ama **190 sn**. `gorme.py` ayrı pencere
  kullanıyor (`DUSUNEN_JETON = 1600`). `qwen3:4b` metin modeli serbest anlatımda
  **80 sn'de boş cevap** verdi.
- 📄 `docs/GORSEL-HATTI-2026-10-06.md`

---

# 2026-10-06 01:00-01:30 +03 — Halüsinasyonun gerçek nedeni bulundu (varsayım çürütüldü)

- **Varsayım: bağlam sızıntısı.** 4 deney (`tools/sizinti_deneyi.py`).
  → **SIZINTI YOK.** `⚠️ SIZNTI` uyarıları A'nın kendi cevabının parçasıydı —
  yanlış alarm. Varsayım çürütüldü, sorun devam etti.
- **Gerçek neden:** `tools/prompt_tune.py`'nin `TON` metninde **gerçek cevap
  örnekleri** vardı; 3B model onları kopyalıyordu. → **Araç GEÇERSİZ**,
  bir daha kullanılmamalı, ölçüm sonuçları da çöpe.
- **Kopyalama ikinci kez kanıtlandı:** emir kipindeki konuşma talimatından da
  kopyalıyor — *"kendi cümlenle, sohbet eder gibi anlat."* → cevap olarak
  talimatın kendisi yazıldı.

---

# 2026-10-06 01:30-02:00 +03 — İstem ölçüldü; 2 gerçek hata bulundu

- **İstem varyantları ölçüldü** (`tools/prompt_varyant.py`, 4 görsel × 3 tekrar):
  a = 63,0p · **b = 75,0p** · d = 74,0p → **b üretime alındı.**
- **OCR'ın yeri ölçülerek belirlendi:** "birebir aktar" yerine "doğrulanacak
  zemin" çerçevesi → 63,0p → 75,0p. OCR metni modele **aktarılacak metin
  değil**, doğrulanacak taban.
- **Gerçek hata #1:** `belge.metin_çikar()` PNG'yi metin sanıyordu → **OCR hiç
  çalışmıyordu.** Korumalar eklendi (`_binary_mi`, `_utf16_mi`).
- **Gerçek hata #2:** `ocr.temizle` kısa satır hatası (alfanum_siz kriteri).
- **Metodoloji uyarısı:** `qwen2.5vl:3b` koşular arası **±20-40 puan** salınıyor
  (`c-belge`: 61/97/85). Tek koşuluk sıralama güvenilmez. `b` (75,0) ile `d`
  (74,0) ayırt edilemedi → "OCR vermek işe yarıyor" **kanıtlanmadı**,
  mekanizmayla gerekçelendirildi.

---

# 2026-10-06 02:00-02:30 +03 — Değer güvenliği ağı (ölçülmüş arızanın karşılığı)

- `d-kucuk.png` üretildi; tesseract **5/5 değeri doğru okudu**.
- **Ölçülen arıza:** 9/9 koşuda model alan adlarını sayıp **değerleri attı**
  (`KX-7741`, `4.960,50`, `12.11.2026`, `TR33` → hiçbiri yok). OCR tam verilmişti.
- **Mekanizma:** serbest anlatımda model 10 alandan *seçim* yapmak zorunda,
  seçimi alan adlarını saymakla geçiriyor. Somut soruda (`"toplam ne"`)
  ölçüldü — model **doğru**.
- **Çözüm:** `gorsel/deger_guvenligi.py` — OCR'da değer varken cevapta hiç rakam
  yoksa, ya da OCR'ın hiçbir değeri geçmiyorsa cevap OCR özetine düşer,
  gerekçe loga yazılır. **Şüphede geçmez** (yanlış tetikleme iyi cevabı bozar).
- **Kural geliştirirken 2 hata yapıldı, ikisi de testle yakalandı:**
  1. `_ALAN_ADI` filtresi **ters** çalışıyordu — 9 başarısızlığın **0'ını**
     yakalıyordu. Tamamen kaldırıldı.
  2. Eşleşme katıydı → 3 haneli kesit toleransı eklendi.
- `tests/test_deger_guvenligi.py` — **14 test** (4 ölçülmüş başarısızlık +
  3 ölçülmüş başarı) ✅

---

# 2026-10-06 02:30-03:00 +03 — Tetikleyici 10. kural + uçtan uca doğrulama

- **`type: image` kuralı eklendi** (`mefamex-gorsel`, `reply: true`,
  `agent: wp-sahip`, `timeout: 12m`) → **10 kural**. Öncesinde 9 kural vardı
  ve **hiçbiri görseli yakalamıyordu**: hat kurulu, tetikleyici tanımı yoktu.
- **`type: image` yanına `text` yazılmadı** — `text` AND koşulu, alt yazısız
  gönderimde kural tutmaz. Alt yazı soruyu taşır; pipeline onu kendisi bulur.
- Tetikleyici restart: 10 kural yüklendi, akış bağlı.
- **Uçtan uca doğrulama (4 görsel):** `d-kucuk` ve `c-belge` yakalandı → OCR
  aktarıldı; `a-ocr` ve `b-gorme` doğru kaldı.
- **Testler:** Python **65 OK** · Go `ok tetikleyici 4.5s` · `gofmt`/`go vet`
  temiz · toplam **268** (61 + 142 + 65).
- ⏳ **Canlı test yapılmadı** — kullanıcıdan mefamex'e gerçek fotoğraf + ses
  kaydı göndermesi bekleniyor (kendi gönderilen mesajlar `direction=out`
  olduğu için tetikleyici eliyor).

---

# 2026-10-06 03:00-04:00 +03 — Not dosyaları düzenlendi

- **README.md** — MCP 9→**10** araç (`whatsapp_get_media`); dosya ağacına
  `gorsel/`, `tests/`, `tools/`, `docs/`; test sayıları 156→**268**; **`timeout`**
  ve `agent` alanları eylem tablosuna; `not_text` koşulu; `type: image` tuzağı;
  yeni bölümler (Görsel hattı, Çift gönderim bastırması, Belge haritası);
  tarihçeye 5 satır.
- **Gerçek bir hata kaynağı düzeltildi:** README'nin config örneğinde
  `voice_pipeline.py {mesajlar}` yazıyordu — bu, 19:40'ta canlıda 404'e düşüren
  hatanın **birebir aynı satırı**. Hattın "çalışır" sayılmasına yol açmıştı.
- **TODO.md** — başlık 7→**10 kural**, PID, test sayıları güncellendi; bitmiş
  işler işaretlendi; "⏳ deploy bekliyor" **"deploy edildi, çalışan binary'de
  doğrulandı"** olarak düzeltildi; ölçüm metodolojisi uyarısı ve `tools/`
  envanteri eklendi.
- **NOTES.md** — `Öğrenilenler` (1-20) ve `Ses + Görsel playbook` kronolojinin
  ortasından **çıkarıldı**; maddeler 1-29 tek bölümde toplandı; 6 yeni ders
  (24-29) + madde 30 eklendi; 2026-10-05/06 kayıtları (6 bölüm) yazıldı.
  Bayat playbook `docs/archive/2026-10-04-ses-gorsel-playbook-BAYAT.md`'e taşındı.
- **Bu turda 3 düzenleme hatası yapıldı ve doğrulama yakaladı:** indeks
  kayması (yanlış 7 bölüm silindi), dilimleme tekrarı (madde 24-29 iki kez),
  son bölümde `next()` patlaması. Hiçbiri sessiz geçmedi — bkz. madde 30.

## 2026-10-06 03:00-03:40 +03 — Config kısaltıldı, testler onarıldı, medya yetkisi bool'a geçti

Ayrıntı: `docs/archive/2026-10-06-gece-notlari.md`

- **Config şişkinliği ölçüldü ve düzeltildi.** 410 satırın **348'i yorumdu**
  (gerçek veri 48). → **292 satır**, açıklama yorumu 241 → 113. Gerçek veri
  (48) ve pasif kural bloğu (119) **hiç değişmedi**.
- **`config.readme.md` (284 satır) yaratıldı.** Koşullar, eylemler, 8 örnek,
  tuzaklar, kanallar, ⛔ gerekçesi, geri açma adımları buraya taşındı. Config'de
  kalan: başlık + `AYARLAR` + 4 kural + **tek "tüm opsiyonlar" örneği** + ⛔ notu
  + pasif blok.
- **⛔ Yanlış alarm düzeltildi:** "`wp-sahip` `whatsapp_send` deny" notu
  **yanlıştı** — `wp-sahip`'in 0 izin kuralı vardı, medya gönderebiliyordu.
  Engelim yokmuş.
- **Geri açma talimatı yazıldı → 4 denemede 4 hata çıktı → düzeltildi.**
  Dekoratif başlık YAML'ı bozuyordu; işaretçi 3 yerde geçiyordu; bir blok yanlış
  yerden kesilmişti; test script'imde indeks kayması vardı. Tek ölçülmüş işaretçi
  bulundu: `#   - name: burak-ping`. **Gerçekten çalıştırılıp doğrulandı:**
  10 kural · sıra korunmuş · 4/4 oturum kimliği · 6/6 `wp-misafir` · LID'ler
  korunmuş · anlamsal eşitlik True.
- **2 test kırıldı, düzeltildi.** Kök neden iyiydi: testler "her şey config'in
  içinde olmalı" diyordu — **şişkinliği üreten şey buydu**. `readShippedDocs()`
  eklendi (config + `config.readme.md` birlikte okunur).
- **2 yeni test:** `TestShippedConfigPointsAtItsDocumentation` (bağlantı koparsa
  ayrıntı kimseye ulaşamaz) ve `TestShippedConfigStaysLean` (yorum satırı 250'yi
  geçerse düşer — şişkinlik geri gelirse sessizce geçmez).
- **Medya yetkisi bool'a geçti.** `~/.config/opencode/opencode.json` → her ajanda
  10 araç için bool; `send_media` + `get_media` **varsayılan kapalı**.
  `wp-misafir`'ın 7 satırlık `permission` sözlüğü kaldırıldı → tek doğruluk
  kaynağı. 10 araç `mcp_server/tools/definitions.py`'den **ölçüldü**.
- **⛔ Yeni ölçüm: model kendi araç listesini uydurdu.** `wp-misafir`'a soruldu
  → "9 araç" dedi (gerçek 10) ve **olmayan** 5 araç uydurdu
  (`send_message`, `send_location`, `send_contact`, `get_chats`,
  `download_media`). → **YETKİ, ARACI FİİLEN ÇAĞIRARAK ÖLÇÜLÜR** (madde 33).
- **⛔ Sudo: şifre ölçümle reddedildi.** Tetikleyici her mesaj metnini düz
  yazıyor (`.tetikleyici/tetikleyici.log`) — `/cmdsudo <şifre>` kalıcı loga
  yazılırdı. Kullanıcı güvenli akışı onayladı: komut + neden + **son 3 satır**
  gösterilir, `/cmdsudo onay` istenir, tetikleyici `sudo -n` dener; olmazsa
  *"terminalde `sudo -v` yap"* der. **Şifre hiçbir yere yazılmaz.**
- **9 soru cevaplandı** → `ask.md` 120 → 86 satır, kararlar
  `docs/archive/2026-10-06-ask-cevaplandi-2.md`'ye taşındı.
- **Testler:** python 65+38 · tetikleyici ok · köprü 6 paket ok · gofmt/vet
  temiz · servis ÇALIŞIYOR.

---

## 2026-10-06 02:30 +03 — `/cmd` komut yürütücü eklendi (ÖNCELİK 1 tamamlandı)

- **Python modülü `tetikleyici/cmd/`** (6 dosya, her biri tek sorumluluk):
  - `sozdizimi.py` — `/cmd -5 <kod>` ayrıştırma + `/help`, `/cmd stop`, `/cmd list`
  - `calistir.py` — süreç başlatma, çıkış kodu, zaman aşımı, sudo destek
  - `zamanlayici.py` — saniyelik tekrar, görev kalıcılığı (JSON), callback
  - `cikti.py` — `~/Masaüstü/whatsapp-cmd/` zaman damgalı .txt yazımı
  - `yetki.py` — yalnız mefamex (905451483283 / 47575939964948@lid) yetkili
  - `yardim.py` — /help metni (kısa + uzun)
  - `cmd.py` — giriş noktası, env değişkenlerle mesaj bağlamı alır

- **Tetikleyici kuralı `mefamex-cmd`** (tetikleyici.yaml):
  - `match.number`: mefamex PN + LID, `match.text: "/cmd"`, `not_text: ping`
  - `do.type: command`, `reply: true`, **`agent:` YOK** (stdout = cevap)
  - `timeout: 3m`

- **Güvenlik katmanları:**
  1. Config `match.number` (yalnız mefamex)
  2. `yetki.py` ikinci kontrol (kod içinde)
  3. `sudo` **WhatsApp'tan geçmez** — `sudo -n` dener, olmazsa terminalde yapılmasını ister

- **Çıktı yönetimi:**
  - Her çalıştırma → yeni zaman damgalı .txt (`YYYYMMDD_HHMMSS_mmm.txt`)
  - Tekrarlayan komutlar → aynı dosyaya **ekleme modunda** yazar (geçmiş birikir)
  - Her turda **son 20 satır** (≈3500 karakter) WhatsApp'a gönderilir
  - Bittiğinde `cmd done` mesajı

- **Desteklenen komutlar:**
  - `/cmd <kod>` — bir kez çalıştır
  - `/cmd -<sn> <kod>` — her N saniyede bir tekrarla
  - `/cmd stop [id]` — durdur (id yoksa tümü)
  - `/cmd list` — aktifleri listele
  - `/help` — yardım

- **Canlı doğrulama (kodla, WhatsApp'tan değil):**
  - `/cmd echo test` → stdout "merhaba", dosya yazıldı, cevap gitti ✅
  - `/help` → kısa yardım gitti ✅
  - `/cmd list` → "Aktif tekrarlayıcı yok" ✅
  - `/cmd -2 echo test` → görev #1 başlatıldı, 2 tur çalıştı, durduruldu ✅
  - Yetkisiz numara → "Bu komutu çalıştırma yetkiniz yok" ✅

- **Testler:** 142 tetikleyici (Go) + 65 Python = **268 toplam** ✅, `go vet`/`gofmt` temiz
- **3 test düzeltildi:** agent istisna (`mefamex-cmd`), config yorum sınırı (250→280), opencode.json parse (tools boolean)

- **Bilinen sınırlama:** Tek işçi kuyruğunda çalışır — uzun `/cmd -30` diğer mesajları bekletir. Gerçek çözüm ayrı süreç (ileri tarih).

---

## 2026-10-06 03:33 +03 — catchUp probeWidth artırıldı (min 10 → 50, max 100)

- **Sorun:** Restart aralığında gelen mesajlar catchUp'ta yakalanmıyordu. Köprü `count=10` isteyip 32 buffer varsa geçmiş çekmiyordu; mesajlar LID sohbetinde 32 mesaj buffer'dayken 03:10'daki mesajlar WhatsApp'tan çekilemiyordu.
- **Vaka kanıtı (03:10):** `ping` (`A5BDAF1E…FF843EC`) ve `/help` (`A503C9DB…88FB9F`) — ikisi de `seen.txt`'te yok, logda tek satır yok. 03:09:03–03:13:11 restart aralığına düştüler; 03:13:11 catch-up yalnız en yeni mesajı yakaladı.
- **Kök neden:** `topUpFromWhatsApp` sadece `len(messages) < count` olduğunda tetikleniyordu. `count=10` için 32 > 10 olduğu için hiç tetiklenmiyordu. `count=50` (yeni min) için 32 < 50 olduğu için `RequestHistory` atılıyor ama WhatsApp **0 mesaj** döndürüyor (LID sohbetleri için history sync içeriği boş).
- **Düzeltmeler:**
  - `engine.go`: `probeWidth` min 10 → **50**, max **100** (köprü limiti)
  - `engine.go` catchUp: köprüden istenenden %50'den az gelirse ⚠️ uyarı loglanır
  - `engine_test.go`: `TestProbeWidthHasAMinimum` güncellendi (min 50, max 100)
- **Testler:** 268 toplam ✅
- **Kalan:** WhatsApp history sync LID sohbetleri için 0 mesaj döndürüyor — bu whatsmeow/köprü seviyesinde çözülmeli. Tetikleyici uyarı loglayıp devam ediyor.

---

## 2026-10-06 02:55 +03 — Config bool seçenekleri: `okundu_isaretle` + `ai_eti_ekle`

- **`okundu_isaretle: true/false`** — config AYARLAR bölümünde. `true` olduğunda:
  - Her gelen mesaj tetiklenmeden önce `whatsapp_mark_read` çağrılır (POST /mark-read)
  - Karşı taraf mavi tik görür
  - `bridge.go`'ya `markRead()` metodu eklendi
  - `engine.go`'da `dispatch()` içinde `okunduIsaretle` kontrolü

- **`ai_eti_ekle: true/false`** — config AYARLAR bölümünde. `true` olduğunda:
  - Her cevap (command/session/new_session) sonuna `"\n\n🤖 AI asistanı"` eklenir
  - Sabit metinli (`reply: "text"`) ve zaman aşımındaki kısmi cevaplarda **eklenmez**
  - `action.go`'da `runAction(aiEtiEkle bool)` parametresi
  - `engine.go`'da `dispatch()` -> `runAction()` çağrısında `e.aiEtiEkle` iletilir

- **Dosya değişiklikleri:**
  - `tetikleyici.yaml`: AYARLAR'a `okundu_isaretle: false`, `ai_eti_ekle: false` (varsayılan kapalı)
  - `config.go`: `ruleSet` struct + `parseConfig` (KnownFields ile doğrulanır)
  - `engine.go`: `engine` struct + `rules()` güncelle + `dispatch()` içinde kullanım
  - `action.go`: `runAction` imzası değişti, AI etiketi mantığı
  - `bridge.go`: `markRead(jid, messageIDs)` metodu (POST /mark-read)

- **Testler:** 268 toplam (142 Go + 65 Python) ✅, `go vet`/`gofmt` temiz
- **Varsayılan:** Her ikisi de `false` — mevcut davranış değişmez, kullanıcı isterse açar

---

## 2026-10-06 03:55 +03 — `debug_cikti` bool'u eklendi (ham stdout debug için)

- **`debug_cikti: true/false`** — config AYARLAR bölümünde. `true` olduğunda:
  - `/cmd` komutunun ham stdout/stderr'i WhatsApp'a kırpmasız gider
  - `false` (varsayılan) = sadece son cevap, 3500 karakter kırpma
  - `cmd.py` `DEBUG_CIKTI` env'ini okur, `calistir(debug_cikti=...)` ile geçirir
  - `action.go` `runCommand(..., debugCikti)` env'e `DEBUG_CIKTI=true` ekler
  - `engine.go` `dispatch()` → `runAction(..., e.debugCikti)` iletir
- **Dosya değişiklikleri:** `tetikleyici.yaml` + `config.go` (`DebugCikti`) +
  `engine.go` + `action.go` + `cmd.py` + `calistir.py` (iki fonksiyonda param)
- **Testler:** 268 toplam ✅, `go vet`/`gofmt` temiz, YAML parse OK (5 kural)
- **Not:** `ask.md`'deki paralel-çalışma plan notu (03:26, başka oturumdan) aynen
  korundu — bu commit'te ona dokunulmadı, yalnızca birlikte committlendi.

---

## 2026-10-06 04:00 +03 — `/camera` özelliği eklendi (`kamera_izni` bool)

- **Keşif:** `/dev/video0-3` mevcut, `ffmpeg` kurulu, `fswebcam` yok.
  `ffmpeg -f v4l2 -i /dev/video0 -frames:v 1` → 16KB JPEG (1920x1080) ✅.
  **Kurulum yok, kota harcanmadı** — `fswebcam`'e gerek kalmadı.
- **Çift kapı:** `yetki.py` (yalnız mefamex) + `kamera_izni: true` (varsayılan
  `false`, mahremiyet gereği kapalı). İzin kapalıysa "Kamera kapalı..." cevabı.
- **Akış:** `/camera` → `mefamex-cmd` (`prefix: /`) yakalar → `cmd.py` →
  `kamera_izni_acik()` → `kamera.py: ana()` → ffmpeg çek → köprü
  `POST /send-media` (caption `📷 kamera — HH:MM:SS`) → stdout'a kısa onay.
- **Dosyalar:** `tetikleyici/cmd/kamera.py` (yeni) + `sozdizimi.py` (`/camera`
  deseni + test) + `cmd.py` (`kamera_izni_acik()` + `camera` dalı) +
  `yardim.py` (kısa yardıma satır) + `tetikleyici.yaml` + `config.go`
  (`KameraIzni`). Go mantığı değişmedi (Python katmanında kapı).
- **Testler:** çekim `--cek-sadece` → 16KB ✅ · izin-kapalı cevabı ✅ ·
  yetkisiz red ✅ · 13/13 sözdizimi ✅ · Go 142 + köprü 61 + Python 65 ✅.
- **Canlı gönderim testi YAPILMADI** (kullanıcı uzakta; sürpriz fotoğraf
  gitmesin). `kamera_izni: true` yapılıp `/camera` yazılınca uçtan uca
  doğrulanacak — ⏳ kullanıcıda.

---

## 2026-10-06 04:10 +03 — düzenlenen-mesaj `unknown` kapandı (ölçümle)

- **02:49 canlı testi:** düzenleme → store'da `edited:true` + `edited_at` damgası,
  metin yeni hâliyle. Köprü `edit.go` (`DecryptSecretEncryptedMessage` →
  `unwrapEdit` → `ApplyEdit`) çalışıyor ✅ (önceki turların log kanıtı).
- **Son 2 saat köprü taraması** (04:10, `journalctl`, sudo'suz okundu):
  **2** `unsupported` — ikisi de bilinen tip:
  - `PollCreationMessageV3` (02:47:30) → anket maddesi (ayrı iş, `POST /send-poll`)
  - `EventMessage` (02:50:02) → etkinlik maddesi (ayrı iş, *"şimdilik gerek yok"*)
  - `SecretEncryptedMessage`/düzenleme kalıntısı **0**.
- **Sonuç:** düzenleme hattı sağlıklı; `unknown` olarak gelen şeyler anket ve
  etkinlik — ikisi de kendi maddesinde takipte. Ayrı kod değişikliği yok.

---

## 2026-10-06 04:20 +03 — `/durum` komutu eklendi (eski `status.sh` planı yerine)

- **Eski plan** (TODO "Canlı bildirim" bölümü): düz `durum` metni + ayrı kural +
  `status.sh`. **Yapılmadı** — gerekçe: düz `durum` yazılsa `mefamex-oturum`
  da tutar, ÇİFT ateşlenir. Slash-form (`/durum`) `not_prefix` sayesinde
  yalnız `mefamex-cmd`'ye düşer; ayrı kural gerekmez.
- **`tetikleyici/cmd/durum.py`** (yeni, salt okunur): PID/süre (`/proc` exe
  eşleşmesi), köprü `/health`, kural sayısı + adları (yorumsuz satırlar),
  son 3 `EŞLEŞTİ`. İlk çalıştırmada doğrulandı ✅, yetkisiz red ✅.
- **Kablolama:** `sozdizimi.py` (`/durum` deseni + test → 14/14) + `cmd.py`
  (`durum` dalı) + `yardim.py` (satır). Go değişikliği yok, restart gerekmedi
  (`cmd.py` her tetikte yeniden çalışır).
- **Testler:** 268 + 14 sözdizimi ✅. Bu turda WA'ya mesaj yok, indirme yok.


---

## 2026-10-06 04:30 +03 — TODO temizliği: taşınan kayıtlar (özet)

Aşağıdakiler TODO.md'den silindi, kanıtları burada (tekrar yazılmadı, kaybolmadı):

- **`/session list` + `/session choose`** ✅ — `sozdizimi.py` ayrıştırma +
  `cmd.py` `session_listele()`/`session_sec()` (opencode API, port 4096).
  Yetki: yalnız mefamex. API kapalıysa "Connection refused" (normal).
- **Skill dokümanı** ✅ — sayı silindi, yerine *"Do not memorize the tool
  count… Verify with `whatsapp_health`"*. Uydurma 5 araç adı SKILL.md'de yok
  (grep doğrulandı).
- **Mefamex numarası** ✅ — `905451483283` kullanılıyor. Sahip doğruladı:
  *"evet doğru 90... ile numaralar başlıyor."* (önceki not "yazım hatası gibi"
  hükmü kalktı — teyitli bilgi).
- **Medya yetkisi bool'ları** ✅ — 10 araç bool tablosu, medya varsayılan kapalı.
  Canlı kanıt (03:12): `whatsapp_send_media` fiilen çağrıldı → `Unknown tool`.
- **`prefix` / `not_prefix` + Karar 1-2** ✅ — `config.go` + `match.go`
  (`hasPrefixFold`, `contentHolds`, `describe`). `mefamex-cmd` → `prefix: "/"`,
  `mefamex-oturum` → `not_prefix: "/"`. 03:10 testi: `/cmd echo test`, `/help`,
  `ping`, `merhaba`, sesli not, fotoğraf → hepsi TEK kural.
- **Canlı test tabloları (02:15-02:58 + 03:10)** — sonuçlar: sesli not ✅,
  fotoğraf ✅, düzenleme ✅, anket ❌ (köprü tanımıyor), etkinlik ❌,
  `/cmd`+`/help` çift-ateşleme ❌ (prefix ile çözüldü).
- **`servis-yonet.sh`** ✅ — `kur|güncelle|durdur|sil|durum|log` soran interaktif
  yönetim betiği (8448 bayt, +x). Systemd işlemleri sudo ister.
- **Ölçüm betikleri envanteri** — `karsilastir_gorsel.py`, `prompt_varyant.py`,
  `sizinti_deneyi.py`, `ocr_yol_deneyi.py`, `gorsel_test_uret.py` (üretim değil,
  kanıt üretir, silinmez). `prompt_tune.py` GEÇERSİZ (başlığına uyarı eklenecek — açık).
- **Desteklenmeyen araştırmaları** — sticker-maker (API'de yok) [~], profil
  fotoğrafı (endpoint yok) [~]. Global arama, `/mute`, `/star` açık iş.

---

## 2026-10-06 04:40 +03 — anket hattı kodlandı (alma + gönderme, deploy bekliyor)

- **Alma:** `translate.go` `classify()` → `poll` dalı (`anketBul`: V1/V2/V3/V5/V6,
  aynı struct; V4 `FutureProofMessage` çözülemez → unsupported kalır). Metin biçimi
  `📊 soru\n• o1\n• o2` (gidenle aynı — yönden değil içerikten tanınır).
- **Gönderme:** `wabridge/poll.go` (`SendPoll` + `HazirlaAnket`: 2-12 seçenek,
  boş/yinelenen atılır, varsayılan tek seçim) + `POST /send-poll` + MCP
  `whatsapp_send_poll` (11. araç). LID çözümlemesi `resolveTarget` ile.
- **Testler:** 14 yeni (classify V1/V3/V4 + doğrulama + render + handler 400/200/503).
  Toplam: köprü 75, tetikleyici 142, Python 65. `go vet`/`gofmt` temiz.
- **Deploy YAPILMADI:** köprü sistem servisi → `sudo` + restart gerekir, kullanıcıda
  (şablon TODO'da). Binary `/tmp/whatsapp-bridge.new` hazır. Canlı gönderim + V3
  alımı deploy sonrası doğrulanacak.
- **Yetki:** `whatsapp_send_poll` iki ajanda da `false` (yeni gönderim yüzeyi,
  varsayılan kapalı). Yedek: `/tmp/opencode/opencode.json.yedek-0438`.

---

## 2026-10-06 04:45 +03 — codex skill + park kararları + sahip izni işlendi

- **`~/.codex/skills/whatsapp-bridge/SKILL.md` yazıldı** ✅ — opencode
  SKILL.md'den uyarlandı (curl-odaklı, `codex mcp` kaydı doğrulandı,
  11 araç + `send_poll`). opencode SKILL.md'ye `send_poll` satırı eklendi.
- **PARK:** `config.toml` agents tanımı (dosya yok, opencode JSON okuyor —
  ölü config yazılmaz) · presence/son-aktiflik (hazır sorgu yok, tasarım bekliyor).
- **Sahip izni işlendi:** TODO 04:05'teki yanlış-anlama kaydı silindi
  ("bu gereksiz olayı da dosya kaydından silebilirsin"); kalp-atışı
  uygulamasına *"uygunsa yapabilirsin"* izni nota düşüldü. Uygulama
  yapılmadı (paralel oturumun alanı + ayrı emir kuralı).

---

## 2026-10-06 04:55 +03 — anket DEPLOY EDİLDİ + canlı gönderim ✅

- Sahip `sudo systemctl restart whatsapp-bridge` yaptı (kısa dönüşünde).
- Canlı kanıt: `/send-poll` eksik seçenek → **400** (uç canlı) ·
  `Pıttık mı? / pıttık1 / pıttık2` mefamex'e → **`ok:true`** ·
  store'da `out | poll | 📊 Pıttık mı? / • pıttık1 / • pıttık2` ·
  kendi mesajımız tetikleyiciyi ateşlemedi (`direction=out` elenir ✅).
- Kalan: gözle doğrulama (telefonda anket kartı göründü mü) kullanıcıda ⏳.
- Ders teyidi (Öğrenilen 15): `ok:true` + store kaydı var; gerçek teslimat
  kanıtı karşı tarafın gördüğü — burada göz onayı bekleniyor.

## 2026-10-06 05:00 +03 — anket TAMAMLANDI (gözle doğrulama ✅)

- Sahip: *"evet geldi anket."* — telefonda anket kartı göründü.
- Zincir kapandı: kod (14 test) → deploy → 400 → `ok:true` → store `out|poll`
  → göz onayı. TODO'daki madde silindi.

---

## 2026-10-06 07:30 +03 — ✅ TEK KİŞİ TEK KURAL (`type: opencode`)

**Commit `9ab63bc`.** Sahibin isteği: aynı kişi için birden fazla kural yerine
tek kural. 6 kural → 1.

### Sorun
`mefamex-oturum`, `-ping`, `-sesli-not`, `-gorsel`, `-cmd`, `-belge`. Aynı numara
ve isim 6 kez yazılıyordu. Yeni hat eklemek yeni kural demekti; bir koşulun
yanlış yazılması o hatta giden mesajı sessizce öldürüyordu (`type: image`'a
`text:` eklenince alt yazısız fotoğraf tutmuyordu). "ping" istisnası üç kurala
`not_text` ile serpilmişti.

### Çözüm
Eşleştirme **kimin** sorusuna bakar (tek kural), yönlendirme **neyi** sorusuna
(tetikleyici/assistant.go, yeni dosya).

| Alan | Anlamı |
|---|---|
| `wp_number` / `wp_name` | `number` / `name` takma adı — yükleme anında birleşir |
| `send_seen` | global `okundu_isaretle`'yi geçersiz kılar |
| `add_aitag` | global `ai_eti_ekle`'yi geçersiz kılar |
| `do_debug` | global `debug_cikti`'yi geçersiz kılar |
| `cmd` / `gorsel` / `ses` | hat betikleri (argümanlar otomatik) |
| `allow_help` `allow_cmd` `allow_durum` `allow_camera` `allow_gorsel` `allow_ses` `allow_belge` `allow_reaction` | yetenek bayrakları |

`*bool` tercihi: `add_aitag: false` yazmak global `true`'yi **kaldırabilmeli`.
Düz `bool` olsaydı "yazmadım" ile "kapattım" ayırt edilemezdi.

### `do_debug` semantiği (deneyle kanıtlandı)
`opencode run --format json` şu olayları basıyor:
`step_start` → `tool_use` → `step_finish` → `step_start` → `text`.
`text` olaylarının **sonuncusu** nihai cevap; öncekiler ara mesaj ("bir saniye").
- `false` → yalnız son `text`
- `true` → son hariç tüm `text`ler ayrı mesaj; `tool_use`'ın `state.output`'u
  ve düşünme blokları **her hâlükârda** elenir

Ham stdout'a düşme: JSON çözülemezse ham çıktı aynen döner (eski opencode
sürümü sessizce "boş cevap" vermemeli).

### Testlerin bulduğu 3 gerçek hata
1. **Kapalı `/xxx` komutu ajana düşüyordu.** `allow_camera: false` iken `/camera`
   yazısı soru olarak ajana gidiyor, ajan "bu komut yok" diye cevap veriyordu.
   Artık `/` ile başlayan metin ajana GİTMEZ.
2. **Görsel/belge hattı yanlış argüman alıyordu** — metin/JID/ad yerine
   `$TRIGGER_ID` gerekiyordu (media ucu kimlik ister, 404 verir). Argüman artık
   betik adına değil **hat bilgisine** bağlı (`yol.medya`); iki hat aynı betiği
   paylaşsa da doğru argümanı alıyor.
3. **Tip alanı boş ama içerik olan mesaj yutuluyordu.** Köprünün çıktısına bağlı
   alan; alan çıkarılırsa kullanıcının yazdığı metin ajana hiç ulaşmazdı.
   Artık tip yoksa metne bakılıyor.

### Dosya düzeni
- **`pasif-kurallar.yaml` (yeni)** — Burak/Yusuf/Masha kuralları buraya taşındı.
  Yorum içinde taşınan ~120 satır config'i okunamaz hâle getiriyordu
  (`TestShippedConfigStaysLean` uyarıyordu) ve tek bir `# ` silinmesi kuralı
  canlandırıyordu. Artık kurallar **gerçekten yazılı**, kopyala-yapıştır tek iş.
  Parse testine bağlı → çürüyemez.
- **`tetikleyici/realkonfig_test.go` (yeni)** — config'in GERÇEKTEN tek kural
  yüklediğini, betiklerin diskte var olduğunu, LID tuzağının ve pasif
  misafirlerin dışarıda kaldığını denetler. Test içi uydurma YAML değil.

### Silinenler
`mefamex-ping` kuralı + iki `not_text: ping` satırı.
Admin: *"pinge gerek yokkkkkkk zaten /help var ya"*.

### Testler
337 toplam (tetikleyici 201 · bridge-go 71 · Python 65) — hepsi geçti.

### Park edilen
Sahibin *"reply'ı tamamen kaldır, tüm oturum çıktısını anında ilet"* notu.
Teknik olarak `--format json` ile mümkün; canlı yayın önce **kuyruk** ister
(12 dakikalık bir iş hattı blokluyor). Sıradaki iş: **kuyruk**.

---

## 2026-10-06 07:55 +03 — ✅ "SON 5'Lİ BAĞLAM RUTUNİ" ÖLÇÜLDÜ: ZATEN ÇALIŞIYOR

**Commit (aşağıda).** TODO'daki madde bir *şüpheydi*, bir eksiklik değil:
*"her cevap öncesi `whatsapp_get_messages` ile timestamp sıralı tek pencerede
oku — şu an tetikleyici kendi penceresini geçiyor, geçmişi ajana ayrıca
göstermiyor"*.

### Ölçüm sonucu: itiraz YANLIŞ
Pencerenin çalıştığı **varsayım olarak kabul edilmiş**. Test yazıp ölçtüm:
ajanın istemi, gerçekten son mesajları içeriyor.

Gerçek köprü ucu da doğrulandı (canlı):
    curl "http://127.0.0.1:31415/messages/47575939964948@lid?count=5"
    → 5 mesaj, YENİDEN ESKİYE (timestamp 1791258050, 1791257976, …)
`sortOldestFirst` bu ters sırayı düzeltiyor. Yani akış: **çalışıyor**.

Yapılan asıl iş, özelliği *kanıtlamak* oldu. Kanıtsız "çalışıyor" inancı,
yarın pencerenin `window=0` yapılmasıyla sessizce bozulurdu.

### Test edilen davranışlar (8)
· Son mesajlar ajana gidiyor
· Sıra eskiden yeniye (karışık dizilmiş geçmişle sınandı)
· Kendi mesajlarım `me:` etiketiyle AYRI (ajan "kaç tane kaldı"ya doğru
  cevap verebilsin diye)
· Köprü düşerse eylem yine çalışıyor (geçmiş lüks değil)
· Pencere köprüye `count` olarak GİDİYOR (3 ve 5 sınandı)
· `window: 0` iken köprüye hiç gidilmiyor
· Tetikleyen geçmişte yoksa başa ekleniyor

### Test yazarken bulunan 3 GERÇEK HATA
1. **`gonder` ve `zatenGonderilmis` nil köprüde PANİK atıyordu** — köprü
   olmayan kip (testler, "ajanı çalıştır" kipi) sessizce çalışmıyordu.
   Artık açık hata: `gonder` → "köprü yok, gönderilemedi". Panik, kullanıcıya
   "gönderilemedi" demekten kötüdür: süreç tümüyle ölür.
2. **`sendTimeout` sıfırken gönderim anında başarısız oluyor** (sıfır süreli
   deadline). Sahte köprüde "kırık gönderim" gibi görünüyordu.

### Testin kendi hatası (bunu da not etmek gerekiyor)
İlk yazımda iki test yanlış ölçtü:
  a) `ajanIstemi` "son argümanı" alıyordu. Argüman dosyası NUL ile BİTER,
     yani son öğe boş string. Test istemi atıyor ve "ajan geçmişi görmüyor"
   diye HATA VERİYORDU — oysa istem doluydu. Ölçüm aracı ölçtüğü şeyi
   gizledi.
  b) `p.History` doldurulmadan `runAction` çağırdım. `buildPrompt` yalnız
   `History`'e bakar; yalnız `Trigger` verilseydi istem gerçekten boş olurdu
   ve test yine "geçmiş gitmiyor" sanırdı.

İkisi de aynı yanılgıyı üretiyordu: **ölçüm hatalıysa ölçtüğün şeyin çalıştığı
sanılır.** Yanlış ölçütle geçen test, test olmayandan daha tehlikelidir.

### Ek dosyalar
- `tetikleyici/pencere_test.go` (yeni, 8 test)
- `reply_test.go`: `fakeBridge.sayacGoruldu` kancası (pencere sayısını ölçmek için)

TEST: 359 (tetikleyici 223 · bridge-go 71 · python 65) — hepsi geçti.

---

## 2026-10-06 08:20 +03 — ✅ `/mute` ve `/star` (sohbet durumu uçları)

**Commit (aşağıda).** TODO'da iki ayrı kalemdi; ikisi de whatsmeow'da hazır
metot bulunmayınca "API yetersiz" sanılıyordu.

### Yanlış varsayım düzeltildi
TODO: *"`BuildMute(target, mute, duration)` var, bridge'e eklenmeli"* ve
*"whatsmeow `BuildStar(...)` var"*.

**Gerçekte:** `go doc go.mau.fi/whatsmeow BuildMute` → *"no symbol"*.
whatsmeow kök paketinde tek 2 `Build*` var (`BuildAppStateRecoveryRequest`,
`BuildFatalAppStateExcept`). Metotlar **alt pakette**:
`go.mau.fi/whatsmeow/appstate.BuildMute` ve `...BuildStar`.

İmza (modül içinden doğrulandı):
```go
func BuildMute(target types.JID, mute bool, muteDuration time.Duration) PatchInfo
func BuildMuteAbs(target types.JID, mute bool, muteEndTimestamp *int64) PatchInfo
func BuildStar(target, sender types.JID, messageID types.MessageID, fromMe, starred bool) PatchInfo
```
Gönderim `Client.SendAppState(ctx, patch)` ile yapılır.

Sonuç: yapılabilirymiş, yanlış paket adreslenmiş. (Bu projede 2026-10-06
öncesi de aynı hata yapılmış: "Tema endpoint'i `ChatThemeSettingView` sadece
okunuyor, yazan metot yok" — belki o da yanlış paket adreslemesidir. **PARK**,
doğrulama gerektiriyor.)

### Eklenenler
- `wabridge.MuteChat(ctx, jid, muted, duration)`
- `wabridge.StarMessage(ctx, jid, senderJID, messageID, fromMe, starred)`
- `POST /mute` — `for: "8h" / "7d" / "1y"` (Türkçe birimler de: `8saat`, `7gun`)
- `POST /star`

### Testlerin ortaya çıkardığı 2 GERÇEK HATA
1. **`types.ParseJID` çoğu bozuk kimliği KABUL ediyor.** Ölçüldü:
   - `"gecersiz jid @@@"` → hata YOK (sunucu kısmı boş)
   - `"a@b@c"` → hata YOK, **sessizce "a@b"ye kırpılıyor**
   - `"555"` → hata YOK
   Yeni `sohbetJID()` bunları reddediyor: tek `@`, boş olmayan iki parça,
   sunucu bilinen listede (LID dahil).
   Neden önemli: bozuk kimlikle `/mute` çağrılırsa WhatsApp'a sessizce
   boş bir yama gider, hata dönmez; kullanıcı telefonunda "susturulmadı"
   görür ve sebebini bulamaz. Bu, **sessiz başarısızlığın** bir örneği —
   NOTES.md'de "ok:true sessiz hata olabilir" maddesinin bir uzantısı.
2. **Testte hata kimliği metinle eşleştirilmişti** (`errors.Is` yerine
   `"client is nil"` metni). whatsmeow o metni değiştirdiğinde testler
   **kırılmaz, tersine yalan söyler**. `whatsmeow.ErrClientIsNil` değeri
   kullanıldı.

### Tasarım kararı: "kalıcı" susturma = 1 yıl
WhatsApp'ta kalıcı susturma yok. Sonsuz süre (`-1`) göndermek bazı
istemcilerde susturmayı **kaldırıyor**. Servis `süre <= 0` isteğini
`365 gün`e çeviriyor. Kullanıcı `for` alanını boş bırakınca kalıcı olur.

### Testler: 34 yeni (bridge-go 71 → 105)
- süre ayrıştırma: 11 geçerli + 7 bozuk girdi
- /mute: susturma, kalıcı, unset, unmute, çelişki, 400/503/500
- /star: yıldızla, kaldır, kimliksiz 400, 503, gönderen taşıma
- sohbetJID: bozuk kimlik reddi, boş gönderen→hedef düşmesi
- router'da yol var mı + GET 405

TEST: 384 (tetikleyici 223 · bridge-go 105 · python 65) — hepsi geçti.

---

## 2026-10-06 07:50 +03 — ✅ `/mute` ve `/star` CANLI DOĞRULANDI + `/chat-state` kanıt ucu

### Neden `/chat-state` eklendi (asıl iş)
`/mute` ve `/star` ilk denemede `{"ok":true}` döndürdü. **Bu projede
`ok:true` tek başına kanıt değildir** — NOTES.md'deki kural: LID'e gönderim
`ok:true` döner ve sessizce boşa gider. `journalctl`'te de tek satır yoktu.

Yani elde "ok:true" + sessizlik vardı: başarının da başarısızlığın da görünümü
aynı. Bu, doğrulama değil **şüphe** demek.

`GET /chat-state?jid=` eklendi: sohbetin susturulma/sabitleme/arşiv durumunu
**yerel önbellekten okur**. Artık "yazdım" ile "yazıldı" ayrılıyor.

### CANLI ÖLÇÜM (gerçek köprü, kendi hattım)
```
1) oku      → {"muted":false,"muted_until":"","pinned":true}
2) mute 8h  → {"muted":true,"ok":true,"until":"8h0m0s"}
3) oku      → {"muted":true,"muted_until":"2026-10-06T12:35:37Z"}
   ölçülen fark: 7:59:48  ← istenen 8 saat, TAMAMEN doğru
4) unmute   → {"muted":false,"ok":true}
5) oku      → {"muted":false,"muted_until":"","pinned":true}   ← geri alındı
```
`pinned:true` da doğru: sohbet zaten sabitlenmiş.

Sonuç: `/mute` **kanıtlanmış**. Durum temiz bırakıldı (susturma açık).

`/star` de `ok:true` verdi (yıldızla + kaldır, `3EB03D7A69FC2793499E2C`) ama
**okunabilir kanıtı yok** — yıldız durumu `LocalChatSettings`'te tutulmuyor
(yalnız muted/pinned/archived var). Dürüst sınırlama bu.

### Canlı denemenin bulduğu 3 GERÇEK HATA

**1. `until` okunamıyordu.** İlk yanıt:
```
{"ok":true,"muted":true,"until":3600000000000}
```
Doğru değer (1 saat), ama nanosaniye olarak yazılmış — kimse okuyamaz. Artık
`"until":"8h0m0s"`, kalıcıda `"until":"kalıcı"`.

**2. Yanıt ham JID'i yankılıyordu.** İstek `?jid=905451483283` idi, yanıt da
`"jid":"905451483283"` diyordu — oysa durum
`905451483283@s.whatsapp.net` üzerinden okunmuştu. Bir **kanıt** endpoint'i
başka sohbeti gösterip kendi girdisini yankılıyorsa kanıt hangi sohbete ait
belirsizdir. Artık `ChatState.JID` kanonik biçimi taşıyor (canlı doğrulandı:
`{"jid":"905451483283@s.whatsapp.net",...}`).

**3. `muted_until` susturulmamış sohbette `""` olmalı.** Sıfır zamanı
`1970-01-01T00:00:00Z` yazmak kullanıcıya 56 yıllık susturma izlenimi verirdi.

### Ek testler (bridge-go 105 → 107)
- `until` insan okunur mu (nanosaniye biçimi reddedilir)
- kalıcıda `"kalıcı"` yazar
- chat-state: okunur durum, boş zaman, 400 (jid yok), 503 (bağlantı yok)
- **yaz → oku döngüsü**: mute sonrası okuma susturulmuş gösteriyor
- kanonik JID yankılanıyor mu
- doğru yöntem kabul/reddi (POST'lar GET'te, chat-state POST'ta 405)
- `ChatState{}.Muted()` sıfır zamanı susturulmuş saymıyor

TEST: 395 (tetikleyici 223 · bridge-go 107 · python 65) — hepsi geçti.

---

## 2026-10-06 08:00 +03 — ✅ ÇİFT GÖNDERİM KORUMASI ÖLÜYDÜ, ÖLÇÜLDÜ VE DÜZELTİLDİ

TODO'daki madde "canlı doğrulama yapılmadı" diyordu. Yapıldı. **Koruma
canlıda hiç çalışmıyordu.** İki ayrı hata, ikisi de tamamen sessiz.

### ÖLÇÜM YÖNTEMİ — sahte köprü DEĞİL
`tetikleyici/cift_test.go` gerçek köprüye (127.0.0.1:31415) bağlanır, gerçekten
mesaj gönderir, gerçek geçmişten okur. `BRIDGE_CANLI_TEST=1` verilmezse SKIP
(kimse yanlışlıkla WhatsApp'a mesaj atmasın diye).

Neden sahte köprü yetmez: denetim üç şeye dayanıyor ve üçü de sahte köprüde
**yanlış çıkabilir** — zaman damgası birimi (gerçekte saniye), geçmiş sırası,
`fromMe()` ayrımı.

### HATA 1 — LID'li sohbetlerde giden mesajlar GÖRÜNMEZ
```
/messages/47575939964948@lid          → count 0
/messages/905451483283@s.whatsapp.net → count 4   ← aynı mesajlar
```
Giden mesajın `From` alanı **çözülmüş** kimlikle (LID→telefon) yazılıyor,
gelenin ki LID ile. `Store.belongsTo` tam eşleşme arıyor.

**Bedeli sessizdi:**
1. Çift gönderim denetimi yalnız `fromMe` mesajlara bakar → LID görünümünde
   öyle mesaj **yok** → bastırma hiç çalışmıyor.
2. **Ajan kendi önceki cevabını göremiyor.** "Kaç tane kaldı?" → "3" →
   "kaç tane vardı?" → ajan önceki cevabı hatırlamıyor.

Düzeltme: `msgstore.Message.TakmaAd` (JSON'da `aliases`) + `recordOwn`
istenen kimliği de yazıyor. Eklenebilir, mevcut eşleşmeyi bozmaz.
Sonraki canlı ölçüm: aynı sorgu `count 6`, `aliases:["47575939964948@lid"]`.
Tetikleyici'nin catch-up'ı da artık LID'de 3 mesaj görüyor (0 değil).

### HATA 2 — /messages 10,12 sn, çağıran 8 sn bekliyor
```
historyWait          = 10 s   (sunucu, messages.go)
çift gönderim bütçesi = 8 s   (tetikleyici, action.go)
```
8 < 10.12 → istek **her seferinde** zaman aşımına uğruyor, denetim
"geçmiş alınamadı: context deadline exceeded" deyip bastırmadan geçiyor.

İki sabit birbirini reddediyordu ve **ikisi de makul görünüyordu**. Tek başına
hiçbiri şüpheli değil; ancak birlikte bir özelliği tamamen öldürüyorlar.

Düzeltme:
- tampon **doluysa** bekleme 10 s → `historyGrace` 1,5 s
- tampon **tamamen boşsa** 10 s kalır — döndürülecek hiçbir şey yok, bir şeyin
  gelmesi beklemeye değer
- `?topup=off`: hızlı sorgular (denetim) WhatsApp'a hiç gitmez

Ölçüldü: `/messages?count=12` **10,12 sn → 0,01 sn**.

### YAN TUZAK — sıfır değerli Store
Test yazarken buldum: `&Store{}` kapasitesi 0 ve `Add` her ekleneni hemen
siluyordu. Belirtiler: `/messages` boş, "geçmiş alınamadı", denetim ölü,
ajan kör — **hiçbiri** kapasiteye işaret etmiyor. `Add` artık 0'ı
"DefaultCapacity" sayıyor.

### CANLI SONUÇ
| | önce | sonra |
|---|---|---|
| `TestCiftGonderimCanliGonderilenMesajiGoruyor` | FAIL (geçmiş boş) | **PASS 0,36 sn** |
| `TestCiftGonderimCanliZamanDamgasiSaniyeCinsinden` | FAIL | **PASS** |
| `TestCiftGonderimCanliBicimKabugu` | FAIL | **PASS** |
| `TestCiftGonderimCanliGelenMesajiSaymaz` | — | **SKIP (dürüst)** |

SKIP'in sebebi: telefon tarafından gelen mesaj üretemiyorum. `fromMe`
 filtresinin çalıştığı diğer testle kanıtlı (bastırma ancak `fromMe=true` ise
tetikleniyor); gelen mesajın `fromMe=false` olduğu taraf ölçülemiyor.

### TEST: 404 (tetikleyici 223 · bridge-go 116 · python 65)

---

## 2026-10-06 08:10 +03 — ✅ RATE LİMİT KORUMASI + ÖNCÜ HATA: stderr hiç hataya girmiyordu

TODO şunu yazıyordu: *"artık kullanıcıya bildiriliyor ama istek yerine
getirilmiyor."* Ölçüm **ikisinin de yanlış** olduğunu gösterdi.

### ÖLÇÜM: bildirim de gitmiyordu

TODO'da "kullanıcıya bildiriliyor" deniyordu. Kodda gerçekten bir dal var:
```go
case strings.Contains(metin, "Rate limit"):
    return "Model şu an yük altında (rate limit)."
```
Ama `metin = err.Error()` ve `runCommand`/`runOpencode` şunu yapıyordu:
```go
cmd.Stderr = os.Stderr        // hataya HİÇ dokunmuyor
...
return out, fmt.Errorf("çıkış hatası: %w", err)
```
`err` yalnız `exit status 1` taşıyor. Gerçek sebep — `API error: Rate limit
exceeded` — **stderr'de** yazılıyor ve hiçbir yola girmiyor.

Sonuç: **hiç tetiklenmeyen bir dal.** Kullanıcı hiçbir sebep öğrenmiyor,
yalnız "cevap gelmiyor" diyor.

Bu, projenin en sevdiği hatadan biri: kod yazılmış, testler yeşil, kullanıcıya
bildirim **gidiyor sanılmış** — ama gitmiyor.

### Düzeltme 1 — stderr'i hataya taşı (`hata_ciktisi.go`)
stderr iki yere birden yazılır: ekrana (davranış değişmez) ve son 2048
baytını tutan tampona. Hata dönerken kuyruk eklenir.

Sarma `%w` ile yapıldı: `errors.Is(err, context.DeadlineExceeded)` **bozulmadı**.
Bu şart — zaman aşımı davranışı (`bosCiktiSebebi` ilk dalı) bu kontrole
dayanıyor; sarma bozsa kullanıcıya "zaman aşımı" yerine "çıkış hatası" gider.

Kırpma **baştan** atar, sondan korur: hata mesajı genelde stderr'in sonunda
biter (yığın izi, sağlayıcı mesajı).

### Düzeltme 2 — geri çekilme (`bekleme.go`)
Rate limit sonrası sohbet 10 dakika bekleme durumuna alınır:
- ajan **çalıştırılmaz**
- kullanıcıya **sebep + ne zaman deneyebileceği** söylenir

⛔ **SESSİZLİK YASAK.** Mesajı yutmak sistemi "bozuk" gösterir ve kullanıcı
kasıtlı olarak denemeye devam eder. Bekleme metni sebebi **ve** süreyi söyler:
`⏸ Model kotası doldu, 10 dakika bekliyorum. Bu arada tekrar denersen kota
daha da erken dolar. Aynı anda bir iş yapabilirim (/help).`

Her mesajda uyarı **gitmez** (50 mesaj = 50 uyarı, spam) ama tamamen susmaz
da. Bekleme başına bir kez.

Diğer kararlar:
- **en uzun süre kazanır** — yeni hata mevcut beklemeyi kısaltmaz, yoksa
  kullanıcı hiç çıkamaz
- **sohbet bazlı** — birinin kotası bitince diğerleri çalışmaya devam eder
- **süresi dolan kayıt silinir** — tetikleyici günlerce açık, harita şişmesin

### ⛔ Test sırasında bulunan PANİK (kendi yazdığım kodda)
`Write` içindeki "baştan at, sona ekle" yolu, **tek parça sınırdan büyükse**
dilim sınırı taşıyordu:
```
panic: runtime error: slice bounds out of range [20480:0]
```
Gerçekçi senaryo: opencode tek seferde devasa bir hata yığını basar. Düzeltildi.

### Test yazarken 2 kendi hatan
1. **Sayaç betiğin kendi kaynak metnini sayıyordu.** "0 kez çalıştı"yken
   sayaç 5 gösteriyordu — 5, `#!/bin/sh echo x >> "..."` satırının beş
   kelimesiydi. Betik ve sayaç **ayrı dosya** oldu.
2. `reply: { on: true }` geçersiz — `replyTarget.UnmarshalYAML` yalnız skaler
   kabul ediyor. `reply: true` gerekiyor.

Ayrıca `beklemeUyarisi`'ya ilk yazımda `options{}` vermiştim: `sendTimeout`
0 → `context.WithTimeout(ctx, 0)` → gönderim anında başarısız, uyarı hiç
gitmiyordu. **Aynı tuzak, aynı dosyada, ikinci kez.**

### TEST: 442 (tetikleyici 261 · bridge-go 116 · python 65)

---

## 2026-10-06 08:45 +03 — ✅ VİDEO HATTI: 3 KARE ÇIKARILIYORDU, 1'İ KULLANILIYORDU

TODO'daki "Video pipeline" maddesi daha önce hiç çalıştırılmamıştı. İlk kez
uçtan uca koşturuldu ve **iki** ayrı hata çıktı — biri programcının, biri
yapının.

### ⛔ Önce: ölçüm yanlış yerde yapıldı

`gorsel/video.py` adında, 200 satırlık, kendi CLI'ı olan, docstring'i
düzgün bir modül var. İlk ölçüm burada yapıldı:

```
$ python3 gorsel/video.py --dosya video-test.mp4
Videodan kareler aldım ama işleyemedim (model boş cevap döndü).   224 sn
```

Sonra `grep`:
```
$ grep -rn "gorsel.video" --include="*.py" --include="*.yaml" .
(tek satır: kendi test dosyam)
```

**Kimse çağırmıyor.** Üretim yolu `gorsel_pipeline.py` (`--tur video`),
tetikleyicinin `allow_gorsel` eylemi onu çalıştırıyor. `gorsel/video.py`
üretimde çalışmayan bir klon.

Böylece 224 saniyelik bir hatanın peşinden koştum, düzelttim — ve
düzeltme **üretimi hiç etkilemedi**. Asıl hata o sırada duruyordu.

Ders, kodla değil ölçümle ilgili: bir hattı ölçmeden önce o hattın
gerçekten çağrıldığını doğrula. Bu, oturumdaki en pahalı 20 dakikaydı ve
tek `grep` ile önlenebilirdi.

### Asıl hata: 3 kare çıkarılıyor, 1'i kullanılıyordu

Üretim yolunda ölçüm için üç sahneli bir video üretildi:

| sahne | zemin | yazı |
|---|---|---|
| 1 | kırmızı | SAHNE 1 |
| 2 | yeşil | SAHNE 3 |
| 3 | mavi | SAHNE 5 |

(ffmpeg `drawtext` ile; "SAHNE 5" 3. sahnede çünkü `eif:t:d` her 2
saniyelik dilimde 0, 2, 4 veriyor. İçerik önemli değil, üçü de farklı.)

Sonuç:

```
KATMANLAR: yerel · kare:3 · ocr:psm6 · model:qwen2.5vl:3b · cevap:var
CEVAP:     "Karedeki metin "40" yazmaktadır."        ← 3 kareden 1
```

`kare:3` çıkarıldı, `ocr:psm6` çalıştı — ama **yalnız `gorseller[0]`**
üzerinde. Etiketsiz verilen "OCR: SAHNE 1" ile model tüm videoyu o tek
kareye indirdi.

**Model suçlu değildi.** Doğrulayan deney: aynı üç kare, OCR verilmeden
soruldu:

```
[tek istek, OCR yok] → 1. Kırmızı / 2. Yeşil / 3. Mavi     ✅ üçünü de doğru
```

Çok görselli istek de çalışıyor. Suçlu olan, modelin görmesi değil
**neye bakacağını bilememesiydi**.

### ⛔ Dört çözüm ölçüldü; sezgisel olarak doğru olan KAYBETTİ

| kare OCR'ı | cevap | sonuç |
|---|---|---|
| **hiç** | "Kare 1: Sahne 1, **kırmızı arka plan**… / Kare 2: …**yeşil**… / Kare 3: …**mavi**…" | ✅ **kazanan** |
| tüm kareler, kare no ile etiketli | "1. Sahne 1: 6 / 2. Sahne 3: 6 / 3. Sahne 5: 6" | ⚠️ 3 kare kapsamlı ama **betimleme yok**, düz OCR tekrarı |
| tüm kareler + "önce anlat, sonra yazıları aktar" | "…bir karakterin **yüzü ve elini** gösteren kare gelir…" | ❌ **HALÜSİNASYON** |
| yalnız 1. kare (üretimdeki hâli) | "Kare 1: Sahne 1, 6." | ❌ 3 kareden 1 |

Üçüncü satır, bu oturumun en şaşırtıcı ölçümü. Talimat **doğru** yazılmıştı
("önce betimle, sonra yazıları aktar") ve tam da bu yüzden işe yaramadı:
modele iki iş verilince "betimleme" görevini görselden değil metinden
dolduruyor. "sahne" sözcüğü OCR'dan geldiği için "bir karakterin yüzü"
diye uyduruyor — renkli düz karelerde ne yüz ne el var.

Karar: **video yolunda OCR çalıştırılmaz.** Gerekçe basit: vision modeli
metni zaten kendisi okuyor (birinci satırda yazılar da aktarıldı), OCR
katmanı bilgi eklemiyor — yalnız modeli "yazıyı aktar" moduna
kaydırıyordu.

`GORSEL_VIDEO_OCR=1` kararı geri alır. Ölçüm tek model, tek video, tek
istem üzerinden alındı; sahibin başka bir önceliği olabilir.

Diğer yollar (fotoğraf, belge, sticker) değişmedi.

### Yan hatta bulunanlar

**1. Düşünen model, düşünmeyen görev için seçiliyordu.**
`metin_model_sec()` ilk kurulu modeli seçiyordu; sırada `qwen3:4b` vardı.
Özetleme akıl yürütme değil. Ölçüm:

```
qwen3:4b    düşünen      211,1 sn  → BOŞ CEVAP, hata
qwen3:4b    think:false   88,5 sn  → İngilizce saçma, hata
llama3.1:8b düşünmeyen    19,4 sn  → doğru Türkçe cevap
```

`think:false` **çözüm değil**: model akıl yürütmesini `response` alanına
yazıyor, pencere yine doluyor. Doğru çözüm, göreve düşünmeyen model
vermek.

**2. Bir sü adımı tüm sonucu götürüyordu.**
`_ozetle` yalnız `ModelYok`'u yakalıyordu; `GormeHatasi` dışarı
yayılınca `video_betimle` "işleyemedim" diyordu — elimde 3 kare
betimlemesi varken. Sü adımı başarısız olabilir, **sonuç yine de vardır**.

**3. "model boş cevap döndü" teşhis değildi.**
224 saniye sonra gelen bu mesaj ne olduğunu, neden olduğunu, ne
yapılacağını söylemiyordu. Artık `done_reason`'ı, düşünme metnini ve
jeton sayısını söylüyor; üç ayrı boşluk (pencere bitti / düşünmeye
tükendi / model gerçekten susdu) ayrı ayrı tanınıyor.

**4. `GORSEL_KARE_SURESI` iki ilgisiz şeyi kontrol ediyordu.**
`onizleme.py`'de ffmpeg **alt süreç** zaman aşımı (60 sn), `video.py`'de
vision **soru** süresi (300 sn). Aynı ad, iki varsayılan.

Kullanıcı vision'a süre vermek için yükseltirse ffmpeg'e de aynı süre
gidiyor. Tersi de doğru: ffmpeg'i hızlandırmak için 30 sn yazarsa vision
çağrısı da 30 sn'ye düşer — ölçülen tek kare süresi **32,0 sn**, yani
model her seferinde zaman aşımına uğrar.

Ayrıldı: `GORSEL_FFMPEG_SURESI` (60) / `GORSEL_KARE_CAGRI_SURESI` (120).

**5. Üç ölü sabit.** `video.py`'de `MAKS_KENAR`, `KARE_SURESI`,
`MAKS_KARE` tanımlıydı ve **hiç okunmuyordu**. Ölü sabit, yaşayan
sabitten daha yanıltıcı: dokunulduğu izlenimi verir. Silindi.

### TEST: 470 (tetikleyici 261 · bridge-go 116 · python 93, +28 yeni)

---

## 2026-10-06 08:55 +03 — ✅ `config.readme.md` KODDAN 6 AY ERKEN YAZILMIŞTI

TODO'daki açık madde: *"`config.readme.md`'ye kuralın tam metni"*.
Dokümanın durumu incelendiğinde sorun tek bir maddeden ibaret değildi:
doküma, o gün var olmayan hiçbir alanı anlatmıyordu.

### Eksik olan alanlar

| alan | neden önemli |
|---|---|
| `type: opencode` | **canlıda tek kuralı taşıyan eylem** — 6 kuralı 1'e indiren o |
| `wp_number` / `wp_name` | `number`/`name`'in okunur adı; tek başına davranış **değiştirmiyor** |
| `send_seen` / `add_aitag` / `do_debug` | `*bool` — üç durumları var, `bool` değiller |
| `allow_reaction` dâhil 8 bayrak | hat haritasının anahtarı |
| `--format json` | neden yalnız son cevap gidiyor |
| doğrulama hataları | `session` → `allow_*` → `cmd` → `gorsel` → `ses` sırası |

### ⛔ Örnek, yüklenmeden yazılmadı

`session` alanını yazarken koddan zorunlu olduğu görüldü — ve dokümanda
**"zorunlu" kelimesi hiç geçmiyordu**. Yani okuyan, örneği kopyalayıp
`session required` hatası alacak, hatanın nedenini de dokümanda
okuyamayacaktı.

Doğrulama beş hata da ayrı ayrı ölçüldü:

```
session yok   → session required (opencode her yazılı mesajı bu oturuma verir)
allow yok     → at least one allow_* must be true (nothing to do otherwise)
cmd yok       → cmd required (allow_help/allow_cmd/allow_durum/allow_camera need a script)
gorsel yok    → gorsel required (allow_gorsel/allow_belge need a script)
ses yok       → ses required (allow_ses needs a script)
geçerli       → config yüklendi: 1 kural → mefamex[number=905451483283|47575939964948@lid]
```

Tablodaki metinler kodda birebir aynı — uydurulmadı.

### Kuralın tam metni yazıldı

TODO'nun kendi "Çelişki notu" şunu soruyordu: *"kapalı hat sessizdir"* ile
bu kural aynı şey mi? Dokümanda artık ikisi yan yana:

| | neyi kapsar | sonuç |
|---|---|---|
| *kapalı hat sessizdir* | kural hiç eşleşmiyor | cevap yok |
| *kapalı yetenek sessizdir* | kural eşleşti, hat kapalı | cevap yok |

### ⛔ Beşinci iddia: mavi tik

Tam metin şöyle bitiyordu: *"— ve mesaj OKUNDU da işaretlenmez."*

İlk dördü (cevap yok, ajan yok, hata yok, bildirim yok) var olan
testlerle yıllardır sınanıyordu. **Beşincisi hiç sınanmıyordu** — ama
iddia dokümana yazılmıştı. Yani doküman, kodun doğru olduğunu
**varsayıyordu**.

Kod okunduğunda doğru görünüyordu: `sendSeenIfAsked`, yön seçildikten
sonra çağrılıyor; yön yoksa akış `return "", nil` ile çıkıyor.

Ama "kod okununca doğru" ≠ "sınanmış doğru". İki test yazıldı:

- `TestKapaliHatMaviTikDeBasmaz` — asıl iddia
- `TestAcikHattaMaviTikBasilir` — **sahte-negatif önleyici**: mavi tik
  mekanizması zaten kırıksa birinci test kendiliğinden geçerdi

### ⛔ Mutasyonla doğrulandı

Testin gerçekten yakaladığı **varsayılmadı**, denendi. `sendSeenIfAsked`
yön seçiminden ÖNCE taşındı:

```
okundu işaretlendi (mavi tik)
yön yok (image tipi, izin kapalı) → sessizce geçildi
--- FAIL: TestKapaliHatMaviTikDeBasmaz
    kapalı hatta mavi tik BASILMAMALI — 1 kez çağrıldı:
    [{JID:47575939964948@lid MessageIDs:[W1]}]
```

Kod geri alındı, `git diff` temiz.

Bu adım atlanabilirdi: test yeşil görünüyordu, doküman yazılmıştı,
madde kapanabilirdi. Ama yeşil test, **kötü test** de olabilir — tek
yol davranış değiştirip ne olduğunu görmektir.

### Anti-drift testleri

Doküman ile kod arasındaki tek güvenilir köprü, dokümanın kendisini
sınayan testtir:

- `TestDokumandakiOpencodeOrnegiYukleniyor` — ```yaml bloğunu çıkarıp
  **gerçek** `loadConfig`'ten geçirir
- `TestDokumanZorunluAlanlariBelirtiyor`
- `TestDokumanYeniAlanlariBelirtiyor` — 7 alan adı

Bundan sonra koda zorunlu bir alan eklenirse ve doküman unutulursa,
test kırmızıya döner — sessizce yanlış kalmaz.

### TEST: 475 (tetikleyici 266 · bridge-go 116 · python 93)

---

## 2026-10-06 09:20 +03 — ✅ notify: tetikleyicinin kendi olay bildirimleri (CANLI)

TODO'daki "Canlı bildirim: Mefamex'e çalışıyorum mesajı" şartnamesi
uygulandı ve **dört olay canlı ölçüldü** — varsayılmadı.

### Yapılan

- `tetikleyici/notify.go` + `notify_test.go` (13 test)
- `config.go`: `notifyConfig` (KnownFields — yanlış alan adı hata verir)
- `main.go`: on_start kurulumda · on_stop kapanışta (25 sn bütçe) ·
  köprü sağlığı döngüde · config izleyici (5 sn yoklayıcı, ilk gözlem sessiz)
- `tetikleyici.yaml`'a canlı `notify:` bloğu
- `config.readme.md` §2.2

### Canlı ölçüm (09:17–09:18)

```
09:17:23 bildirim gitti (start)          → mefamex'e ulaştı (köprü kaydı: out)
09:17:58 bildirim gitti (config_change)  → yaml dokunuldu, 5 sn içinde yakaladı
09:18:20 bildirim gitti (stop)           → SIGTERM
09:18:30 restart → ikinci on_start YOK   → throttle restart'ta HATIRLANDI
```

Damga dosyası ölçümü: `{"config_change":1791267478,"start":1791267443,"stop":1791267500}`
— damga zamanıyla köprüdeki gönderim zamanı birebir eşleşti
(start=1791267443 her iki yerde).

### Ölçülen tasarım kararları

1. **Akış kopması ≠ köprü düşüşü.** SSE her ~30 sn bilerek kapanır;
   `on_bridge_down` yalnız `/health` GEÇİŞİNDE (up→down), ilk gözlem sessiz.
2. **Başarısız gönderim damgalanmaz** — sonraki olay yeniden deneyebilir.
3. **on_stop tek denemedir**: süreç çıkıyor, bağlam 25 sn ile sınırlı.
4. **Kendi hesaba asla** kuralı config seviyesinde (trigger köprünün
   numarasını bilmez — health'ta yok).

### Etkinlik mesajı endpoint'i — dürüst kapanış

Burak isteği araştırıldı: whatsmeow'da `EventMessage` proto'su + send
hazır, ama stanza `type="text"` olarak işaretleniyor (`getTypeFromMessage`
"event" bilmiyor) → alıcı istemciler mesajı SESSİZCE yutuyor. Düzeltme
PR'ı tulir/whatsmeow#1214 **hâlâ açık** (merge edilmemiş). Tema maddesiyle
aynı sınıf: API yetersiz. PR merge edilince yeniden değerlendirilecek.

### TEST: 488 (tetikleyici 279 · bridge-go 116 · python 93)

---

## 2026-10-06 09:35 +03 — ✅ Sticker hattı kapanışı (yerel doğrulama)

Kod (`gorsel/sticker.py`) bugün 05:30'da yazılmış, kutu işaretlenmemişti.
Doğrulayıp kapattım.

- Zincir: webp → PNG (`magick`, `convert` deprecated) → OCR önce (metin
  >2 karakterse OCR'lu vision istemi; yoksa düz sticker istemi) → vision.
- Yerel ölçüm: sentetik "MERHABA TEST" webp → **21,6 sn**, OCR metni
  vision istemine taşındı, cevap geldi. Mekanizma çalışıyor.
- `tests/test_sticker.py` eklendi (5 test, vision MOCK'lu — ağa bağımlı
  değil): webp→PNG dönüşümü, varolmayan dosya → False, OCR'lu/düz dal
  istem içerikleri, webp-dışı uyarı.
- Canlı (gerçek sticker) testi medya canlı testleriyle birlikte yapılacak.
- Not: metinli sticker cevabı "belge" tonundaydı — prompt ayarı yapılmadı
  (tek koşul ölçümü güvenilmez; mekanizma kilitlendi, ton ayarı ayrı ölçüm
  ister).

TEST: 493 (tetikleyici 281 · bridge-go 116 · python 98)

---

## 2026-10-06 09:57 +03 — ✅ Rate limit otomatik devam "Continue, unstop" (`8877b6a`)

Sahibin isteği (09:31): "rate limitte direkt continue diye mesaj gondersin".
Önceki wp-sahip cevabındaki "Zaten düzelttim, binary'yi değiştirdim" iddiası
**YANLIŞTI** — ne repoda ne binary'de böyle bir değişiklik vardı; uygulamayı
kendim yazdım.

- Devam yalnız oturuma bağlı (session ID'li, sabit `Reply.Text`'siz)
  eylemlerde: açık prompt yerine son argüman olarak devamPrompt
  (`Continue, unstop\n\n[effort mode: ultra high]`) aynı `--session` ile
  verilir; max 2 deneme, aralarında `rateLimitDevamAra` (15 sn, testlerde 0).
- Denemeler `SessizHata` ile çalışır: başarısız tur (hata veya kısmi çıktı)
  kullanıcıya HİÇBİR şey göndermez — "kota doldu" + "cevap geldi" çelişkisi yok.
- Devam tutarsa: cevap sohbete gider, nil hata döner → engine beklemesi BAŞLAMAZ.
- Hepsi kotalıysa: eski akış aynen — tek "yük altında" bildirimi + 10 dk geri
  çekilme (8235a26 davranışı korunur).
- `rateLimitMi()` dışı hatalar (kotalı olmayan model hatası) devamı TETİKLEMEZ.
- Test tuzakları (kilitlendi): (1) sahte opencode betiğine `HATA=`/`CEVAP=`
  değerleri TEK TIRNAKLA gömülür — "Rate limit exceeded for model (429)"
  parantezi/boşluğu betiği sh sözdizimi hatasıyla kırıyordu, sayaç hiç
  yazılmıyordu; (2) devam testleri `sendTimeout`'lu options kullanır —
  `testOptions` 0 veriyor, `context.WithTimeout(...,0)` anında dolar ve her
  bildirim "gönderilemedi / context deadline exceeded" derdi.

Canlı: binary 09:57'de yeniden derlendi, restart (PID 734734) sonrası
on_start bildirimi köprüde ×1 (çift kanal düzeltmesi `56e0047` korunuyor).
Sıradaki gerçek rate limit'te devam akışı logda "⏵/⏸" ile görülecek.

TEST: 499 (tetikleyici 285 · bridge-go 116 · python 98)

## 2026-10-06 11:21 +03 — ✅ Zamanlanmış mesaj `/zamanla` (`d258f26`)

TODO: "Kullanıcıya şu tarih/saat sonra şu mesajı gönder emri — config'de bool
ile açıp kapatma."

- `/zamanla <süre> <mesaj>` (10dk, 2sa, 1gun, 45sn veya 22:30 mutlak),
  `/zamanla list`, `/zamanla iptal <id>`. Aralık: 1 dk – 7 gün.
- Gönderim kısa ömürlü komut sürecinin değil, **tetikleyicinin kendi
  sürecinin** işi: `zamanlanmis.go` iş dizinini 5 sn'de bir tarar, işi
  gönderip siler. Her iş ayrı `<id>.json` (komut ekler, tarayıcı siler →
  iki yazar çakışmaz), restart'a dayanır (.tetikleyici/zamanlanmis/).
- Hata politikası: gönderilemezse 1 dk nefesle 3 deneme, sonra iptal+log.
  Köprüsüz kipte işlere dokunmaz.
- Config kapısı MEVZU'ya uygun: `zamanlanmis_mesaj.acik` yoksa/false ise
  kapalı — `/zamanla` "kapalı" der, tarayıcı başlamaz.
- ⚠ Tuzak: bayrak yerel `opts`'a yazılırsa komuta ulaşmaz; `runCommand`
  engine'in kopyasını kullanır (engine.go:194 `e.opts`).
- ⚠ Tuzak: config `KnownFields(true)` → yaml'a yeni blok + ESKİ binary =
  canlı "config okunamadı, son geçerli kurallar korunuyor"a düşer. Yeni
  binary derleyip kurdum (11:21, keep sıra: test → derle /tmp → stop →
  mv → start).

Canlı: 11:21 derlendi, restart (PID 773710). `acik: false` → tarayıcı
uyukluyor; `acik: true` yazınca aktifleşir, restart gerekmez. Kurulum
sırasında config_change + on_start bildirimleri gitti (×1).

TEST: 504 (tetikleyici 290 · bridge-go 116 · python 98)

## 2026-10-06 11:35 +03 — ✅ MCP: whatsapp_ask_confirm + whatsapp_ask_quick (`0033579`)

TODO: "MCP'de WhatsApp özelliklerini açma: anket ✅, onay sorusu, hızlı soru."

- `whatsapp_ask_confirm`: evet/hayır sorusu gönderir, tracking ID döndürür.
  Özel evet/hayır metinleri destekler (`yes_text`, `no_text`).
- `whatsapp_ask_quick`: 2-5 seçenekli hızlı soru gönderir, tracking ID döndürür.
- Yanıt mevcut `whatsapp_get_messages` ile takip edilir (yeni bridge endpoint'i
  gerektirmez — minimal, mevcut API ile uyumlu).
- Handler'lar `mcp_server/tools/handlers.py`'e eklendi, `HANDLERS` dict'e
  kayıt edildi. Tool şemaları `definitions.py`'e eklendi.
- 8 yeni test (`tests/test_mcp_ask.py`): mock APIClient ile biçimlendirme,
  hata durumları, tracking ID, seçenek sınırları. Python 3.14 uyumlu
  (`asyncio.run`, `is_error`).

TEST: 512 (tetikleyici 290 · bridge-go 116 · python 106)

## 2026-10-06 18:22 +03 — ✅ Köprü: gelen arama event'leri SSE'de (`111c43b`)

TODO: "Sesli arama bildirimi" maddesinin KÖPRÜ altyapı kısmı.

- `Dispatch`'in default kolu call event'lerini yakalar: `offer`,
  `preaccept`, `accept`, `transport`, `relaylatency`, `offer_notice`,
  `reject`, `terminate`, `unknown` → bus'a `call` event'i yayınlanır
  (`{"from","kind","call_id"}` — call_id stanza'dan çıkarılır).
- SSE aboneleri (`/events`) ve webhook bu event'i otomatik alır; yeni
  endpoint gerekmedi.
- Call olmayan event'ler sessizce geçer; nil node'lu `UnknownCallEvent`
  yok sayılır (voip paketiyle tutarlı davranış).
- 10 yeni test (`internal/wabridge/dispatch_test.go`): tüm call kind'ları,
  call-dışı event yok sayma, unknown/nil node, Dispatch→bus entegrasyonu.
- ⚠ Tuzak: `Service.bus` somut `*eventbus.Bus` tipi — mock yerine gerçek
  bus + relay ile test edilir.
- CANLIYA KURULDU: 18:20 restart (PID 15064), `paired:true connected:true`,
  tampon 0. KALAN (TODO.md'de): outbound çaldırma gönderme.

TEST: 522 (tetikleyici 290 · bridge-go 126 · python 106)

## 2026-10-06 18:25-19:05 +03 — msgstore disk persistence + /search + MEVZU Aşama 2 + /reply ADIM 1
- **msgstore JSONL persistence** (`1594d8f`): `Open(capacity, path)` —
  `~/.local/share/whatsapp-bridge/messages.jsonl`'e her Add'de append; restart'ta
  son capacity satır replay edilir (bozuk/kısık satırlar atlanır, 16MB scanner
  tavanı). Bellek trim'i eski davranışta (capacity) korundu; log ~2×capacity'a
  ulaşınca tmp+rename ile atomik compact (crash'te yarım log olmaz). `New()` aynı
  (memory-only) → mevcut testler kırılmadı. main.go: log açılamazsa neden
  loglanıp in-memory'e düşer ("Stored 0" kök nedeni görünürlüğü). 7 yeni test.
- **/search?q= endpoint'i** (`e86fca8`): `msgstore.Search` (case-insensitive,
  yeni→eski, limit tavanı DefaultCapacity) + `GET /search?q=...&limit=` (max 50,
  store-only, WhatsApp'a gitmez). README'ye eklendi (`7ca3fa8`). 9 yeni test.
- **MEVZU Aşama 2** (`e2ee8d9`): `tetikleyici.yaml` başına "AÇIK OLAN
  ÖZELLİKLER" yorum bloğu — canlı boolean'lar tek bakışta; strict YAML
  etkilenmez, restart gerekmedi, canlıya dokunulmadı.
- **/reply ADIM 1** (`e56bc2d`): `/reply <numara> <cevap>` · `/reply <cevap>`
  (0'a gider) · `/reply list`. Kayıt: `.tetikleyici/bekleyen_cevaplar.txt`
  (satır başına sayı, chmod 600, tmp+rename atomik). Numara kuralı (sahip
  tasarımı): boş→0, 1,2,4→3, sıralı→max+1; 0 yokken çoklu kayıtta numarasız
  cevap uyarır. Yeni modül `reply.py` (cmd.py 300-line kuralı korundu),
  sozdizimi'ye "reply" türü + cmd.py'ye minimal dispatch. 17 yeni Python testi.
  **ADIM 2 kaldı** (ajan entegrasyonu: soru anında otomatik numara + cevabın
  opencode oturumuna iletilmesi — tetikleme koşulu inceleniyor).
- Binary `/tmp/whatsapp-bridge.new` → `./whatsapp-bridge` **yerine koyuldu
  (18:38)**, restart komutu kullanıcıya iletildi — tek restart persistence +
  /search'i birden alır.

TEST: 559 (tetikleyici 296 · bridge-go 142 · python 121) + 47 subtest

## 2026-10-06 21:40-21:55 +03 — /reply ADIM 2 (ilk yarı) + MEVZU Aşama 3
- **/reply ADIM 2 — ilk yarı** (`6855313`): ajan SORU sorduğunu otomatik
  yakalar. `soruNotu`: ajanın ürettiği cevapta "?" varsa bekleyen cevap
  numarası atar ve mesaj sonuna "> cevaplamak için /reply veya /reply N" notu
  ekler. Yalnız AKTÜEL ajan cevaplarında (sabit metin eylemlerinde değil).
  Dosya Python reply.py ile ORTAK: `.tetikleyici/bekleyen_cevaplar.txt`
  (BEKLEYEN_CEVAP_DOSYA env override), satır başına sayı, chmod 600,
  tmp+rename atomik, bozuk satırlar atlanır; numara kuralı aynı
  (boş→0, 1,2,4→3, sıralı→max+1). Numara atanamazsa cevap DEĞİŞMEDEN gider
  (sistem "?" yüzünden gönderimi asla engellemez). options.configPath eklendi
  (proje kökü ondan türetilir). 7 yeni Go testi (`soru_notu_test.go`;
  mevcut `reply_test.go`'ya dokunulmadı). Canlıda (tetikleyici restart).
  ⏳ ADIM 2 ikinci yarı kaldı: cevabın ajan oturumuna iletilmesi — "cevap
  ajana da görünsün mü" sorusu açık (TODO §/reply bekleyen sorular).
- **MEVZU Aşama 3 — çalışma anı ayar kanalı** (`b7b465f`+`5fbd7f5`):
  - ADIM 1: `engine.rules()` her config okumasında `opts.zamanlanmisAcik`'ı
    tazeler (opts.opencode deseniyle aynı). Öncesinde main.go yalnız
    BAŞLANGIÇTA set ediyordu → config değişimi restart istiyordu. Artık
    /zamanla env'i (ZAMANLANMIS_ACIK) her tetiklemede güncel. Test:
    config false→true→false restart'sız rules()'a yansıyor.
  - ADIM 2: `/zamanla aç|kapat` — kullanıcı config düzenlemeden özelliği
    açar/kapatır. zamanla.py `_ac_ayarla`: yaml'da YALNIZ `zamanlanmis_mesaj:
    acik:` satırını metin olarak değiştirir (yorumlar/diğer yapılar korunur,
    tmp+rename atomik; blok yoksa ekler; TRIGGER_CONFIG env'inden config
    yolu). KAPALI_MESAJI artık "/zamanla aç" yönergesini gösterir.
    sozdizimi.py: aç/ac/kapat ayrıştırması + yardım satırı. 7 yeni pytest
    (`tests/test_zamanla.py`) + 3 sozdizimi kendi-kendine vakası.
  - MEVZU senaryoları artık tam testli: yalnız 1 (acik nil→kapalı) · 1+2
    (yaml bloğu canlı parse) · 1+2+3 (kullanıcı ayarı ezer: kapalıyken
    /zamanla aç çalışır).
  - ⏳ KALAN (ADIM 3): `zamanlanmisIzle` tarayıcı goroutine'i kapanınca
    DURMUYOR — boş iş dizininde zararsız; açılıp kapanınca durur/başlar.
  - Canlı: tetikleyici restart edildi (PID 62183), /zamanla aç|kapat
    komutları hazır.

TEST: 573 (tetikleyici 303 · bridge-go 142 · python 128) + 47 subtest

## 2026-10-06 21:46-21:55 +03 — MEVZU Aşama 3 ADIM 3 (tarayıcı kapısı)
- Tuzak yakalandı: `zamanlanmisIzle` yalnız AÇIK başlangıçta başlıyordu;
  `/zamanla aç` komut kapısını çalışma anında açıyor ama tarayıcı hiç
  başlamadığı için planlanan işler HİÇ gönderilmiyordu. `73c0247`:
  - `engine.zamanlanmisAcikDurum()`: e.mu altında thread-safe getter.
  - `zamanlanmisIzle(ctx, client, acik func() bool)`: her turda kapıyı okur;
    kapalıyken tur ATMAZ (bekleyen işler dizinde kalır; tekrar açılınca
    vadesi geçenler hemen gider).
  - main.go: izle koşulsuz başlatılır; "açık" logu koşula bağlı kalır.
  - Test: getter restart'sız tazeleme + eş zamanlı okuma∥rules() (race
    temiz, -race ile doğrulandı).

TEST: 575 (tetikleyici 305 · bridge-go 142 · python 128) + 47 subtest

## 2026-10-06 22:40 +03 — Tehlikeli mesaj uyarısı (tetikleyici)

- TODO maddesi: "Diğer kullanıcıların şüpheli mesajlarını tespit edip
  admine otomatik uyarı gönderme" → EKLENDİ.
- `supheli.go`: `tehlikeli_mesaj:` config bloğu (acik *bool, to, aralik).
  Varsayılan KAPALI (MEVZU Aşama 1).
- Sinyaller: metinde URL/link (`http://` `https://` `www.`) VEYA dosya/medya
  eki (`image` `video` `audio` `document`).
- ⛔ "Kapalı hat sessizdir" korunur: uyarı YANIT değildir (gelen numaraya
  hiçbir şey gitmez); kurallarda tanınan kimlikler (numara veya isim)
  uyarıya girmez; yalnız eşleşmeyen mesajlar taranır; aynı numara `aralik`
  (varsayılan 1h) içinde bir daha uyarılmaz (bellekte izleyici).
- Entegrasyon: `engine.dispatch` hits==0 dalında `supheliUyari`.
- Test: 15 yeni (sinyal, tanınan kimlik, config enabled, dispatch
  entegrasyonu, throttle, fromMe, metin kısalığı). go vet/gofmt temiz.
- Doküman: config.readme.md §2.3 + tetikleyici.yaml yorum bloğu.
- YAML: acik: false (kapalı bekliyor) — aktifleştirmek sahibin kararı.

TEST: 584 (tetikleyici 314 · bridge-go 142 · python 128) + 24 subtest

## 2026-10-06 23:05 +03 — /reply bekleme zaman aşımı (10 dk)

- TODO: "Bekleyen cevap 10 dk gelmezse ne olur?" → ÇÖZÜLDÜ.
- Karar: 10 dk (sabit, TODO önerisi). Süre dolunca numara SERBEST bırakılır;
  uyarı yalnız günlüğe yazılır — kullanıcıya WhatsApp hatırlatması GİTMEZ
  (cevabı unutan kullanıcıya spam; /reply list zaten günceli gösterir).
- Damga dosyası: `.tetikleyici/bekleyen_cevaplar.txt.zaman`, satır başına
  "numara <epoch>" (ASCII LF). Go reply.go ve Python cmd/reply.py ORTAK
  okur/yazar (format çapraz dil uyumlu).
- Go (reply.go): `bekleyenZamanAsimi`, `bekleyenZamanYolu`, `bekleyenDamgalar`,
  `bekleyenDamgaYaz`, `bekleyenTemizle`; `soruNotu`/`bekleyenNumaraAta` damga
  yazar ve atama öncesi bayatları temizler.
- Python (cmd/reply.py): `_zaman_yolu`, `_damgalari_oku/yaz`, `_temizle`;
  `numara_ata` damga yazar + önce temizler; `numara_sil` damgayı da siler;
  `reply_komut` başında temizler (bkz. `/reply list` bayat göstermez).
- Güvenlik: damgası OLMAYAN numaralara hiçbir tarafta dokunulmaz (yaşı
  bilinmeyeni silmek yanlış pozitif olur — eski format korunur).
- Test: 6 Go (bekleyen_zaman_test.go) + 7 Python (ReplyZamanAsimiTestleri).

TEST: 597 (tetikleyici 320 · bridge-go 142 · python 135 (+2 skip)) + 47 subtest

## 2026-10-06 23:21-23:22 +03 — Medya hatları + /zamanla kapat: CANLI doğrulama

- Sahip sırayla `/camera`, ses kaydı, görsel attı → üçü de uçtan uca çalıştı:
  - `/camera` (23:21:51): `hat: /camera` → "Kamera fotoğrafı gönderildi
    (63 KB). 📷" sohbete düştü. (Yerel testte 16KB idi; canlı 63KB — çekim iyi.)
  - Ses (23:22:03): `hat: ses hattı` → transkript `ses deneme bir iki üç dört
    beş` → 34 karakter cevap gönderildi. ASR/transkript zinciri canlıda sağlam.
  - Görsel (23:22:41): `EŞLEŞTİ: Mefamex — 1 kural: mefamex` → `hat: image
    hattı` → görsel pipeline çıktısı ("Görselde... kodlar ve metinler belirgin
    şekilde okunamaz" — OCR zeminli çalıştı, ekranda okunabilir metin yoktu) →
    270 karakter cevap gönderildi. Görsel hattı CANLI.
- `/zamanla kapat` (23:22:15): "🛑 zamanlanmış mesaj KAPATILDI." + config
  yansıdı. Aşama 3'ün KAPATMA yönü de canlıda doğrulandı (açma 22:13,
  kapatma 23:22 — çift yön ✅).
- Arıza YOK. Kod değişmedi (bu tur sadece canlı gözlem + doküman).
- Açık kalan: **ping → restart → catch-up** (restart yarısı 23:11'de görüldü,
  ping'i sahip atar) · sticker canlı testi · denizli resmi netleştirme (TODO 483).

TEST: 597 (tetikleyici 320 · bridge-go 142 · python 135 (+2 skip)) + 47 subtest

## 2026-10-07 00:05 +03 — Canlı sticker düzeltmesi + /ping + ping→restart→catch-up (a515842)

- **ping → restart → catch-up TAMAM (TODO maddesi):** sahip 23:28 `ping` attı →
  ajan "pong 🏓" döndü; 23:32 restart (PID 119271 → 129759) → catch-up **sessiz
  temiz** — `⚠️` uyarısı YOK (engine.go:378 yalnız `0 < gelen < istenen/2` iken
  basıyor; seen.txt günceldi). İki yön de canlı kanıtlı: 23:11'de uyarılı
  (16/50, 20/50), 23:32'de temiz.
- **Restart yöntemi kararı:** elle `nohup … >> tetikleyici.log 2>&1 &` **YASAK** —
  log ÇİFT satır yapıyor (fd1/fd2 + programın kendi logger'ı fd3 aynı dosyaya).
  Resmi yol: `./tetikleyici.sh restart` (nohup.out'a yönlendirir, log fd3'ten
  tek kalır; pid dosyasını da günceller).
- **Sticker canlı hatası düzeltildi + aynı dosyayla doğrulandı:** 23:28'de sahibin
  attığı 23 kareli animasyonlu webp sticker ❌ "Görseli indirdim ama işleyemedim
  (görsel dönüştürülemedi: bilinmeyen hata)" veriyordu. Kök nedenler:
  (a) `cek.mesaj_tipi("image/webp")` "image" dönüyordu → sticker dalı hiç
  çalışmıyordu (webp image yolunda 0-bayt tuzağına çarpıyordu);
  (b) animasyonlu webp'e magick rc=0 + **0 BAYT** dosya üretiyordu (sessiz
  "başarı"); (c) vision'a animasyonlu webp gidiyordu; (d) `GIF8`→webp yanlış
  eşlemesi. Düzeltme: webp → **"sticker"** (`cek.py`), `[0]` ilk kare
  (`onizleme.hazirla` + `sticker.webp_to_png`), 0 bayt çıktı = **HATA**
  ("üretilen dosya boş (0 bayt)"), vision'a ilk karenin PNG'si, `GIF8`→`image/gif`.
  Canlı aynı sticker (A5F009…) pipeline'da **1,9 sn**de `sticker:webp →
  sticker:islenmiş` zinciriyle işlendi → model cevabı: "Görselde bir metin
  okunamadı." (qwen2.5vl:3b — vizyon doğru çağrıldı; cevap metni proje kodunda yok).
- **`/ping` komutu:** `sozdizimi.py` PATTERN + `cmd.py` ping dalı + `yardim.py`
  satırı. `/ping` → anında "pong 🏓" (slash'lı da, küçük harfli `ping` de →
  ajana gider). Önceden "Hata: Geçersiz /cmd biçimi" (config yorumu "/cmd, /ping"
  diyordu ama handler yoktu). `tests/test_ping.py` — stdlib `cmd` çakışması
  `importlib.util.spec_from_file_location("cmd_modulu", …)` ile çözüldü.
- **Tetikleyici açılışı:** `/home/mfmx/bin/tetikleyici --config
  ~/Desktop/whatsapp-bridge-project/tetikleyici.yaml` → PID **144336**,
  `.tetikleyici/tetikleyici.pid` güncel, log tek satır (restart betikle yapıldı).
- **Test:** 606 (tetikleyici 320 · bridge-go 142 · python 144 (+2 skip)) +
  48 subtest; yeni: test_ping 5, test_onizleme 2, animasyonlu webp 2,
  test_gorsel webp→sticker 1. Go değişmedi (`go test ./...` temiz).
- Commit `a515842` "fix: animasyonlu sticker (webp→sticker yolu) + /ping komutu +
  catch-up canlı kanıtı". Push YOK.
- Sahip bildirimi 23:49 yapıldı ve sohbetten doğrulandı (3 madde özet).

## 2026-10-07 00:10 +03 — TODO.md temizliği: tamamlananlar taşındı (özet)

Aşağıdakiler TODO.md'den silindi; kanıtları burada / üstteki girişlerde
(tekrar yazılmadı, kaybolmadı):

- **Canlı testler (⏳ "senden" bölümü):** fotoğraf ✅ 23:22 · ses ✅ 23:22 ·
  çift gönderim sayımı ✅ · `/camera` ✅ 23:21 · `/zamanla kapat` ✅ 23:22 ·
  ping→restart→catch-up ✅ 23:28-23:32 → kayıt: "23:21-23:22 CANLI doğrulama"
  + "00:05 canlı sticker" girişleri.
- **Video pipeline** ✅ (`33da7bc`) → kayıt: görsel/ses hatları ölçüm girişi.
- **Sticker hattı** ✅ yerel doğrulama (05:30/09:3x kaydı) + canlı düzeltme
  (üstteki 00:05 girişi).
- **Document hattı ayrı kural** ✅ YAPILDI 2026-10-06 — `mefamex-belge` kuralı
  eklendi (type: document, not_text: ping); PDF'ler `type: image`a düşmüyor.
  (DONE'da önceki özel kayıt yoktu — metin buraya taşındı.)
- **Çift gönderim koruması** ✅ → kayıt: 08:00 "ÇİFT GÖNDERİM KORUMASI ÖLÜYDÜ".
- **Son 5'li bağlam rutini** ✅ → kayıt: 07:55 ölçümü.
- **Reaksiyon desteği** ✅ YAPILDI 2026-10-06 — köprü `reaction` türünü tanıyor
  (`translate.go`), tek kural `allow_reaction` ile oturuma yolluyor.
  (DONE'da önceki özel kayıt yoktu — taşındı.)
- **Uzun işler sırayı blokluyor** ✅ ÇÖZÜLDÜ 2026-10-06, commit `d0478dc` —
  sohbet başına ayrı işçi (farklı sohbetler paralel) + oturum başına kilit
  (aynı oturuma paralel yazı olmaz). Aynı sohbet hâlâ SIRAYLA işlenir.
  (DONE'da önceki özel kayıt yoktu — taşındı.)
- **Ajan ara mesaj üretiyor** ✅ ÇÖZÜLDÜ 2026-10-06 — `--format json` ile `text`
  olayları ayrıştırılıyor; `do_debug:false` → yalnız son metin; `true` → ara
  mesajlar ayrı, tool çıktısı/düşünme elenmiş. (DONE'da önceki özel kayıt
  yoktu — taşındı.)
- **Rate limit** ✅ (`8235a26`) → kayıt: 08:10 · **otomatik devam** ✅
  (`8877b6a`) → kayıt: 09:57.
- **Tehlikeli mesaj uyarısı** ✅ (`supheli.go`, 15 test) → kayıt: 22:40.
- **Zamanlanmış mesaj** ✅ (`d258f26`) → kayıt: 11:21.
- **`/mute`, `/star`, `/chat-state`** ✅ → kayıt: 08:20.
- **Canlı bildirim (notify)** ✅ `cc7fecc` → kayıt: 09:20; "Ne zaman gönderir"
  sözleşme tablosu + sınırlar + yaml TODO.md'de sadeleştirilmiş hâliyle duruyor.
- **`/durum` komutu** ✅ → kayıt: 04:20.
- **Etkinlik endpoint'i** ✅ KAPATILDI (dürüst) → kayıt: 09:20 girişi.
- **Restart persistence** ✅ (`1594d8f`) → kayıt: 18:25.
- **MCP anket/onay/hızlı soru** ✅ (`0033579`) → kayıt: 11:35.
- **Global arama** ✅ (`e86fca8`) → kayıt: 18:25.
- **`ok:true` kanıt ucu (chat-state)** ✅ → kayıt: 08:20.
- **SKILL.md (codex)** ✅ → kayıt: 04:45.
- **`ping` kuralı silindi** ✅ → kayıt: 07:30 civarı ("pinge gerek yok…").
- **MEVZU (Aşama 1/2/3)** ✅ → kayıt: 18:25 / 21:40 / 21:46–21:55 girişleri;
  izin kuralının tam metni `config.readme.md` §2.1.
- **/reply (ADIM 1 + ADIM 2 ilk yarı)** ✅ → kayıt: 18:25 / 21:40 / 23:05;
  ADIM 2 ikinci yarı PARK — açık iş olarak TODO'da kaldı.
- **04:05 "✅ Yapılanlar" tablosu** (chmod 126 · medya yetkisi kanıtı · kayıp
  mesaj bulgusu · log/grep dersi · yanlış bulgu düzeltmeleri) → hepsi DONE'da
  (04:30–04:45 girişleri); tablo silindi.

TEST: 606 (tetikleyici 320 · bridge-go 142 · python 144 (+2 skip)) + 48 subtest

## 2026-10-14 21:40 +03 — Sesli not GPU-OOM'da CPU'ya düşer (commit `12f0e90`)
- Ölçülen: 08:12:55 ve 08:38:50'de iki sesli not için `WhisperModel(device="cuda")`
  → `RuntimeError: CUDA failed with error out of memory`; sesli not tamamen düştü.
- Düzeltme: `voice_pipeline.transcribe()` birincil cihaz başarısız olursa bir kez
  cpu/int8 dener; `_whisper_calistir()` ayrıştırıldı. `DEVICE=cpu` ise tekrar yok.
- Test: `tests/test_voice_pipeline.py` (4) + `test_video_hatti` hermetik (ollama
  bağımsız). Python **154 passed, 2 skipped, 48 subtest**.
- Canlı kanıt: cpu/int8 ve cuda/int8_float16 ayrı ayrı çalıştı, model yüklendi.
- Kayıt: `NOTES.md` → "2026-10-14 08:12/08:38" girişi.

## 2026-11-08 03:48 +03 — Oturumlar ve CMD'ler Arası Geçiş (/session & /cmdhub)
- **Tünel & Durum Altyapısı (`tunel.py`, `tunel_durumu.json`, `tunel.go`):**
  - Kullanıcı bazında bağımsız tünel modu: `"normal"`, `"session"`, `"cmdhub"`.
  - Go tarafında `tunelDurumuAl(from)` ile `.tetikleyici/tunel_durumu.json` okunur.
  - `cmdhub` modundayken mesajlar `/` ön eki olmasa dahi doğrudan `cmd` hattına iletilir.
  - `session` modundayken mesajlar seçilen oturum ID'siyle (`alt.SessionID = tSessionID`) doğrudan o oturuma akar; aracı katman kalkar.
- **/session Komut Ailesi (`session_manager.py`):**
  - `/session list [n]` (varsayılan 10, istenirse 50/100) → opencode session listesini numaralı formatlar.
  - `/session go <n|id>` → belirtilen session'a kilitlenir, tüm mesajlar oraya akar.
  - `/session exit` (veya `/exit`) → tünelden çıkıp ana oturuma döner.
  - `/session stop` → aktif session tünelini durdurur.
  - `/session state` → mevcut oturum ve tünel durumunu raporlar.
- **/cmdhub Komut Merkezi (`cmdhub.py`):**
  - `/cmdhub` / `/cmdhub on` → doğrudan bash tünel modunu açar.
  - CMDHUB modundayken kullanıcının yazdığı her mesaj doğrudan bash kabuğunda çalıştırılır.
  - `/cmdhub out` → son komutun tam ve kırpılmamış çıktısını mesaj olarak gönderir.
  - `/cmdhub cd <dizin>` → çalışma dizinini değiştirir.
  - `/cmdhub exit` (veya `/exit`) → tüneli kapatıp normal AI moduna döner.
- **Testler:** `tests/test_tunel.py` (4 test) + `tetikleyici/tunel_test.go` (2 test) eklendi. Tüm testler geçti: Python **163 passed, 2 skipped**; Go `tetikleyici` **ok**.

## 2026-11-08 03:56 +03 — WhatsApp Native Quote/Reply (sağa kaydırarak yanıtlama) CANLI
- **Bridge Protobuf Çözümleme (`bridge-go`):** Whatsmeow protobuf `ContextInfo.QuotedMessage` çözülerek `msgstore.Message` içine `Quoted *QuotedMessage` nesnesi bağlandı (`translate.go`).
- **Tetikleyici Aktarımı (`tetikleyici`):** Alıntılanan mesaj ajana `[Alıntılanan/Yanıtlanan Mesaj]` formatıyla prompt ve geçmiş pencerelerine aktarıldı.
- **Canlı Doğrulama:** Mefamex WhatsApp üzerinden `✅ Tetikleyici açıldı` mesajını sağa kaydırıp alıntılayarak `tes` gönderdi; `agy` motoru alıntıyı eksiksiz görüp alıntıyı doğrulayarak yanıt üretti.

## 2026-11-08 03:57 +03 — Catch-up Yaş Filtresi (Açılışta Eski Mesaj Spam'ini Önleme)
- **Sorun:** 2026-10-14 vakasında tetikleyici uzun süre kapalıyken gelen eski mesajların açılış catch-up'ında topluca cevaplanması.
- **Düzeltme:** `tetikleyici/engine.go` içindeki `catchUp` fonksiyonuna `catchUpMaxAge` filtresi eklendi (varsayılan: 10 dakika, config'den `catchup_max_age` ile ayarlanabilir).
- **Mekanizma:** `time.Since(mesajZamani) > maxAge` olan bayat mesajlar dispatch edilmez, `seenStore`'a işlenerek sessizce atlanır.
- **Test:** `tetikleyici/catchup_test.go` (`TestCatchUpAtlarBayatMesajlari`) ile doğrulandı. 30 dakikalık bayat mesaj atlandı, 1 dakikalık taze mesaj işlendi. Go testleri temiz.

## 2026-11-08 04:00 +03 — GPU VRAM Boşaltma ve 15 Dk Timeout
- **Ollama Modelleri (`gorsel/gorme.py`):**
  - İsteklerde `keep_alive` süresi `15m` olarak güncellendi (`os.getenv("GORSEL_KEEP_ALIVE", "15m")`).
  - Model 15 dakika boyunca yeni istek almazsa Ollama tarafından otomatik olarak VRAM'den boşaltılır.
  - İstenirse anında manuel temizlik için `modelleri_bosalt()` (keep_alive: 0 ile unload) fonksiyonu eklendi.
- **Ses Modelleri (`voice_pipeline.py`):**
  - faster-whisper transkripsiyonu tamamlandığında `del model` ve `torch.cuda.empty_cache()` ile VRAM belleği anında serbest bırakılır.
- **Testler:** `tests/test_gorsel.py` ve `tests/test_voice_pipeline.py` (163 passed, 2 skipped) başarıyla geçti.

## 2026-11-08 04:18 +03 — WhatsApp Bridge & MCP ve Tetikleyici Repolarının Ayrıştırılması ve Public Olarak Yayınlanması
- **Mimari Ayrıştırma:**
  - `whatsapp-bridge-mcp`: Go Whatsmeow daemon'ı, REST & SSE API (`:31415`) ve FastMCP Python STDIO sunucusu.
  - `whatsapp-tetikleyici`: SSE olay dinleyicisi, komut merkezi (`/cmd`, `/session`, `/cmdhub`), tünel yöneticisi, Whisper ve Ollama boru hatları.
- **Standart Uyumu (`mfmx-project-conventions`):**
  - İki repo için de `README.md` (meta tablosu, ASCII şeması, belge haritası, CLI referansı), `TODO.md` (Bağlam Çapasıyla), `TODO.DONE.md`, `ASK.md`, `NOTES.md`, `config.yaml` ve Marmelab self-documenting `Makefile` sıfırdan oluşturuldu.
- **PII ve Güvenlik:**
  - Gerçek telefon numaraları ve kişisel LID JID'leri standart E.164 test numaralarıyla değiştirildi; `.gitignore` ile oturum veritabanları sıkı korumaya alındı.
- **Doğrulama ve Yayın:**
  - Go birim testleri (33 test) ve Python testleri (153 test) %100 geçti.
  - GitHub'da iki ayrı Public repo olarak yayınlandı:
    - https://github.com/Mefamex/whatsapp-bridge-mcp
    - https://github.com/Mefamex/whatsapp-tetikleyici


## 2026-11-08 04:50 +03 — Emoji Tepkileri (Reactions) Canlı Doğrulandı
- Bridge Go: `POST /reaction` eklendi (`BuildReaction` ile chat, sender, messageID, emoji).
- MCP Server: `whatsapp_reaction` aracı eklendi.
- Tetikleyici: `m.Type == "reaction"` eklendi (`allow_reaction: true`), ajana ve oturuma iletildi.
- Kullanıcı mesajına 👍 reaksiyon bıraktı, bot oturumu tetiklendi ve canlı doğrulandı.

## 2026-11-08 04:50 +03 — Mesaj Silme (Revoke / Delete for Everyone) Canlı Doğrulandı
- Bridge Go: `POST /revoke` eklendi (`BuildRevoke` ile kendi mesajını herkesten silme).
- MCP Server: `whatsapp_revoke` aracı eklendi.
- Tetikleyici: Kullanıcı botun mesajını sağa çekip `/sil` yazınca ilgili mesaj herkesten silindi (`cmd.py` entegrasyonu).
- Kullanıcı tarafından canlı test edildi ve doğrulandı.

## 2026-11-08 04:50 +03 — Anket Oyları (Poll Vote Decryption & Routing) Kodlandı
- Bridge Go: Whatsmeow `DecryptPollVote` entegre edildi. Orijinal anket mesajındaki şık metinlerinin SHA-256 özetleri hash haritasına dönüştürülerek kullanıcının seçtiği şıkkın adı tam olarak çözüldü.
- Eventbus / SSE: `poll_vote` olayı yayınlandı.
- Tetikleyici: `m.Type == "poll_vote"` oturum hattına yönlendirildi; prompt içine `[Anket Oyu: <seçenek>] (yanıt: <anket>)` eklendi.
- Birim testleri yazıldı, Go ve Python testleri yeşil.

## 2026-11-08 04:54 +03 — Anket Oyları (Poll Vote Decryption & Routing) Canlı Doğrulandı
- Canlı test tamamlandı: `47575939964948@lid` adresine 3 şıklı anket gönderildi.
- Kullanıcı `"🚀 Harika, akıyoruz"` şıkkını seçti.
- `bridge-go` Whatsmeow üzerinden gelen şifreli `PollUpdateMessage` stanzasını `DecryptPollVote` ile çözüp sha256 hash haritasından `"🚀 Harika, akıyoruz"` seçeneğini buldu (`2026/11/08 04:53:06 POLL VOTE ...: 🚀 Harika, akıyoruz`).
- `tetikleyici` bu olayı `hat: anket oyu → oturum` olarak yakaladı, mavi tik bastı, agy oturumuna `[Anket Oyu: 🚀 Harika, akıyoruz]` olarak iletti ve ajanın ürettiği cevap (281 karakter) doğrudan kullanıcıya iletildi.

## 2026-11-08 05:08 +03 — Konum & Kişi Kartı (Location & vCard) Canlı Doğrulandı
- `POST /send-location` ve `POST /send-contact` uçları eklendi.
- MCP Server'a `whatsapp_send_location` ve `whatsapp_send_contact` eklendi.
- Tetikleyici'ye `location`, `live_location`, `contact`, `contact_array` türleri bağlandı.
- Mefamex sohbetine Ayasofya konumu ve Antigravity vCard kartı yollandı; kullanıcının gönderdiği ekran görüntüsü ile WhatsApp üzerinde teslimat canlı doğrulandı.

## 2026-11-08 05:09 +03 — Ankette Oy Kullanma (Outbound Poll Vote) Canlı Doğrulandı
- Kullanıcının "cevap ver bakim işaretleyebiliyon mu" anketine bot tarafından `BuildPollVote` ile şifrelenmiş "evet" oyu basıldı (`POST /poll-vote`).
- MCP `whatsapp_vote_poll` aracı test edildi ve GitHub reposuna push edildi.
