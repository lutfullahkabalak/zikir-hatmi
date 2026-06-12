import SwiftUI

struct PresenceSheet: View {
    @Binding var isPresented: Bool
    let users: [PresenceUser]

    var body: some View {
        NavigationStack {
            ZStack {
                RamadanBackground()
                ScrollView {
                    VStack(alignment: .leading, spacing: 16) {
                        Text("AKTIF KULLANICILAR")
                            .soraFont(11, weight: .medium)
                            .tracking(3)
                            .foregroundStyle(RamadanTheme.mutedText)

                        Text("Şu an bağlı: \(users.count)")
                            .soraFont(22, weight: .semibold)

                        Text("Kullanıcı adını menüden ayarlayabilirsin.")
                            .soraFont(14)
                            .foregroundStyle(RamadanTheme.mutedText)

                        ForEach(users) { user in
                            HStack(spacing: 12) {
                                InitialsAvatar(name: user.name, size: 36)
                                Text(user.name)
                                    .soraFont(15, weight: .medium)
                                    .lineLimit(1)
                                Spacer()
                            }
                            .padding(.horizontal, 12)
                            .padding(.vertical, 10)
                            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
                            .overlay {
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(RamadanTheme.glassBorder, lineWidth: 1)
                            }
                        }
                    }
                    .padding(24)
                }
            }
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Kapat") { isPresented = false }
                }
            }
        }
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
    }
}
