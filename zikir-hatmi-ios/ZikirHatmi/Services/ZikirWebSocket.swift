import Foundation
import Observation

@Observable
@MainActor
final class ZikirWebSocket {
    var count = 0
    var target = 50
    var connected = false
    var activeUsers: [PresenceUser] = []

    private var webSocketTask: URLSessionWebSocketTask?
    private var receiveTask: Task<Void, Never>?
    private var reconnectTask: Task<Void, Never>?
    private var retryAttempt = 0
    private var manualClose = false
    private var hasOpened = false

    private let session = URLSession(configuration: .default)
    private let encoder = JSONEncoder()
    private let decoder = JSONDecoder()

    private var shareCode: String?
    private var token: String?
    private var usernameProvider: (() -> String)?

    func configure(usernameProvider: @escaping () -> String) {
        self.usernameProvider = usernameProvider
    }

    func connect(shareCode: String, token: String) {
        disconnect(manual: true)
        self.shareCode = shareCode
        self.token = token
        manualClose = false
        hasOpened = false
        openConnection()
    }

    func disconnect(manual: Bool = true) {
        manualClose = manual
        reconnectTask?.cancel()
        reconnectTask = nil
        receiveTask?.cancel()
        receiveTask = nil
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        connected = false
        activeUsers = []
        hasOpened = false
    }

    func increment(amount: Int = 1) {
        let nextAmount = max(1, min(1000, amount))
        send(.increment(amount: nextAmount))
    }

    func setState(count: Int, target: Int) {
        self.count = count
        self.target = target
    }

    func sendUsernameUpdate(_ name: String) {
        guard connected else { return }
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        send(.setName(name: trimmed.isEmpty ? nil : trimmed))
    }

    private func openConnection() {
        guard let shareCode, let token else { return }

        let url = AppConfig.wsURL(shareCode: shareCode, token: token)
        let task = session.webSocketTask(with: url)
        webSocketTask = task

        task.resume()
        task.sendPing { [weak self] error in
            Task { @MainActor in
                guard let self else { return }
                if error == nil {
                    self.handleOpen()
                } else {
                    self.handleDisconnect()
                }
            }
        }

        receiveTask = Task { [weak self] in
            await self?.listen()
        }
    }

    private func listen() async {
        guard let task = webSocketTask else { return }

        while !Task.isCancelled {
            do {
                let message = try await task.receive()
                handle(message)
            } catch {
                if !Task.isCancelled {
                    handleDisconnect()
                }
                break
            }
        }
    }

    private func handle(_ message: URLSessionWebSocketTask.Message) {
        let data: Data?
        switch message {
        case .data(let payload):
            data = payload
        case .string(let text):
            data = text.data(using: .utf8)
        @unknown default:
            data = nil
        }

        guard let data else { return }

        do {
            let message = try decoder.decode(ServerMessage.self, from: data)
            switch message {
            case .state(let count, let target):
                self.count = count
                self.target = target
            case .completed:
                self.count = self.target
            case .presence(let users):
                self.activeUsers = users
            }
        } catch {
            // Ignore malformed messages.
        }
    }

    private func handleOpen() {
        guard !hasOpened else { return }
        hasOpened = true
        connected = true
        retryAttempt = 0
        reconnectTask?.cancel()
        reconnectTask = nil

        let name = usernameProvider?().trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        send(.hello(name: name.isEmpty ? nil : name))
    }

    private func handleDisconnect() {
        connected = false
        activeUsers = []
        hasOpened = false
        webSocketTask = nil

        guard !manualClose, shareCode != nil, token != nil else { return }

        let delay = min(pow(2.0, Double(retryAttempt)) * 1.0, 20.0)
        retryAttempt += 1

        reconnectTask?.cancel()
        reconnectTask = Task { [weak self] in
            try? await Task.sleep(for: .seconds(delay))
            guard !Task.isCancelled else { return }
            self?.openConnection()
        }
    }

    private func send(_ message: ClientMessage) {
        guard let task = webSocketTask else { return }

        do {
            let data = try encoder.encode(message)
            let text = String(decoding: data, as: UTF8.self)
            task.send(.string(text)) { [weak self] error in
                guard let self else { return }
                if error != nil {
                    Task { @MainActor in
                        self.handleDisconnect()
                    }
                }
            }
        } catch {
            // Ignore encoding errors.
        }
    }
}
