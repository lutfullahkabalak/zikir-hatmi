import UIKit

enum HapticService {
    private static let generator = UIImpactFeedbackGenerator(style: .light)

    static func tapIfEnabled() {
        guard UserPreferences.shared.hapticsEnabled else { return }
        generator.prepare()
        generator.impactOccurred()
    }
}
