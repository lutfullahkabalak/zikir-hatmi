# Zikir Hatmi Frontend

Vue 3 + Nuxt UI tabanlı bu arayüz, hatim halkalarını tek panel üzerinden yönetebilmek için örnek bir landing sayfası sunar. Proje Vite ile derlenir, Nuxt UI'nin Vue eklentisi ve TailwindCSS ile temalandırılır.

## Teknoloji Yığını
- Vue 3 + TypeScript + Vite (Rolldown tabanlı derleyici)
- vue-router 4 (Nuxt UI bileşenleri için opsiyonel peer)
- Nuxt UI (Vue eklentisi ve Vite plugini)
- TailwindCSS 4 + @tailwindcss/vite

## Başlangıç
1. Bağımlılıkları kurun: `npm install`
2. Geliştirme sunucusunu açın: `npm run dev`
3. Üretim yapısı almak için `npm run build`, önizlemek için `npm run preview`

## Nuxt UI Entegrasyonu
- `src/main.ts` dosyasında `@nuxt/ui/vue-plugin` global olarak kaydedilir.
- `vite.config.ts` içinde `@nuxt/ui/vite` plugini etkin.
- `src/assets/main.css` dosyası Tailwind direktifleri ile birlikte `@nuxt/ui` stillerini içe aktarır.
- `src/App.vue` bileşeni `UApp`, `UCard`, `UButton`, `UInput`, `USelectMenu`, `UAlert`, `UBadge`, `UIcon` gibi bileşenleri örnekler.

## Tailwind Ayarları
- `tailwind.config.ts` dosyası yazı tiplerini ve özel aurora arka plan gradyanını genişletir.
- `@tailwindcss/vite` plugini `vite.config.ts` içinde etkin durumdadır; `src/assets/main.css` dosyasında `@import "tailwindcss";` + `@import "@nuxt/ui";` sıralaması kullanılır.

## Sonraki Adımlar
- Nuxt UI bileşenlerini sayfalarınıza parçalı olarak taşıyın.
- Tailwind temalandırmasını marka renkleriyle güncelleyin.
- Gerçek form uç noktaları ekleyerek `handleSubscribe` fonksiyonunu backend ile bütünleştirin.
