## Plan: Zikir Hatmi Gerçek Zamanlı Sayaç

Mevcut Go backend'deki WebSocket altyapısını sayaç durumu yönetecek şekilde genişletip, Vue frontend'i tek ekranlı bir zikir sayacına dönüştüreceğiz. Bir kullanıcı butona bastığında sayı artacak ve WebSocket üzerinden tüm bağlı istemcilere anında yansıyacak.

### Steps

1. **Backend'e sayaç durumu ekle** — `main.go` içindeki `hub` struct'ına `count int` ve `target int` alanları ekle. Yeni bağlanan istemcilere mevcut `{count, target}` bilgisini JSON olarak gönder. Gelen `"increment"` mesajlarında `count`'u artır ve güncel durumu **tüm** istemcilere (gönderici dahil) broadcast et. Hedefe ulaşıldığında `"completed"` durumu gönder.

2. **WebSocket mesaj protokolü tanımla** — Backend ↔ Frontend arası JSON mesajlar: sunucudan → `{"type":"state","count":42,"target":1000}` ve `{"type":"completed"}`, istemciden → `{"type":"increment"}`. Struct'ları `Message` olarak `main.go`'da tanımla.

3. **Frontend WebSocket composable yaz** — `src/composables/useZikirSocket.ts` dosyası oluştur. `ref(count)`, `ref(target)`, `ref(connected)` reactive değişkenlerini tut. Otomatik bağlantı, yeniden bağlanma (exponential backoff), ve `increment()` fonksiyonu sağla.

4. **Zikir sayaç ekranını oluştur** — `App.vue` içeriğini değiştir: ortada SVG `<circle>` ile dairesel progress bar, onun içinde sayı (`count`) ve artırma butonu (`UButton`), altta küçük yazıyla hedef sayı (`target`). Tailwind CSS ile koyu tema, ortalanmış düzen. Progress bar doluluk oranını `count/target` ile hesapla ve SVG `stroke-dashoffset` ile animasyonlu göster.

5. **CORS ve Vite proxy ayarla** — `vite.config.ts`'de `/ws` yolunu `localhost:8080`'e proxy'le (`ws: true`). Backend'deki `CheckOrigin` zaten açık, geliştirme için yeterli.

### Further Considerations

1. **Hedef sayı nereden gelecek?** Backend'de sabit kodlanmış (ör. 1000) mı olsun, yoksa başlangıçta kullanıcının belirleyebileceği bir input mu ekleyelim?
2. **Sayaç sıfırlama** — Hedefe ulaşıldığında otomatik sıfırlansın mı, yoksa bir "Yeni Hatim" butonu mu gösterelim?
3. **Kalıcı depolama** — Şu an in-memory olacak (sunucu kapanırsa sayaç sıfırlanır). İleride SQLite/dosya ile kalıcılık gerekir mi?
