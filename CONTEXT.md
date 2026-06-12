# Zikir Hatmi — Proje Context

## Genel
Monorepo: Go backend, Vue 3 frontend, SwiftUI iOS uygulaması.

## Frontend deployment (güncel)
- **Production**: multi-stage Docker build (`npm run build` → `dist/`) + **nginx:alpine**
- **Port**: container içi `80`, compose ile host `8999:80`
- **API proxy**: nginx `/hatims` ve `/ws/` isteklerini `http://backend:8080`'e yönlendirir (same-origin, `VITE_API_BASE` gerekmez)
- **SPA**: vue-router `createWebHistory` için `try_files` fallback
- **Cache**: `/assets/` immutable 1y; `index.html` no-cache
- **Universal Links**: `/.well-known/apple-app-site-association` → `application/json`

## Lokal geliştirme
- HMR için: `cd zikir-hatmi-frontend && npm run dev` (Vite proxy `/hatims`, `/ws`)
- Docker compose: prod build simülasyonu (`docker compose up`)

## Son değişiklikler
- **2026-06-12**: Frontend production'da Vite dev server yerine nginx ile statik build servis ediliyor. Yavaş yükleme / sayfa gelmeme sorunu (özellikle mobil Chrome) bu kök nedenden giderildi.
