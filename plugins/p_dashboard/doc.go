// Package p_dashboard implements the central launchpad, top bar navigation buttons, and theme toggling for the Lago application portal.
//
// # Registrations and Features Added
//
// # Pages
//
//   - "dashboard.HomeRedirectStub" -> components.ContainerColumn
//     Redirection layout stub targeting home page landing.
//   - "dashboard.AppsPage" -> components.ShellTopbarScaffold
//     The main central launchpad screen containing the launch buttons for all active plugins.
//
// # Top Bar Navigation Elements
//
// Registers UI widgets into components.RegistryTopbar:
//
//   - "dashboard.appsPageButton" -> components.ButtonLink
//     Icon link directing users back to the dashboard apps grid.
//   - "dashboard.themeButton" -> pcomps.ThemeButton
//     Theme toggle control.
//   - "dashboard.userDropdown" -> pcomps.UserDropdown
//     User profile, credentials operations, and logout options dropdown widget.
//
// # Routes
//
//   - "dashboard.AppsPage" -> lago.Route
//     Maps the "/dashboard/" path to resolve dashboard view layout actions.
//
// # Views
//
//   - "dashboard.AppsView" -> views.View
//     Compiles and renders the central apps grid page under security validation.
//
// # Patches Applied
//
//   - "p_users.LoginSuccessView": Patches views to route successful logins directly to the dashboard apps grid.
//   - "core.HomeView": Patches home page redirects to guide authenticated sessions to "/dashboard/" and guest sessions to "/users/login/".
package p_dashboard
