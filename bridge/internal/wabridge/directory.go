// whatsapp-bridge-project/bridge-go/internal/wabridge/directory.go

package wabridge

import (
	"context"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"

	"whatsapp-bridge/internal/appcfg"
)

// sohbetJID, kullanıcıdan gelen sohbet kimliğini çözer ve bozuk olanı
// REDDEDER.
//
// NEDEN AYRI FONKSİYON (2026-10-06 ölçümü)
// types.ParseJID ÇOK YUMUŞAKTIR: yalnız kullanıcı kısmında nokta veya iki
// nokta üst üste gelirse hata verir. Ölçülen davranışlar:
//
//	"gecersiz jid @@@"  → hata YOK   (sunucu kısmı boş çıkıyor)
//	"a@b@c"             → hata YOK   (a@b'ye sessizce KIRPILIYOR)
//	"555"               → hata YOK
//
// Bu köprü için tehlikeli: /mute bozuk kimlikle çağrılırsa WhatsApp'a
// SESSİZCE boş bir sohbet yaması gider, hata da dönmez. Kullanıcı telefonunda
// "susturulmadı" görür, sebebini bulamaz. Bu yüzden sunucu kısmı ve
// kullanıcı kısmı AYRI ayrı doğrulanır.
func sohbetJID(jid string) (types.JID, error) {
	normalize := appcfg.NormalizeJID(strings.TrimSpace(jid))
	if strings.TrimSpace(normalize) == "" {
		return types.JID{}, errJIDRequired
	}

	// @ işareti birden fazlaysa kırpma sessizce yanlış hedefe yol açar.
	if strings.Count(normalize, "@") != 1 {
		return types.JID{}, errJIDMalformed
	}

	parcalar := strings.SplitN(normalize, "@", 2)
	kullanici, sunucu := parcalar[0], parcalar[1]
	if kullanici == "" {
		return types.JID{}, errJIDMalformed
	}
	if sunucu == "" {
		return types.JID{}, errJIDMalformed
	}

	// Sunucu, bildiğimiz gerçek sunuculardan biri olmalı. "b@c" gibi uydurma
	// sunucular WhatsApp'a gönderilirse sessizce kaybolur.
	//
	// LID sunucusu ayrıca LİSTEDİR: bu projede sohbetler büyük ölçüde
	// @lid ile gelir ve LID'i kabul etmezsek asıl kullanıcıyı reddetmiş
	// oluruz.
	switch sunucu {
	case types.DefaultUserServer, types.LegacyUserServer,
		types.GroupServer, types.BroadcastServer, types.NewsletterServer,
		types.HiddenUserServer, types.HostedLIDServer, types.MessengerServer:
	default:
		return types.JID{}, errJIDMalformed
	}

	return types.ParseJID(normalize)
}

// ChatState is WhatsApp's own view of a chat's settings, as this device knows it.
//
// NEDEN VAR: /mute ve /star "ok:true" döndürüyor, ama bu projede "ok:true
// sessiz başarısızlık olabilir" diye bir kural var (bkz. NOTES.md). Yama
// gönderildi diye WhatsApp'ın KABUL ETTİĞI anlamına gelmiyor. Bu tip, yazma
// işleminden sonra OKUNABİLİR tek kanıt: ayar gerçekten değişti mi?
type ChatState struct {
	// JID, durumun okunduğu KANONİK sohbet kimliği.
	//
	// Neden ayrı alan: istek "905550000001" (serversız) ile gelirse
	// normalize edip "905550000001@s.whatsapp.net" üzerinde çalışılır.
	// Yanıt ham girdiyi yankılarsa kullanıcı "hangi sohbet?" diye
	// durur — halbuki kanıt başka bir sohbeti gösteriyor. Bir kanıt
	// endpoint'i kimliğini kendisi doğrulayamıyorsa işe yaramaz.
	JID string
	// MutedUntil, susturmanın bitiş zamanı. Sıfır değilse susturulmuş demektir.
	MutedUntil time.Time
	// Pinned, sohbetin sabitlenmiş olup olmadığı.
	Pinned bool
	// Archived, sohbetin arşivlenmiş olup olmadığı.
	Archived bool
}

// Muted reports whether the chat is currently silenced.
func (c ChatState) Muted() bool { return !c.MutedUntil.IsZero() }

// ChatSettings reads a chat's settings back from the local store.
//
// DİKKAT: bu, sunucudan taze soru DEĞİL, yerel önbelleğin durumudur. Yama
// gönderildikten sonra önbellek güncellenir (whatsmeow yeniden eşitleme
// tetikler), ama anlık değildir. Bu yüzden "doğrulandı" demek için hemen
// okumak yerine biraz beklenmelidir.
func (s *Service) ChatSettings(ctx context.Context, jid string) (ChatState, error) {
	target, err := sohbetJID(jid)
	if err != nil {
		return ChatState{}, err
	}

	bulunan, err := s.client.Store.ChatSettings.GetChatSettings(ctx, target)
	if err != nil {
		return ChatState{}, err
	}

	return ChatState{
		JID:        target.String(),
		MutedUntil: bulunan.MutedUntil,
		Pinned:     bulunan.Pinned,
		Archived:   bulunan.Archived,
	}, nil
}

// Contact is the JSON shape of a WhatsApp contact.
type Contact struct {
	JID          string `json:"jid"`
	Name         string `json:"name"`
	PushName     string `json:"push_name"`
	FirstName    string `json:"first_name"`
	BusinessName string `json:"business_name"`
	Found        bool   `json:"found"`
}

// Contacts returns every contact stored in the session database.
func (s *Service) Contacts(ctx context.Context) ([]Contact, error) {
	stored, err := s.client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Contact, 0, len(stored))
	for jid, c := range stored {
		result = append(result, Contact{
			JID:          jid.String(),
			Name:         c.FullName,
			PushName:     c.PushName,
			FirstName:    c.FirstName,
			BusinessName: c.BusinessName,
			Found:        c.Found,
		})
	}
	return result, nil
}

// Group is the JSON shape of a joined group.
type Group struct {
	JID          string `json:"jid"`
	Name         string `json:"name"`
	OwnerJID     string `json:"owner_jid"`
	CreatedAt    int64  `json:"created_at"`
	Participants int    `json:"participants"`
}

// Groups returns every group the account has joined.
func (s *Service) Groups(ctx context.Context) ([]Group, error) {
	joined, err := s.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Group, 0, len(joined))
	for _, g := range joined {
		result = append(result, Group{
			JID:          g.JID.String(),
			Name:         g.Name,
			OwnerJID:     g.OwnerJID.String(),
			CreatedAt:    g.GroupCreated.Unix(),
			Participants: g.ParticipantCount,
		})
	}
	return result, nil
}

// MarkRead sends read receipts for the given messages of a chat.
func (s *Service) MarkRead(ctx context.Context, jid string, messageIDs []string) error {
	target, err := types.ParseJID(appcfg.NormalizeJID(jid))
	if err != nil {
		return err
	}

	own := s.client.Store.ID
	if own == nil {
		return errNotPaired
	}

	ids := make([]types.MessageID, len(messageIDs))
	for i, id := range messageIDs {
		ids[i] = types.MessageID(id)
	}

	return s.client.MarkRead(ctx, ids, time.Now(), target, *own)
}

// MuteChat silences (or unsilences) a chat for a bounded duration.
//
// SESSİZLEŞTİRME BİR APP-STATE YAMASIDIR, mesaj değil: telefon uygulaması
// bildirimleri susturur, mesajlar hâlâ gelir. Bu yüzden whatsmeow'da
// SendAppState + appstate.BuildMute kullanılır.
//
// Süre: 0 = kalıcı sustur (BuildMute sonsuza kadar). WhatsApp'ta kalıcı
// susturma yok; "süresiz" her zaman "1 yıl" gibi büyük bir süreye çevrilir
// çünkü sunucu -1'i "kalıcı" olarak yorumlar ama bazı istemciler bunu
// "sıfır süre" diye okuyup susturmayı KALDIRIR. 1 yıl güvenli tarafta.
func (s *Service) MuteChat(ctx context.Context, jid string, muted bool, duration time.Duration) error {
	target, err := sohbetJID(jid)
	if err != nil {
		return err
	}

	if muted && duration <= 0 {
		// Sessizlik "kalıcı" istenmiş. Sonsuz süre yerine 1 yıl: WhatsApp
		// kendi arayüzünde de kalıcı seçenek yok, kullanıcı 8 saat / 1 hafta
		// / 1 yıl seçer.
		duration = 365 * 24 * time.Hour
	}
	if !muted {
		duration = 0
	}

	return s.client.SendAppState(ctx, appstate.BuildMute(target, muted, duration))
}

// StarMessage marks (or unmarks) a message as a favourite.
//
// YÖNLENDİRME: mesajın gönderen JID'i gerekir. LID sohbetlerinde gönderen
// alanı boş olabilir; bu durumda mesajın From'u kullanılır. İkisi de
// normalize edilir, aksi hâlde telefon uygulamasında yıldız görünmez
// (sessiz başarısızlık — okunmuyor, hata da vermiyor).
func (s *Service) StarMessage(ctx context.Context, jid, senderJID, messageID string, fromMe, starred bool) error {
	target, err := sohbetJID(jid)
	if err != nil {
		return err
	}

	// Gönderen boşsa hedefi kullan: birebir sohbette (bireysel mesaj)
	// gönderen hedefle aynıdır. Gruptaysa yanlış olur, o yüzden
	// çağıran (HTTP katmanı) grup durumunu doğrulamalı.
	if strings.TrimSpace(senderJID) == "" {
		senderJID = jid
	}
	sender, err := sohbetJID(senderJID)
	if err != nil {
		return err
	}

	if strings.TrimSpace(messageID) == "" {
		return errMessageIDRequired
	}

	return s.client.SendAppState(ctx,
		appstate.BuildStar(target, sender, types.MessageID(messageID), fromMe, starred))
}
