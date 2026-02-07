# Özelleştirme Planı

1. Gorilla WebSocket kütüphanesini projeye ekle ve `go.mod` dosyasında modülü tanımla.
2. İstemci bağlantılarını ve yayın (broadcast) mantığını yöneten basit bir `hub` oluştur.
3. `/ws` uç noktasını HTTP sunucusuna kaydederek bağlantıları WebSocket'e yükselt ve gelen mesajları diğer istemcilere dağıt.
4. Sağlık kontrolü için `GET /healthz` uç noktasını sun ve sunucu için nazik kapatma (graceful shutdown) akışı kur.
5. Kurulum, çalışma ve örnek istemci adımlarını açıklayan bir README hazırla.
