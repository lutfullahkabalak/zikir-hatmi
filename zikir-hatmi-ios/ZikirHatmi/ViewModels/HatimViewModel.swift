import Foundation
import Observation

@Observable
@MainActor
final class HatimViewModel {
    let shareCode: String
    var showCreatedBanner: Bool

    var hatimTitle = ""
    var loading = true
    var joinLoading = false
    var errorMessage: String?
    var requiresPasswordJoin = false
    var token: String?

    let socket = ZikirWebSocket()

    init(shareCode: String, showCreatedBanner: Bool = false) {
        self.shareCode = shareCode
        self.showCreatedBanner = showCreatedBanner
        socket.configure {
            UserPreferences.shared.username
        }
    }

    var isCompleted: Bool {
        socket.target > 0 && socket.count >= socket.target
    }

    var shareURL: URL {
        AppConfig.shareURL(shareCode: shareCode)
    }

    func load() async {
        loading = true
        errorMessage = nil
        requiresPasswordJoin = false
        socket.disconnect()

        do {
            let state = try await APIClient.shared.getHatim(shareCode: shareCode)
            hatimTitle = state.title
            socket.setState(count: state.count, target: state.target)

            if let storedToken = TokenStore.load(for: shareCode) {
                token = storedToken
                socket.connect(shareCode: shareCode, token: storedToken)
                loading = false
                return
            }

            if state.requiresPassword {
                requiresPasswordJoin = true
                loading = false
                return
            }

            joinLoading = true
            let join = try await APIClient.shared.joinHatim(shareCode: shareCode, password: nil)
            try TokenStore.save(token: join.token, for: shareCode)
            token = join.token
            socket.connect(shareCode: shareCode, token: join.token)
        } catch let error as APIError {
            switch error {
            case .httpError(let status) where status == 404:
                errorMessage = "Hatim bulunamadı."
            default:
                errorMessage = "Hatim bilgileri alınamadı."
            }
        } catch {
            errorMessage = "Hatim bilgileri alınamadı."
        }

        joinLoading = false
        loading = false
    }

    func reloadAfterJoin() async {
        await load()
    }

    func increment(amount: Int = 1) {
        HapticService.tapIfEnabled()
        socket.increment(amount: amount)
    }

    func updateUsername(_ name: String) {
        UserPreferences.shared.setUsername(name)
        socket.sendUsernameUpdate(name)
    }

    func dismissCreatedBanner() {
        showCreatedBanner = false
    }

    func teardown() {
        socket.disconnect()
    }
}
