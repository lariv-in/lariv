// Package p_website provides website routing configurations loaded from the database.
//
// # Dynamic Views Security Note
//
// Note: We cannot allow a dynamic view that will serve any arbitrary file under a directory,
// since it might give arbitrary read access using Go templates and the custom filesystem.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_website" -> p_website.WebsiteConfig (TOML section [plugins.p_website])
//     newPageRootDir: VNode path for new blank HTML pages on route create (parents created as needed).
//     assetsDir: VNode path for GrapesJS image uploads (defaults to "{newPageRootDir}/assets").
//
// # Database Models
//
//   - p_website.DBRoute: DB model mapping active URL paths to page VNodes in p_filesystem.
//     Optional GrapesProject JSON stores GrapesJS editor state for re-editing.
//     Theme stores a GrapesJSThemes registry key for builder/published styling.
//
// # Pages
//
//   - p_website.DynamicWebsitePage: Renders database-driven website pages or streams static files from p_filesystem.
//   - p_website.RoutesListPage & p_website.RoutesDetailPage: Shell pages for listing and viewing route configurations.
//   - p_website.RoutesCreatePage & p_website.RoutesUpdatePage & p_website.RoutesDeleteForm: Forms and dialogs for managing route records.
//   - p_website.RoutesBuilderPage: Full-page GrapesJS editor for HTML-like page files.
//
// # Views
//
//   - "p_website.DynamicWebsiteView": Resolves requested URL paths against active database routes and delegates to DynamicWebsitePage.
//   - "p_website.RoutesListView" & "p_website.RoutesDetailView": Admin management views for route entities.
//   - "p_website.RoutesCreateView" & "p_website.RoutesUpdateView" & "p_website.RoutesDeleteView": CRUD handlers for route records.
//   - "p_website.RoutesBuilderView": Authenticated GrapesJS builder UI.
//   - "p_website.RoutesBuilderProjectView": GET/POST remote storage for GrapesJS project JSON and HTML export.
//   - "p_website.BuilderAssetUploadView": Authenticated multipart upload for GrapesJS assets into p_filesystem.
//   - "p_website.PublicAssetView": Public GET that streams uploaded asset VNodes for <img src> / CSS.
//
// # GrapesJS Blocks
//
//   - Registers Layout blocks (Section, 2 Columns, 3 Columns, Card) plus catalog
//     blocks for Accordion, Blurb, Button, CTA, Code, Divider, Dropdown, Gallery,
//     Heading, Hero, Icon, Icon list, Image, Link, DotLottie, Person, Pricing tables,
//     Slider, Tabs, Testimonial, Toggleable, Text, and counters via Plugin.GrapesJSBlocks.
//     Layout HTML is loaded from grapesjs_blocks/; catalog HTML from grapesjs_components/.
//
// # GrapesJS Components
//
//   - Registers DomComponents types (p_website.*) via Plugin.GrapesJSComponents for the
//     builder catalog above. Interactive types ship trusted runtime scripts; DotLottie
//     loads @lottiefiles/dotlottie-wc from a pinned CDN only when used (canvas onRender
//     and published HTML inject).
//
// # GrapesJS Traits
//
//   - Registers custom trait type p_website.src-url via Plugin.GrapesJSTraits for media
//     source URL fields (Image, DotLottie, and similar).
//
// # GrapesJS Themes
//
//   - Registers CSS themes via Plugin.GrapesJSThemes (default: p_website.default from
//     grapesjs_theme.css). DBRoute.Theme stores the selected registry key.
//     Theme is selectable on route create/edit forms and the GrapesJS builder toolbar;
//     CSS is applied in the builder canvas and injected into published / public HTML.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/{path...}" (Patches "core.HomeRoute"): Dynamic catch-all route mapped to p_website.DynamicWebsiteView.
//   - "/website/": List view of configured database routes.
//   - "/website/create/": Route creation form (select existing page or create a blank HTML file).
//   - "/website/{id}/": Detail view of a database route.
//   - "/website/{id}/edit/": Route edit form.
//   - "/website/{id}/delete/": Route deletion form.
//   - "/website/{id}/builder/": GrapesJS page builder (shown for .html/.htm/.tmpl pages).
//   - "/website/{id}/builder/project/": GrapesJS remote load/store API.
//   - "/website/{id}/builder/theme/": Persist selected GrapesJS theme for the route.
//   - "/website/builder/assets/": GrapesJS AssetManager upload (stores VNodes under assetsDir).
//   - "/media/{id}/": Public stream of an uploaded asset VNode.
package p_website
