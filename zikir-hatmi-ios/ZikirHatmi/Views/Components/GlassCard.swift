import SwiftUI

struct GlassCard<Content: View>: View {
    @ViewBuilder var content: Content

    var body: some View {
        content
            .padding(20)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .stroke(RamadanTheme.glassBorder, lineWidth: 1)
            }
    }
}

struct GlassSheetBackground: View {
    var body: some View {
        ZStack {
            Color.black.opacity(0.9)
                .ignoresSafeArea()
            Rectangle()
                .fill(.ultraThinMaterial)
                .ignoresSafeArea()
        }
    }
}

struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}

struct InitialsAvatar: View {
    let name: String
    var size: CGFloat = 32

    var body: some View {
        Text(initials(for: name))
            .soraFont(size * 0.35, weight: .semibold)
            .foregroundStyle(.white)
            .frame(width: size, height: size)
            .background(Color.cyan.opacity(0.85), in: Circle())
            .overlay {
                Circle().stroke(Color.black.opacity(0.2), lineWidth: 2)
            }
    }

    private func initials(for name: String) -> String {
        let cleaned = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !cleaned.isEmpty else { return "?" }
        let parts = cleaned.split(separator: " ").filter { !$0.isEmpty }
        let first = parts.first?.first.map(String.init) ?? ""
        let second = parts.count > 1 ? parts.last?.first.map(String.init) ?? "" : ""
        let combined = (first + second).uppercased()
        return combined.isEmpty ? String(cleaned.prefix(1)).uppercased() : combined
    }
}

func initials(for name: String) -> String {
    let cleaned = name.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !cleaned.isEmpty else { return "?" }
    let parts = cleaned.split(separator: " ").filter { !$0.isEmpty }
    let first = parts.first?.first.map(String.init) ?? ""
    let second = parts.count > 1 ? parts.last?.first.map(String.init) ?? "" : ""
    let combined = (first + second).uppercased()
    return combined.isEmpty ? String(cleaned.prefix(1)).uppercased() : combined
}
