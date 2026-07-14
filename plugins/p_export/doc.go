// Package p_export implements Excel spreadsheet (XLSX) creation and table export features for GORM models.
// It maps database metadata catalogs, tracks model relationship graphs, and downloads filtered table records.
//
// # Registrations and Features Added
//
// # Pages
//
//   - "export.Menu" -> components.SidebarMenu
//     Sidebar menu layout detailing XLSX export selections.
//   - "export.Page" -> components.ShellScaffold
//     The main export panel structure wrapping the database table selector widget (exportPickerPage).
//
// # Routes
//
//   - "export.PageRoute" -> lago.Route
//     Maps "/export/" to resolve catalog loading and selection rendering.
//   - "export.DownloadRoute" -> lago.Route
//     Maps "/export/download/" to process dynamic download actions.
//
// # Views
//
//   - "export.PageView" -> views.View
//     Renders the table selection screen, utilizing catalogLayer to fetch and inject GORM schema catalog maps.
//   - "export.DownloadView" -> views.View
//     Processes POST download actions, writing spreadsheet content (sheets, columns, relationship grids) using xlsx.go helper frameworks.
package p_export
