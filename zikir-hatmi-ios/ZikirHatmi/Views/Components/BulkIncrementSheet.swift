import SwiftUI

struct BulkIncrementSheet: View {
    @Binding var isPresented: Bool
    @State private var amountText = ""
    let isCompleted: Bool
    let onApply: (Int) -> Void

    var body: some View {
        NavigationStack {
            ZStack {
                RamadanBackground()
                VStack(alignment: .leading, spacing: 20) {
                    Text("TOPLU EKLEME")
                        .soraFont(11, weight: .medium)
                        .tracking(3)
                        .foregroundStyle(RamadanTheme.mutedText)

                    Text("Zikire kaç tane eklenecek?")
                        .soraFont(22, weight: .semibold)

                    TextField("Örn: 10", text: $amountText)
                        .keyboardType(.numberPad)
                        .padding()
                        .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: 12))
                        .overlay {
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(RamadanTheme.glassBorder, lineWidth: 1)
                        }

                    Button("Ekle") {
                        if let amount = Int(amountText), amount > 0 {
                            onApply(amount)
                            isPresented = false
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(RamadanTheme.primaryAccent)
                    .disabled(parsedAmount <= 0 || isCompleted)
                    .frame(maxWidth: .infinity, alignment: .trailing)

                    Spacer()
                }
                .padding(24)
            }
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Kapat") { isPresented = false }
                }
            }
        }
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
    }

    private var parsedAmount: Int {
        Int(amountText) ?? 0
    }
}
