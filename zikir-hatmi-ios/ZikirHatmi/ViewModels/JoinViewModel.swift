import Observation

@Observable
@MainActor
final class JoinViewModel {
    let shareCode: String

    var password = ""
    var loading = false
    var errorMessage: String?

    init(shareCode: String) {
        self.shareCode = shareCode
    }

    func submit() async -> String? {
        loading = true
        errorMessage = nil
        defer { loading = false }

        do {
            let response = try await APIClient.shared.joinHatim(shareCode: shareCode, password: password)
            try TokenStore.save(token: response.token, for: shareCode)
            return response.token
        } catch let error as JoinError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = JoinError.generic.errorDescription
        }
        return nil
    }
}
