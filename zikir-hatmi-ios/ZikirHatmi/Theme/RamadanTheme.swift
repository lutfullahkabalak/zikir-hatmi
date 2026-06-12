import SwiftUI

enum RamadanTheme {
    static let lightBackground = Color(red: 1.0, green: 0.961, blue: 0.969)
    static let darkBackground = Color(red: 0.031, green: 0.027, blue: 0.051)
    static let lightText = Color(red: 0.078, green: 0.078, blue: 0.078)
    static let darkText = Color(red: 0.969, green: 0.949, blue: 0.957)
    static let primaryAccent = Color.cyan
    static let completed = Color(red: 0.494, green: 0.827, blue: 0.659)
    static let error = Color(red: 0.98, green: 0.45, blue: 0.45)
    static let glassBorder = Color.white.opacity(0.1)
    static let mutedText = Color.primary.opacity(0.65)

    static func background(for scheme: ColorScheme) -> Color {
        scheme == .dark ? darkBackground : lightBackground
    }

    static func text(for scheme: ColorScheme) -> Color {
        scheme == .dark ? darkText : lightText
    }

    static func progressColor(ratio: Double) -> Color {
        let hue = 190.0 + (10.0 - 190.0) * ratio
        let saturation = 0.8
        let lightness = 0.6 - 0.08 * ratio
        return Color(hue: hue / 360.0, saturation: saturation, brightness: lightness)
    }
}

struct RamadanBackground: View {
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        RamadanTheme.background(for: colorScheme)
            .ignoresSafeArea()
    }
}
