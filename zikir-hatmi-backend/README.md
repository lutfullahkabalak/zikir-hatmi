# Zikir Hatmi Backend

Sıfırdan oluşturulmuş bu Go projesi, Gorilla WebSocket kütüphanesini kullanarak gerçek zamanlı iletişim sağlar. `/ws` uç noktasına bağlanan her istemci, gönderdiği mesajları diğer açık bağlantılarla paylaşır.

## Gereksinimler
- Go 1.22+

## Çalıştırma
```bash
go mod tidy
PORT=8080 go run .
```
PORT değeri verilmezse varsayılan olarak `8080` kullanılır.

## HTTP Uç Noktaları
- `GET /healthz`: Servisin ayakta olduğunu bildiren basit JSON cevap.
- `GET /ws`: WebSocket bağlantısı sağlar.

## Örnek İstemci
Tarayıcı konsolundan veya küçük bir HTML dosyasında aşağıdaki örneği kullanabilirsiniz:

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => ws.send("Selamlar!");
ws.onmessage = (event) => console.log("Mesaj:", event.data);
ws.onclose = () => console.log("Bağlantı kapandı");
```

## Yol Haritası
- Kimlik doğrulama ve yetkilendirme ekleme
- Mesaj geçmişini kalıcı bir depoda saklama
- Otomatik testler ve dağıtım betikleri hazırlama
