# Zikir Hatmi (Monorepo)

Bu repo iki parçadan oluşur:
- **Backend**: Go + PostgreSQL, hatim oluşturma/join + her hatime özel WebSocket
- **Frontend**: Vue 3 + Vite + Nuxt UI, hatim ekranı ve gerçek zamanlı sayaç

## Klasörler
- `zikir-hatmi-backend/`
- `zikir-hatmi-frontend/`

## Hızlı Başlangıç (Docker Compose)
Gereksinim: Docker Desktop

```bash
docker compose up -d --build
```

Servisler:
- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`
- Postgres: `localhost:5432`

Durdurmak için:
```bash
docker compose down
```

## Lokal Çalıştırma (Docker’suz)
### 1) Postgres
Bir Postgres instance’ı çalıştırın ve aşağıdaki gibi bir bağlantı verin:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/zikirhatmi?sslmode=disable'
```

### 2) Backend
```bash
cd zikir-hatmi-backend
go mod tidy
PORT=8080 DATABASE_URL="$DATABASE_URL" go run .
```

### 3) Frontend
```bash
cd zikir-hatmi-frontend
npm ci
npm run dev
```

> Frontend, geliştirme modunda `/hatims` ve `/ws` isteklerini Vite proxy ile backend’e yönlendirir.

## Uç Noktalar (Backend)
- `POST /hatims` → hatim oluşturur (opsiyonel şifre + başlık)
- `GET /hatims/{shareCode}` → hatim durumunu getirir
- `POST /hatims/{shareCode}/join` → token üretir (şifre gerekiyorsa doğrular)
- `GET /ws/{shareCode}?token=...` → hatime özel WebSocket

## GitHub Actions
Workflow dosyası: `.github/workflows/ci.yml`

- Push/PR: Go build/test + Frontend build
- **Manuel**: Docker image build/push (GHCR)
  - GitHub → Actions → **CI** → **Run workflow**
  - `tag` input’u ile image tag seçebilirsiniz (default `latest`)

### Raspberry Pi (Multi-arch Images)
Manuel job, şu platformları build eder:
- `linux/arm64` (Raspberry Pi 64-bit)
- `linux/arm/v7` (Raspberry Pi OS 32-bit)
- `linux/amd64` (x86_64)

GHCR image isimleri:
- `ghcr.io/<owner>/<repo>-backend:<tag>`
- `ghcr.io/<owner>/<repo>-frontend:<tag>`

## Portainer / Üretim (GHCR Imajları)
Üretim ortamında (örn. Raspberry Pi üstünde Portainer) imajları build etmek yerine GHCR'den çekmek için `docker-compose.portainer.yml` kullanılabilir.

Portainer → **Stacks** → **Add stack** → **Repository** veya **Web editor** ile bu dosyayı yükleyin.

İsteğe bağlı environment variable'lar:
- `IMAGE_TAG` (default: `latest`)
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `BACKEND_PORT` (default: `8998`), `FRONTEND_PORT` (default: `8999`)

Imajlar:
- `ghcr.io/lutfullahkabalak/zikir-hatmi-backend:<tag>`
- `ghcr.io/lutfullahkabalak/zikir-hatmi-frontend:<tag>`

Stack, servisleri sırayla ayağa kaldırır: önce `db` (healthcheck geçene kadar beklenir), sonra `backend` (healthcheck geçene kadar beklenir), en son `frontend`. Ayrıca backend DB'ye bağlanamazsa hemen düşmez, db hazır olana kadar geri çekilerek tekrar dener.

## Notlar
- Üretimde çalıştırırken `DATABASE_URL` değerini güvenli şekilde sağlayın.
- Port çakışması yaşarsanız (örn. 8080/5173), o portu kullanan süreci kapatın veya compose portlarını değiştirin.
