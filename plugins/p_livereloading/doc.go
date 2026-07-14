// Package p_livereloading implements automated browser refresh / live reloading capabilities during local development.
// It injects client-side javascript monitors into the global HTML shell head and maps WebSocket sync servers.
//
// # Registrations and Features Added
//
// # Shell Head Snippets
//
// Registers UI script widgets into components.RegistryShellHeadNodes in init():
//
//   - "liverealoading.js" -> Script
//     Web browser script executing reconnect client loops. When the backend service reboots (e.g. on code rebuilds), the WebSocket disconnects and the script polls until reconnection is established, triggering a page reload.
//
// # Routes
//
//   - "livereloading.ws" -> lago.Route
//     Maps the "/_livereload" endpoint path to execute WebSocket handler handshakes.
package p_livereloading
