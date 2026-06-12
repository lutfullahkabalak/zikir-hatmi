import SwiftUI

struct HomeView: View {
    let onCreateTapped: () -> Void

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 12) {
                    Text("ZIKIR HATMI")
                        .soraFont(11, weight: .medium)
                        .tracking(4)
                        .foregroundStyle(RamadanTheme.mutedText)

                    Text("Dinamik hatim paylaşımı")
                        .soraFont(32, weight: .semibold)
                        .multilineTextAlignment(.center)

                    Text("Sağ üstteki + ile yeni hatim oluşturun, bağlantıyı paylaşın ve birlikte zikre devam edin.")
                        .soraFont(16)
                        .foregroundStyle(RamadanTheme.mutedText)
                        .multilineTextAlignment(.center)
                }
                .padding(.top, 40)

                GlassCard {
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Nasıl çalışır?")
                            .soraFont(18, weight: .semibold)

                        VStack(alignment: .leading, spacing: 8) {
                            Text("• Hatim oluşturun ve paylaşım linkini kopyalayın.")
                            Text("• Katılımcılar linke girip gerekirse şifreyle katılır.")
                            Text("• Sayım herkes için gerçek zamanlı güncellenir.")
                        }
                        .soraFont(14)
                        .foregroundStyle(RamadanTheme.mutedText)
                    }
                }
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 40)
        }
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button {
                    onCreateTapped()
                } label: {
                    Image(systemName: "plus")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(.black)
                        .frame(width: 36, height: 36)
                        .background(.white, in: Circle())
                }
            }
        }
    }
}
