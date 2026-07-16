// Package lariv is the application kernel: configuration loading, HTTP server wiring,
// and aggregated plugin registries (pages, views, routes, models, migrations, etc.).
//
// # Core Plugin Registrations
//
// The kernel registers a default "core" plugin during boot via CorePlugin().
// Below is a detailed list of features, pages, views, and routes contributed by the core plugin:
//
// # Core Global Layers (Global HTTP Middleware)
//
// Registered under the following keys:
//
//   - "core.AttachRequestLayer" -> views.AttachRequestLayer
//     Attaches raw query parameter map ($get), request object ($request), and start timestamp ($timestamp) to the request context.
//   - "core.DbLayer" -> DBLayer
//     Attaches the active GORM database connection pool instance ($db) to the request context.
//   - "core.LoggingLayer" -> LoggingLayer (Debug mode only)
//     Logs HTTP request methods, paths, durations, and response codes.
//   - "core.CacheDisableLayer" -> CacheDisableLayer (Debug mode only)
//     Injects HTTP headers (Cache-Control, Pragma, Expires) to disable browser caching under local development.
//   - "core.HtmxBoostLayer" -> HtmxBoostLayer
//     Configures response headers for HTMX requests.
//   - "core.EnvironmentLayer" -> EnvironmentLayer
//     Sets up standard environment config context.
//
// # Core Views
//
// Registered under:
//
//   - "core.HomeView" -> views.View
//     Resolves and displays the "core.HomePage" page component.
//
// # Core Pages
//
// Registered under:
//
//   - "core.HomePage" -> components.ShellBase
//     The default home page visual layout scaffolding.
//
// # Core Routes
//
// Registered under:
//
//   - "core.HomeRoute" -> lariv.Route
//     Maps the root path ("/") to render standard text content or the home page view.
//
// # Bundled Plugins
//
// The framework includes a collection of bundled plugins supporting standard app capabilities:
//
//   - Users & Auth: See [github.com/lariv-in/lariv/plugins/p_users] for user management and role controls.
//   - Dashboard: See [github.com/lariv-in/lariv/plugins/p_dashboard] for the centralized launchpad portal.
//   - XLSX Export: See [github.com/lariv-in/lariv/plugins/p_export] for spreadsheet export features.
//   - Filesystem: See [github.com/lariv-in/lariv/plugins/p_filesystem] for local/GCS virtual filesystem drives.
//   - Google GenAI: See [github.com/lariv-in/lariv/plugins/p_google_genai] for Gemini AI client loaders.
//   - Live Reload: See [github.com/lariv-in/lariv/plugins/p_livereloading] for hot browser refreshes.
//   - LLM Assistant: See [github.com/lariv-in/lariv/plugins/p_llm_assistant] for interactive AI chat prompts.
//   - OTP Recovery: See [github.com/lariv-in/lariv/plugins/p_otp] for SMS/email one-time password recovery.
//   - PWA Support: See [github.com/lariv-in/lariv/plugins/p_pwa] for Progressive Web App capabilities.
package lariv
