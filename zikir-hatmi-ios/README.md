# Zikir Hatmi — iOS

Native SwiftUI uygulaması. Web frontend ile aynı backend API ve WebSocket protokolünü kullanır.

## Gereksinimler

- Xcode 16+ (iOS 17 deployment target)
- Apple Developer hesabı (cihaz/TestFlight/App Store için)

## Projeyi açma

```bash
cd zikir-hatmi-ios
open ZikirHatmi.xcodeproj
```

Alternatif: XcodeGen ile yeniden üretmek için:

```bash
xcodegen generate
```

## Yapılandırma

| Ayar | Dosya | Değer |
|------|-------|-------|
| API base URL | `ZikirHatmi/AppConfig.swift` | `https://zikirhatmiapi.abapnews.info` |
| Paylaşım linki | `ZikirHatmi/AppConfig.swift` | `https://zikirhatmi.abapnews.tr/h/{shareCode}` |
| Bundle ID | `project.yml` | `com.zikirhatmi.app` |

## Signing

1. Xcode → Target **ZikirHatmi** → **Signing & Capabilities**
2. Team seçin (Apple Developer Team ID)
3. **Associated Domains** capability'sinde `applinks:zikirhatmi.abapnews.tr` olmalı
4. `zikir-hatmi-frontend/public/.well-known/apple-app-site-association` dosyasındaki `TEAMID` değerini kendi Team ID'nizle değiştirin ve frontend'i deploy edin

Universal Link doğrulama:

```bash
curl -s https://zikirhatmi.abapnews.tr/.well-known/apple-app-site-association
```

Custom URL scheme fallback: `zikirhatmi://h/{shareCode}`

## Özellikler

- Ana sayfa, hatim oluşturma, şifreli katılım
- Gerçek zamanlı sayaç (WebSocket)
- Progress ring, toplu ekleme, aktif katılımcılar
- Kullanıcı adı (UserDefaults)
- Hatim token (Keychain)
- Opsiyonel haptic feedback (sayaç menüsünden)
- Sistem light/dark tema + glass (`.ultraThinMaterial`) tasarım

## TestFlight → App Store

1. App Store Connect'te uygulama oluşturun (`com.zikirhatmi.app`)
2. Xcode → **Product → Archive**
3. **Distribute App** → App Store Connect / TestFlight
4. Gizlilik politikası URL'si ve ekran görüntülerini ekleyin
5. App icon: `Resources/Assets.xcassets/AppIcon.appiconset` (1024×1024 gerekli)

## Mimari

```
Views → ViewModels → APIClient / ZikirWebSocket
                   → TokenStore (Keychain)
                   → UserPreferences (UserDefaults)
```

WebSocket mesaj protokolü web frontend ile birebir aynıdır (`useZikirSocket.ts`).
