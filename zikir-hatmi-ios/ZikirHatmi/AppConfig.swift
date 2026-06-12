import Foundation

enum AppConfig {
    static let apiBaseURL = URL(string: "https://zikirhatmi.abapnews.info")!
    static let shareBaseURL = URL(string: "https://zikirhatmi.abapnews.info")!
    static let urlScheme = "zikirhatmi"

    static func apiURL(path: String) -> URL {
        let normalized = path.hasPrefix("/") ? String(path.dropFirst()) : path
        return apiBaseURL.appending(path: normalized)
    }

    static func wsURL(shareCode: String, token: String) -> URL {
        var components = URLComponents()
        components.scheme = "wss"
        components.host = apiBaseURL.host
        components.path = "/ws/\(shareCode)"
        components.queryItems = [URLQueryItem(name: "token", value: token)]
        return components.url!
    }

    static func shareURL(shareCode: String) -> URL {
        shareBaseURL.appending(path: "h/\(shareCode)")
    }
}
