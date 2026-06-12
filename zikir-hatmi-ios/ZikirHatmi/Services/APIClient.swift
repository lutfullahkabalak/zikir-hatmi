import Foundation

enum APIError: LocalizedError {
    case invalidResponse
    case httpError(status: Int)
    case decodingFailed

    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "Geçersiz sunucu yanıtı."
        case .httpError(let status):
            return "Sunucu hatası (\(status))."
        case .decodingFailed:
            return "Yanıt çözümlenemedi."
        }
    }
}

final class APIClient: Sendable {
    static let shared = APIClient()

    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(session: URLSession = .shared) {
        self.session = session
        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
    }

    func getHatim(shareCode: String) async throws -> HatimState {
        let url = AppConfig.apiURL(path: "/hatims/\(shareCode)")
        let (data, response) = try await session.data(from: url)
        try validate(response: response)
        do {
            return try decoder.decode(HatimState.self, from: data)
        } catch {
            throw APIError.decodingFailed
        }
    }

    func createHatim(title: String, target: Int, password: String?) async throws -> CreateHatimResponse {
        let url = AppConfig.apiURL(path: "/hatims")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let trimmedPassword = password?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        let body = CreateHatimRequest(
            title: title.trimmingCharacters(in: .whitespacesAndNewlines),
            target: target,
            password: trimmedPassword.isEmpty ? nil : trimmedPassword
        )
        request.httpBody = try encoder.encode(body)

        let (data, response) = try await session.data(for: request)
        try validate(response: response)
        do {
            return try decoder.decode(CreateHatimResponse.self, from: data)
        } catch {
            throw APIError.decodingFailed
        }
    }

    func joinHatim(shareCode: String, password: String?) async throws -> JoinResponse {
        let url = AppConfig.apiURL(path: "/hatims/\(shareCode)/join")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let trimmedPassword = password?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        let body = JoinRequest(password: trimmedPassword.isEmpty ? nil : trimmedPassword)
        request.httpBody = try encoder.encode(body)

        let (data, response) = try await session.data(for: request)

        guard let http = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        if http.statusCode == 401 {
            throw JoinError.invalidPassword
        }
        if http.statusCode == 404 {
            throw JoinError.notFound
        }
        guard (200 ... 299).contains(http.statusCode) else {
            throw APIError.httpError(status: http.statusCode)
        }

        do {
            return try decoder.decode(JoinResponse.self, from: data)
        } catch {
            throw APIError.decodingFailed
        }
    }

    private func validate(response: URLResponse) throws {
        guard let http = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }
        guard (200 ... 299).contains(http.statusCode) else {
            throw APIError.httpError(status: http.statusCode)
        }
    }
}

enum JoinError: LocalizedError {
    case invalidPassword
    case notFound
    case generic

    var errorDescription: String? {
        switch self {
        case .invalidPassword:
            return "Şifre hatalı. Lütfen tekrar deneyin."
        case .notFound:
            return "Hatim bulunamadı."
        case .generic:
            return "Katılım sırasında hata oluştu."
        }
    }
}
