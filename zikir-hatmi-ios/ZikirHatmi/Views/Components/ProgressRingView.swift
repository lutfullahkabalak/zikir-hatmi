import SwiftUI

struct ProgressRingView: View {
    let count: Int
    let target: Int
    let isDisabled: Bool
    let onTap: () -> Void

    private let radius: CGFloat = 140
    private let lineWidth: CGFloat = 18

    private var progressRatio: Double {
        guard target > 0 else { return 0 }
        return min(Double(count) / Double(target), 1)
    }

    var body: some View {
        ZStack {
            Circle()
                .stroke(Color.white.opacity(0.1), lineWidth: lineWidth)
                .frame(width: radius * 2, height: radius * 2)

            Circle()
                .trim(from: 0, to: progressRatio)
                .stroke(
                    RamadanTheme.progressColor(ratio: progressRatio),
                    style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
                )
                .rotationEffect(.degrees(-90))
                .frame(width: radius * 2, height: radius * 2)
                .animation(.easeInOut(duration: 0.5), value: progressRatio)

            Button(action: onTap) {
                Circle()
                    .fill(RamadanTheme.primaryAccent)
                    .frame(width: radius * 1.85, height: radius * 1.85)
                    .shadow(color: RamadanTheme.primaryAccent.opacity(0.3), radius: 20, y: 8)
            }
            .buttonStyle(.plain)
            .disabled(isDisabled)
            .opacity(isDisabled ? 0.5 : 1)
        }
        .frame(width: 320, height: 320)
    }
}
