// Package p_pwa implements Progressive Web App (PWA) manifest compilation, service worker mapping, offline fallback pages, and Android asset links hosting.
// It injects webmanifest links into the global HTML shell head and serves static resource dirs.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_pwa" -> p_pwa.PwaConfig
//         Configures manifest parameters (AppName, AppThemeColor, AppIcons, AppShortcuts) and local assets directories (StaticDir, ServiceWorkerPath, OfflineViewName).
//
// # Shell Head Snippets
//
// Registers links in components.RegistryShellHeadNodes in init():
//
//   - "pwa.manifestLink" -> Link
//         Injects <link rel="manifest" href="/app.webmanifest"> tag for browser PWA discovery.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/app.webmanifest" -> pwa.ManifestView
//   - "/serviceworker.js" -> pwa.ServiceWorkerView
//   - "/offline" -> pwa.OfflineView
//   - "/static/pwa/" -> pwa.StaticPwaView
//   - "/.well-known/assetlinks.json" -> pwa.assetLinksView
//
// # Views
//
//   - "pwa.ManifestView": Serves the app.webmanifest JSON formatted directly from the configuration keys.
//   - "pwa.ServiceWorkerView": Serves custom service worker JS scripts or a default caching/offline handler fallback script.
//   - "pwa.OfflineView": Serves default or configured offline pages when network disconnects.
//   - "pwa.StaticPwaView": Serves custom PWA static assets from StaticDir.
//   - "pwa.assetLinksView": Serves Android Digital Asset Links JSON configuration maps.
//
// # Patches Applied
//
//   - "core.Title": Patches the default title tag value in GOMPONENTS shell structures to match the PwaConfig.AppName configuration.
package p_pwa
