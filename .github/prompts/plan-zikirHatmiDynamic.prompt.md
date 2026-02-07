## Plan: Dinamik hatim ve paylaşım akışı

Hatimleri PostgreSQL’de saklayıp her hatime özel WebSocket bağlantısı kuracağız. Kullanıcılar sağ üstteki + ile hatim oluşturacak, şifre opsiyonel olacak; oluşturulunca paylaşım linki üretilecek ve kullanıcı doğrudan sayaç ekranına yönlenecek. Sol üstteki paylaş butonu linki kopyalayacak. Bu, kalıcılık, erişim kontrolü ve çoklu hatim desteği sağlar.

### Steps
1. Backend’de hatim modeli ve DB erişimi tanımla; `hatim` tablosu, `CreateHatim`/`GetHatim`/`JoinHatim` servislerini `main.go` ve yeni db dosyalarında yapılandır.
2. REST uçları ekle: `POST /hatims` (oluştur), `POST /hatims/{id}/join` (şifre doğrula), `GET /hatims/{id}` (durum), ve `GET /ws/{id}?token=` per-hatim WebSocket; `hub`’ı `map[hatimId]*hub` olacak şekilde genişlet.
3. Frontend’de router ve sayfaları oluştur: oluşturma modalı, hatim ekranı, şifre giriş ekranı; `App.vue`’yu yalnızca shell + router-view yap; bağlantı akışını güncelle.
4. `useZikirSocket`’ı `hatimId` ve `token` alacak şekilde güncelle; state’i REST’ten başlatıp WebSocket’le canlı güncelle.
5. UI güncellemeleri: sağ üst `+` butonu ile oluşturma modalı; sol üst paylaş butonu ile link kopyalama; oluşturma sonrası otomatik yönlendirme ve link gösterimi.

### Further Considerations
1. URL yapısı ne olsun? `/h/{shareCode}` mi, `/hatim/{id}` mi? `/h/{shareCode}` olsun.
2. Şifre kontrolü: sadece girişte mi, her socket bağlantısında mı doğrulansın? ilk bağlantıda doğrulansın, sonrasında token ile erişim sağlansın.
3. Şifre hash algoritması tercihi: bcrypt mi argon2id mi? argon2id daha güvenli ve modern, onu kullanabiliriz.
