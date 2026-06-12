import Foundation

enum DeepLinkParser {
    static func route(from url: URL) -> AppRoute? {
        if url.scheme?.lowercased() == AppConfig.urlScheme {
            return routeFromCustomScheme(url)
        }

        guard let host = url.host?.lowercased() else { return nil }
        let allowedHosts = ["zikirhatmi.abapnews.tr", "www.zikirhatmi.abapnews.tr"]
        guard allowedHosts.contains(host) else { return nil }

        return route(fromPathComponents: url.pathComponents.filter { $0 != "/" })
    }

    private static func routeFromCustomScheme(_ url: URL) -> AppRoute? {
        if url.host == "h" {
            let shareCode = url.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
            guard !shareCode.isEmpty else { return nil }
            if shareCode.hasSuffix("/join") {
                let code = shareCode.replacingOccurrences(of: "/join", with: "")
                return code.isEmpty ? nil : .join(shareCode: code)
            }
            return .hatim(shareCode: shareCode)
        }

        return route(fromPathComponents: url.pathComponents.filter { $0 != "/" })
    }

    private static func route(fromPathComponents components: [String]) -> AppRoute? {
        guard components.count >= 2, components[0] == "h" else { return nil }
        let shareCode = components[1]
        guard !shareCode.isEmpty else { return nil }

        if components.count >= 3, components[2] == "join" {
            return .join(shareCode: shareCode)
        }

        return .hatim(shareCode: shareCode)
    }
}
