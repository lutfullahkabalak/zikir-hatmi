import SwiftUI

struct UsernameSheet: View {
    @Binding var isPresented: Bool
    @State private var draft = ""
    let onSave: (String) -> Void
    let onClear: () -> Void

    var body: some View {
        NavigationStack {
            ZStack {
                RamadanBackground()
                VStack(alignment: .leading, spacing: 16) {
                    Text("KULLANICI")
                        .soraFont(11, weight: .medium)
                        .tracking(3)
                        .foregroundStyle(RamadanTheme.mutedText)

                    Text("Kullanıcı adın")
                        .soraFont(22, weight: .semibold)

                    Text("Opsiyonel. Bu cihazda saklanır ve hatimde aktif kullanıcılar arasında görünür.")
                        .soraFont(14)
                        .foregroundStyle(RamadanTheme.mutedText)

                    TextField("Adınız (ör. Ahmet)", text: $draft)
                        .padding()
                        .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
                        .overlay {
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(RamadanTheme.glassBorder, lineWidth: 1)
                        }

                    HStack(spacing: 12) {
                        Button("Kaydet") {
                            onSave(draft)
                            isPresented = false
                        }
                        .buttonStyle(.borderedProminent)
                        .tint(RamadanTheme.primaryAccent)
                        .frame(maxWidth: .infinity)

                        Button("Temizle") {
                            onClear()
                            draft = ""
                            isPresented = false
                        }
                        .buttonStyle(.bordered)
                        .frame(maxWidth: .infinity)
                    }

                    Spacer()
                }
                .padding(24)
            }
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Kapat") { isPresented = false }
                }
            }
            .onAppear {
                draft = UserPreferences.shared.username
            }
        }
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
    }
}
