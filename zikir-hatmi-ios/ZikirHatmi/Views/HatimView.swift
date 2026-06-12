import SwiftUI

struct HatimView: View {
    @State private var viewModel: HatimViewModel
    @State private var bulkIncrementOpen = false
    @State private var presenceOpen = false
    @State private var usernameOpen = false
    @State private var shareOpen = false
    @State private var createOpen = false
    let onNavigateToJoin: () -> Void
    let onNavigateToHatim: (String, Bool) -> Void

    init(
        shareCode: String,
        showCreatedBanner: Bool = false,
        onNavigateToJoin: @escaping () -> Void,
        onNavigateToHatim: @escaping (String, Bool) -> Void
    ) {
        _viewModel = State(initialValue: HatimViewModel(shareCode: shareCode, showCreatedBanner: showCreatedBanner))
        self.onNavigateToJoin = onNavigateToJoin
        self.onNavigateToHatim = onNavigateToHatim
    }

    private var maxVisibleUsers: Int { 4 }

    private var visibleUsers: [PresenceUser] {
        Array(viewModel.socket.activeUsers.prefix(maxVisibleUsers))
    }

    private var hasMoreUsers: Bool {
        viewModel.socket.activeUsers.count > maxVisibleUsers
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            ScrollView {
                VStack(spacing: 24) {
                    if viewModel.showCreatedBanner {
                        createdBanner
                    }

                    if let errorMessage = viewModel.errorMessage {
                        GlassCard {
                            Text(errorMessage)
                                .soraFont(14)
                                .foregroundStyle(RamadanTheme.error)
                        }
                    } else if viewModel.requiresPasswordJoin {
                        GlassCard {
                            VStack(alignment: .leading, spacing: 12) {
                                Text("Bu hatim şifre korumalı.")
                                    .soraFont(16, weight: .semibold)
                                Button("Şifre ile katıl") {
                                    onNavigateToJoin()
                                }
                                .buttonStyle(.borderedProminent)
                                .tint(RamadanTheme.primaryAccent)
                            }
                        }
                    } else {
                        counterContent
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 56)
                .padding(.bottom, 40)
            }

            Button {
                bulkIncrementOpen = true
            } label: {
                Image(systemName: "plus")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(.black)
                    .frame(width: 36, height: 36)
                    .background(.white, in: Circle())
            }
            .padding(.leading, 16)
            .padding(.top, 8)
            .disabled(viewModel.loading || viewModel.joinLoading || viewModel.requiresPasswordJoin || !viewModel.socket.connected || viewModel.isCompleted)
            .opacity(viewModel.loading || viewModel.joinLoading || viewModel.requiresPasswordJoin || !viewModel.socket.connected || viewModel.isCompleted ? 0.5 : 1)
        }
        .navigationTitle(viewModel.hatimTitle)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button {
                        shareOpen = true
                    } label: {
                        Label("Paylaş", systemImage: "square.and.arrow.up")
                    }

                    Button {
                        usernameOpen = true
                    } label: {
                        Label("Kullanıcı adı", systemImage: "person")
                    }

                    Toggle(isOn: Binding(
                        get: { UserPreferences.shared.hapticsEnabled },
                        set: { UserPreferences.shared.hapticsEnabled = $0 }
                    )) {
                        Label("Titreşim", systemImage: "iphone.radiowaves.left.and.right")
                    }

                    Button {
                        createOpen = true
                    } label: {
                        Label("Yeni hatim", systemImage: "plus.circle")
                    }
                } label: {
                    Image(systemName: "ellipsis")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(.black)
                        .frame(width: 36, height: 36)
                        .background(.white, in: Circle())
                }
            }
        }
        .task {
            await viewModel.load()
        }
        .onDisappear {
            viewModel.teardown()
        }
        .sheet(isPresented: $bulkIncrementOpen) {
            BulkIncrementSheet(
                isPresented: $bulkIncrementOpen,
                isCompleted: viewModel.isCompleted
            ) { amount in
                viewModel.increment(amount: amount)
            }
        }
        .sheet(isPresented: $presenceOpen) {
            PresenceSheet(isPresented: $presenceOpen, users: viewModel.socket.activeUsers)
        }
        .sheet(isPresented: $usernameOpen) {
            UsernameSheet(isPresented: $usernameOpen) { name in
                viewModel.updateUsername(name)
            } onClear: {
                viewModel.updateUsername("")
            }
        }
        .sheet(isPresented: $shareOpen) {
            ShareSheet(items: [viewModel.shareURL])
        }
        .sheet(isPresented: $createOpen) {
            CreateHatimSheet(isPresented: $createOpen) { response in
                onNavigateToHatim(response.shareCode, true)
            }
        }
    }

    private var createdBanner: some View {
        GlassCard {
            VStack(alignment: .leading, spacing: 12) {
                Text("Paylaşım bağlantısı hazır")
                    .soraFont(14, weight: .semibold)
                Text(viewModel.shareURL.absoluteString)
                    .soraFont(12)
                    .foregroundStyle(RamadanTheme.mutedText)

                HStack {
                    Button("Linki kopyala") {
                        UIPasteboard.general.string = viewModel.shareURL.absoluteString
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(RamadanTheme.primaryAccent)

                    Button("Kapat") {
                        viewModel.dismissCreatedBanner()
                    }
                    .buttonStyle(.bordered)
                }
            }
        }
    }

    private var counterContent: some View {
        VStack(spacing: 24) {
            Text("\(viewModel.socket.count)")
                .soraFont(56, weight: .semibold)
                .monospacedDigit()

            ProgressRingView(
                count: viewModel.socket.count,
                target: viewModel.socket.target,
                isDisabled: viewModel.loading || viewModel.joinLoading || !viewModel.socket.connected || viewModel.isCompleted
            ) {
                viewModel.increment()
            }

            VStack(spacing: 8) {
                Text("Hedef: \(viewModel.socket.target)")
                    .soraFont(14)
                    .foregroundStyle(RamadanTheme.mutedText)

                Text(viewModel.socket.connected ? "CANLI BAĞLANTI" : "BAĞLANIYOR...")
                    .soraFont(11, weight: .medium)
                    .tracking(3)
                    .foregroundStyle(RamadanTheme.mutedText)

                if viewModel.isCompleted {
                    Text("Hatim tamamlandı.")
                        .soraFont(14, weight: .semibold)
                        .foregroundStyle(RamadanTheme.completed)
                }

                if viewModel.socket.connected && !viewModel.socket.activeUsers.isEmpty {
                    Button {
                        presenceOpen = true
                    } label: {
                        HStack(spacing: -8) {
                            ForEach(visibleUsers) { user in
                                InitialsAvatar(name: user.name, size: 28)
                            }
                            if hasMoreUsers {
                                Text("…")
                                    .frame(width: 28, height: 28)
                                    .background(.ultraThinMaterial, in: Circle())
                            }
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 8)
                        .background(.ultraThinMaterial, in: Capsule())
                        .overlay {
                            Capsule().stroke(RamadanTheme.glassBorder, lineWidth: 1)
                        }
                    }
                    .buttonStyle(.plain)
                    .padding(.top, 8)
                }
            }
        }
        .frame(maxWidth: .infinity)
    }
}
