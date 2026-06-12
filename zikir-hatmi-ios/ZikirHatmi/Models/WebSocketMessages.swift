import Foundation

enum ServerMessage: Decodable {
    case state(count: Int, target: Int)
    case completed
    case presence(users: [PresenceUser])

    private enum CodingKeys: String, CodingKey {
        case type, count, target, users
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let type = try container.decode(String.self, forKey: .type)

        switch type {
        case "state":
            self = .state(
                count: try container.decode(Int.self, forKey: .count),
                target: try container.decode(Int.self, forKey: .target)
            )
        case "completed":
            self = .completed
        case "presence":
            self = .presence(users: try container.decodeIfPresent([PresenceUser].self, forKey: .users) ?? [])
        default:
            throw DecodingError.dataCorruptedError(forKey: .type, in: container, debugDescription: "Unknown message type")
        }
    }
}

enum ClientMessage: Encodable {
    case increment(amount: Int?)
    case hello(name: String?)
    case setName(name: String?)

    private enum CodingKeys: String, CodingKey {
        case type, amount, name
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        switch self {
        case .increment(let amount):
            try container.encode("increment", forKey: .type)
            if let amount {
                try container.encode(amount, forKey: .amount)
            }
        case .hello(let name):
            try container.encode("hello", forKey: .type)
            if let name {
                try container.encode(name, forKey: .name)
            }
        case .setName(let name):
            try container.encode("set_name", forKey: .type)
            if let name {
                try container.encode(name, forKey: .name)
            }
        }
    }
}
