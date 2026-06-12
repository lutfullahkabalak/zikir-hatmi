import Foundation

struct HatimState: Codable, Sendable {
    let shareCode: String
    let title: String
    let count: Int
    let target: Int
    let requiresPassword: Bool
}

struct CreateHatimRequest: Codable, Sendable {
    let title: String
    let target: Int
    let password: String?
}

struct CreateHatimResponse: Codable, Sendable {
    let shareCode: String
    let title: String
    let count: Int
    let target: Int
    let requiresPassword: Bool
    let token: String
}

struct JoinRequest: Codable, Sendable {
    let password: String?
}

struct JoinResponse: Codable, Sendable {
    let token: String
}

struct PresenceUser: Codable, Sendable, Identifiable, Hashable {
    let id: String
    let name: String
}

enum AppRoute: Hashable {
    case home
    case hatim(shareCode: String, showCreatedBanner: Bool = false)
    case join(shareCode: String)
}
