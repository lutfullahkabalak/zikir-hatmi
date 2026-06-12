import Foundation
import Security

enum TokenStore {
    private static let service = "com.zikirhatmi.app.tokens"

    static func tokenKey(for shareCode: String) -> String {
        "hatim-token:\(shareCode)"
    }

    static func save(token: String, for shareCode: String) throws {
        let account = tokenKey(for: shareCode)
        let data = Data(token.utf8)

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]

        SecItemDelete(query as CFDictionary)

        var addQuery = query
        addQuery[kSecValueData as String] = data
        addQuery[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlock

        let status = SecItemAdd(addQuery as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw TokenStoreError.saveFailed(status)
        }
    }

    static func load(for shareCode: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: tokenKey(for: shareCode),
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else {
            return nil
        }
        return String(data: data, encoding: .utf8)
    }

    static func delete(for shareCode: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: tokenKey(for: shareCode),
        ]
        SecItemDelete(query as CFDictionary)
    }
}

enum TokenStoreError: LocalizedError {
    case saveFailed(OSStatus)

    var errorDescription: String? {
        switch self {
        case .saveFailed(let status):
            return "Token kaydedilemedi (OSStatus \(status))."
        }
    }
}
