import Foundation
import Observation

@Observable
final class UserPreferences {
    static let shared = UserPreferences()

    private enum Keys {
        static let username = "zikir-username"
        static let hapticsEnabled = "haptics-enabled"
    }

    var username: String {
        didSet {
            let trimmed = username.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                defaults.removeObject(forKey: Keys.username)
            } else {
                defaults.set(trimmed, forKey: Keys.username)
            }
        }
    }

    var hapticsEnabled: Bool {
        didSet {
            defaults.set(hapticsEnabled, forKey: Keys.hapticsEnabled)
        }
    }

    private let defaults = UserDefaults.standard

    private init() {
        username = defaults.string(forKey: Keys.username) ?? ""
        if defaults.object(forKey: Keys.hapticsEnabled) == nil {
            hapticsEnabled = true
        } else {
            hapticsEnabled = defaults.bool(forKey: Keys.hapticsEnabled)
        }
    }

    func setUsername(_ value: String) {
        username = value.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    func clearUsername() {
        username = ""
    }
}
