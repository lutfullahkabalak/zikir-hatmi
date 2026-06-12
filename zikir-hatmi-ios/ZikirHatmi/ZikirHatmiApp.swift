import SwiftUI

@main
struct ZikirHatmiApp: App {
    @State private var router = AppRouter()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(router)
                .onOpenURL { url in
                    router.handle(url: url)
                }
                .onContinueUserActivity(NSUserActivityTypeBrowsingWeb) { activity in
                    if let url = activity.webpageURL {
                        router.handle(url: url)
                    }
                }
        }
    }
}

@Observable
@MainActor
final class AppRouter {
    var path = NavigationPath()
    var pendingDeepLink: AppRoute?

    func handle(url: URL) {
        guard let route = DeepLinkParser.route(from: url) else { return }
        navigate(to: route)
    }

    func navigate(to route: AppRoute) {
        switch route {
        case .home:
            path = NavigationPath()
        case .hatim(let shareCode, let showCreatedBanner):
            path = NavigationPath()
            path.append(AppRoute.hatim(shareCode: shareCode, showCreatedBanner: showCreatedBanner))
        case .join(let shareCode):
            path = NavigationPath()
            path.append(AppRoute.hatim(shareCode: shareCode))
            path.append(AppRoute.join(shareCode: shareCode))
        }
    }

    func openHatim(_ shareCode: String, showCreatedBanner: Bool = false) {
        navigate(to: .hatim(shareCode: shareCode, showCreatedBanner: showCreatedBanner))
    }

    func openJoin(for shareCode: String) {
        path.append(AppRoute.join(shareCode: shareCode))
    }
}

struct RootView: View {
    @Environment(AppRouter.self) private var router
    @State private var createOpen = false

    var body: some View {
        @Bindable var router = router

        NavigationStack(path: $router.path) {
            HomeView {
                createOpen = true
            }
            .navigationDestination(for: AppRoute.self) { route in
                switch route {
                case .home:
                    HomeView { createOpen = true }
                case .hatim(let shareCode, let showCreatedBanner):
                    HatimView(
                        shareCode: shareCode,
                        showCreatedBanner: showCreatedBanner,
                        onNavigateToJoin: {
                            router.openJoin(for: shareCode)
                        },
                        onNavigateToHatim: { code, showBanner in
                            router.openHatim(code, showCreatedBanner: showBanner)
                        }
                    )
                case .join(let shareCode):
                    JoinView(shareCode: shareCode) { _ in
                        router.navigate(to: .hatim(shareCode: shareCode))
                    }
                }
            }
        }
        .background(RamadanBackground())
        .sheet(isPresented: $createOpen) {
            CreateHatimSheet(isPresented: $createOpen) { response in
                router.openHatim(response.shareCode, showCreatedBanner: true)
            }
        }
    }
}
