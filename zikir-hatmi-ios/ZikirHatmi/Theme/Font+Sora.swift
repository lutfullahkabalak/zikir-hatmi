import SwiftUI

extension Font {
    static func sora(_ size: CGFloat, weight: Font.Weight = .regular) -> Font {
        .custom("Sora", size: size).weight(weight)
    }
}

extension View {
    func soraFont(_ size: CGFloat, weight: Font.Weight = .regular) -> some View {
        font(.sora(size, weight: weight))
    }
}
