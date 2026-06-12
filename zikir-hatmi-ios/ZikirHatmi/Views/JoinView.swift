import SwiftUI

struct JoinView: View {
    @State private var viewModel: JoinViewModel
    let onJoined: (String) -> Void

    init(shareCode: String, onJoined: @escaping (String) -> Void) {
        _viewModel = State(initialValue: JoinViewModel(shareCode: shareCode))
        self.onJoined = onJoined
    }

    var body: some View {
        ScrollView {
            VStack {
                Spacer(minLength: 60)

                GlassCard {
                    VStack(alignment: .leading, spacing: 16) {
                        Text("ZIKIR HATMI")
                            .soraFont(11, weight: .medium)
                            .tracking(3)
                            .foregroundStyle(RamadanTheme.mutedText)

                        Text("Şifre ile katıl")
                            .soraFont(24, weight: .semibold)

                        Text("Bu hatime katılmak için şifre girmeniz gerekiyor.")
                            .soraFont(14)
                            .foregroundStyle(RamadanTheme.mutedText)

                        SecureField("Hatim şifresi", text: $viewModel.password)
                            .padding()
                            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
                            .overlay {
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(RamadanTheme.glassBorder, lineWidth: 1)
                            }

                        Button {
                            Task {
                                if let token = await viewModel.submit() {
                                    onJoined(token)
                                }
                            }
                        } label: {
                            if viewModel.loading {
                                ProgressView()
                                    .frame(maxWidth: .infinity)
                            } else {
                                Text("Katıl")
                                    .frame(maxWidth: .infinity)
                            }
                        }
                        .buttonStyle(.borderedProminent)
                        .tint(RamadanTheme.primaryAccent)

                        if let errorMessage = viewModel.errorMessage {
                            Text(errorMessage)
                                .soraFont(14)
                                .foregroundStyle(RamadanTheme.error)
                        }
                    }
                }
                .padding(.horizontal, 20)

                Spacer()
            }
        }
        .navigationBarTitleDisplayMode(.inline)
    }
}
