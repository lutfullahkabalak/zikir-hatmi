import SwiftUI

struct CreateHatimSheet: View {
    @Binding var isPresented: Bool
    @State private var title = ""
    @State private var target = 50
    @State private var password = ""
    @State private var isCreating = false
    @State private var errorMessage: String?

    let onCreated: (CreateHatimResponse) -> Void

    var body: some View {
        NavigationStack {
            ZStack {
                RamadanBackground()
                ScrollView {
                    VStack(alignment: .leading, spacing: 16) {
                        Text("YENI HATIM")
                            .soraFont(11, weight: .medium)
                            .tracking(3)
                            .foregroundStyle(RamadanTheme.mutedText)

                        Text("Hatim oluştur")
                            .soraFont(22, weight: .semibold)

                        Text("Hedef belirleyin ve dilerseniz şifre ekleyin.")
                            .soraFont(14)
                            .foregroundStyle(RamadanTheme.mutedText)

                        formField("Hatim başlığı (ör. Yasin-i Şerif)", text: $title)
                        formField("Hedef (ör. 50)", text: Binding(
                            get: { String(target) },
                            set: { target = Int($0) ?? 50 }
                        ), keyboard: .numberPad)
                        SecureField("Şifre (opsiyonel)", text: $password)
                            .padding()
                            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
                            .overlay {
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(RamadanTheme.glassBorder, lineWidth: 1)
                            }

                        if let errorMessage {
                            Text(errorMessage)
                                .soraFont(14)
                                .foregroundStyle(RamadanTheme.error)
                        }

                        Button {
                            Task { await submit() }
                        } label: {
                            if isCreating {
                                ProgressView()
                                    .frame(maxWidth: .infinity)
                            } else {
                                Text("Oluştur")
                                    .frame(maxWidth: .infinity)
                            }
                        }
                        .buttonStyle(.borderedProminent)
                        .tint(RamadanTheme.primaryAccent)
                        .disabled(isCreating || title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
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
        .presentationDetents([.large])
        .presentationDragIndicator(.visible)
    }

    @ViewBuilder
    private func formField(_ placeholder: String, text: Binding<String>, keyboard: UIKeyboardType = .default) -> some View {
        TextField(placeholder, text: text)
            .keyboardType(keyboard)
            .padding()
            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
            .overlay {
                RoundedRectangle(cornerRadius: 12)
                    .stroke(RamadanTheme.glassBorder, lineWidth: 1)
            }
    }

    private func submit() async {
        isCreating = true
        errorMessage = nil
        defer { isCreating = false }

        do {
            let response = try await APIClient.shared.createHatim(
                title: title,
                target: max(1, target),
                password: password
            )
            try TokenStore.save(token: response.token, for: response.shareCode)
            isPresented = false
            onCreated(response)
        } catch {
            errorMessage = (error as? LocalizedError)?.errorDescription ?? error.localizedDescription
        }
    }
}
