// Package p_website provides website routing configurations loaded from the database.
//
// # Dynamic Views Security Note
//
// Note: We cannot allow a dynamic view that will serve any arbitrary file under a directory,
// since it might give arbitrary read access using Go templates and the custom filesystem.
//
// # Registrations and Features Added
//
// # Database Models
//
//   - p_website.DBRoute: DB model mapping active URL paths to page VNodes in p_filesystem.
//
// # Pages
//
//   - p_website.DynamicWebsitePage: Renders database-driven website pages or streams static files from p_filesystem.
//   - p_website.RoutesListPage & p_website.RoutesDetailPage: Shell pages for listing and viewing route configurations.
//   - p_website.RoutesCreatePage & p_website.RoutesUpdatePage & p_website.RoutesDeleteForm: Forms and dialogs for managing route records.
//
// # Views
//
//   - "p_website.DynamicWebsiteView": Resolves requested URL paths against active database routes and delegates to DynamicWebsitePage.
//   - "p_website.RoutesListView" & "p_website.RoutesDetailView": Admin management views for route entities.
//   - "p_website.RoutesCreateView" & "p_website.RoutesUpdateView" & "p_website.RoutesDeleteView": CRUD handlers for route records.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/{path...}" (Patches "core.HomeRoute"): Dynamic catch-all route mapped to p_website.DynamicWebsiteView.
//   - "/website/": List view of configured database routes.
//   - "/website/create/": Route creation form.
//   - "/website/{id}/": Detail view of a database route.
//   - "/website/{id}/edit/": Route edit form.
//   - "/website/{id}/delete/": Route deletion form.
package p_website
