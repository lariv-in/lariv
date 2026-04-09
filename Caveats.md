# Caveats When Working On This Codebase

- NEVER write go.mod or go.sum manually, use go mod init, go mod tidy -e, go work use for project management, If you can't run the required commands to manage go.mod and go.sum, then ask the user to run the commands.
- **Discover before you build:** Before designing a new component, getter, or interaction pattern, search and read what already exists—`components/`, `getters/`, `views/` (including `query_patcher*.go`, layer types, and helpers in `crud.go`), `registry/`, and plugins that solve a similar problem. Prefer reusing, composing, or lightly extending existing pieces over adding parallel types or one-off logic.
- In nearly all cases, take the address of components before inserting them into something that requires `PageInterface`. Otherwise, the value will not implement `MutableParentInterface` and its children will not be patchable.

- When you do add a component, it should implement at least `PageInterface` from `components/page.go`.

- If a component has children, it should implement `ParentInterface` from `components/parent.go` so its children can be traversed.
- If a component allows modifying its children, it should implement `MutableParentInterface` from `components/parent.go`.
- If a component is an input, it needs to implement `InputInterface` from `components/input.go` so that `FormComponent` can detect it and parse its fields.

- Whenever something requires a value that can depend on the request, it should use a `Getter` from `getters/getters.go` (shared context key constants live there too: `ContextKeyError`, `ContextKeyGet`, `ContextKeyIn`).
- The `getters/` package is organized by topic in sibling files (no subpackages), e.g. `key.go` for `Key`, `deref.go`, `format.go`, `any.go`, `association.go`, `association_list.go`, `join_association_list.go`, `association_ids.go`, `foreign_key.go`, `select.go` / `select_multi.go`, `navigate.go`, and helpers like `parse_int.go` / `parse_uint.go`. Browse those files or `grep` for `func ` when looking for an existing combinator.

- Before writing a custom getter, always confirm that no existing getter in `getters/` (and no small composition of existing getters) already covers the use case:
   - Use `getters.Deref(getters.Key[*T]("$in.Field"))` for nullable pointer fields instead of writing custom wrapper functions.
   - Use `getters.Format("format", getters.Any(getter1), ...)` to combine multiple getters into a formatted string instead of custom inline functions.
   - For route params like `id`, prefer `getters.Any(getters.Key[uint]("$id"))` instead of writing custom `uint -> string` wrapper getters.
   - For many-to-many filter state stored in `$get`, prefer `getters.AssociationIDs(getters.ContextKeyGet, "Field")` instead of manually unpacking `AssociationIDs`.

- When defining getter arguments, use the most restrictive type possible. `any` is almost always a bad idea.

- For foreign key selectors, the `InputForeignKey.Name`, the selector route/page it opens, and the `GetterSelect(...)` event name all need to match. If a `ParentID` input opens a selector table built for `DestinationID`, the selection event will be dispatched with the wrong name and the input will not update or close its modal.

- The same name-matching rule applies to `InputManyToMany` and `GetterMultiSelect(...)`. Many-to-many selectors also need to preserve `target_input` across the initial modal open and any filter/browse requests inside the modal. If `target_input` is dropped, the selector will dispatch the wrong field name and the chips will not update.

- For `InputForeignKey.Getter`, use `getters.Association[T](getters.Key[uint]("$in.FieldID"))`. It infers the table name from the type `T` via GORM's `db.Model()`.

- For `InputManyToMany.Getter`, prefer preloaded associations plus `getters.Key[[]T]("$in.Field")` instead of custom lookup getters. `InputManyToMany` re-renders from submitted `AssociationIDs`, but update/detail views should still preload the association so initial render and detail pages have the full related objects available.

- For **detail pages**, use `components.FieldManyToMany[T]` (`components/field_manytomany.go`) to show the same many-to-many association read-only. Reuse the same **`Getter`** and **`Display`** as the matching `InputManyToMany[T]` (including `getters.JoinAssociationList[...]` when the association goes through a join table). **`Link`** is optional: when set, it runs with `getters.ContextKeyIn` bound to each related row (same as **`Display`**), e.g. `lago.RoutePath("plugin.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})`. Prefer this over `FieldList` + per-row children for plain association lists; keep **`FieldList`** when each row is not a typed related model (e.g. ad-hoc `[]map[string]any` or heavily custom row UI).

- If the relation is intentionally not declared on the base GORM model and is instead represented by a separate join model, prefer shared getters such as `getters.JoinAssociationList[...]` / `getters.AssociationList[...]` plus shared query patchers instead of ad-hoc plugin-local lookup code.

- Models are not patchable through registries the way pages and views are. If a plugin needs to extend another plugin's data model, prefer a separate extension/join model owned by the new plugin plus page/view/query patches around it. Only add fields directly to the base GORM model when that relationship truly belongs in the base plugin and is intended to be a first-class part of that model.

- The filesystem selector routes have two different behaviors:
   - `filesystem.SelectRoute` and `filesystem.MoveSelectRoute` are directory pickers; directories are selectable.
   - `filesystem.MultiSelectRoute` is a file picker for asset-style many-to-many fields; files are selectable, but clicking a directory should browse into it instead of selecting it.

- time.Time should always be handled with consideration to the timezone, timezone is injected into the context in $tz key (*time.Location).

# Choice fields (string columns)

For a model field that is edited with a fixed set of string options (dropdown), keep **one** canonical map in the plugin’s `models.go` beside the GORM struct:

- **`map[string]string` where the key is the persisted value** (what the database stores and what `<option value>` submits) **and the value is the UI label** shown in selects and read-only views when you want a friendly string.
- **Do not add a parallel `const` block** whose only purpose is to name those keys again for the map or for `switch` statements. Use string literals as map keys (see `p_nirmancampus_students.StudentCategoryChoices`, `p_nirmancampus_programs.universityChoices`, `p_nirmancampus_programs.admissionSessionChoices`).
- **Forms:** `components.InputSelect[string]` with `Choices: getters.Static(registry.PairsFromMap(YourChoices))`. The current value getter should resolve with `registry.PairFromMap(s, YourChoices)` when `s` is non-empty, and fall back to `registry.Pair[string, string]{Key: s, Value: s}` for unknown legacy rows so the UI still renders.
- **Detail / list labels:** Prefer a map lookup (`YourChoices[s]`) or `PairFromMap` instead of duplicating the same options in a `switch`.
- **Generators, form patchers, and tests** must use the **same string literals** as the map keys so inserts and validation stay aligned.
- `registry.PairsFromMap` sorts options by key. If the UI must follow a custom order, build `[]registry.Pair[string, string]` explicitly and document why.

# Environment selector

- `components.Environment[T]` (`components/environment.go`) renders a `<select>` that reads and writes the `environment` JSON cookie; the parsed map is available as `$environment` (`map[string]string`) on the request context.
- `Options` should be a `Getter` returning `[]registry.Pair[T, string]` from `registry/registry.go`: **Key** is the option id (stored in the cookie and sent as `<option value>`), **Value** is the display label. Keys are stringified with `fmt.Sprint` for HTML and for comparing to the cookie string.
- `Default` is optional. When the cookie has no entry for `Key`, it runs; a **zero** `T` means “no default” (do not treat as a selected id).
- Register `Environment` as **`&components.Environment[T]{...}`** inside `[]PageInterface` (and in registries that expect pointers) so the node implements `MutableParentInterface` and remains patchable—the same pointer rule as other `PageInterface` children.
- List views, query patchers, and any other code that scopes data by the user’s choice must read the same cookie value and parse it the same way (e.g. numeric id string for `Environment[uint]`).

# SQL identifiers (PostgreSQL)

- Column names **`start`** and **`end`** must be quoted in GORM `Order` clauses and other raw SQL fragments, e.g. `` Order(`"start" ASC`) `` or `Where(`"start" <= ? AND "end" >= ?`, ...)`, because unquoted `start` is not a valid column reference in PostgreSQL.

# Registry

Anything that needs to be patchable on an app-wide scale should be done via a registry from `registry/registry.go`.

Registries use the `Register` method to add to the registry and the `Patch` method to register a patch for an existing element in the registry.

Existing registries:
   - `lago/registry_commands.go` for adding custom commands
   - `lago/registry_config.go` for adding config fields to `totschool.toml`
   - `lago/registry_generators.go` for adding generators that run when the `generate` command is run
   - `lago/registry_layers.go` for adding global layers, generally not needed
   - `lago/registry_pages.go` for adding pages; always insert a pointer to a `PageInterface` implementer
   - `lago/registry_plugins.go` for adding plugin information, primarily used by `p_dashboard/components/apps_grid.go`
   - `lago/registry_routes.go` for adding HTTP routes
   - `lago/registry_views.go` for adding views (see the views section below)
   - `lago/regsitry_dbinit.go` for adding functions that run after the database is initialized; run model automigrations here

# Plugin page source files (layout convention)

For plugins with many registered pages, prefer **several files in the plugin root** (same package as the rest of the plugin), not a separate `pages/` or `models/` subpackage, unless there is a strong reason to split packages.

- **`pages.go`** — `init()` that calls `registerMenuPages`, `registerFilterPages`, etc., plus **small** shared wiring (e.g. sidebar menus). Keep this file short when possible.
- **`pages_<area>.go`** — larger page trees: name by concern, e.g. `pages_form.go` (create/update and shared form fields), `pages_detail.go` (detail view and closely related UI such as a delete confirmation modal), `pages_table.go` (list filters, main list table, and selection modal when they belong together), `pages_structure.go` (a sizeable secondary feature).
- **Rule of thumb:** treat **detail**, **forms**, and **list/filter/table** surfaces as the primary splits when those definitions are large; merge **tiny** registrations into `pages.go` or into the most related larger file so the layout stays maintainable.
- **Size:** aim for roughly **200–400 lines per file** where practical. Going slightly over is fine for cohesion; avoid many separate files that are only a few dozen lines each unless the boundary is conceptually important.

See `plugins/p_nirmancampus_programs` for a concrete layout using this pattern.

# HTTP routes nested under another app's prefix

When a plugin mounts HTTP routes under another plugin's `AppUrl` (e.g. Students at `/students/`), use an **`addon/<slug>/`** segment after that base so your subtree does not structurally overlap the host's dynamic routes.

- **Convention:** `HostAppUrl + "addon/" + <short-slug> + "/"` as the route prefix for the nested feature (e.g. `p_nirmancampus_students.AppUrl + "addon/academicrecords/"` → `/students/addon/academicrecords/`).
- **Why:** `http.ServeMux` rejects patterns that both match some paths. A host app often has `/prefix/{id}/`, `/prefix/{id}/edit/`, `/prefix/{id}/delete/`, etc. A naive path like `/prefix/academicrecords/...` can match those patterns with `id` interpreted as `academicrecords`, or collide with `/prefix/{id}/delete/` when a segment is `delete`. The fixed `addon/` segment keeps the nested feature's paths longer and unambiguous relative to `{id}` routes.
- Prefer **`PluginTypeAddon`** in `registry_plugins.go` when the nested feature is not its own dashboard app tile; link into it from the host app's menu or UI.

# Views

A view is the primary HTTP handler for a route. A `*views.View` (`views/views.go`) is only:

- which `PageInterface` to render (`PageName` + `PageLookup`)
- an **ordered** list of per-route layers (`Layers`), each implementing `views.Layer` (`Next(View, http.Handler) http.Handler`)

Global HTTP concerns (DB, `$request`, etc.) live in `views.GlobalLayer` and app registration, not inside the view struct. Build routes from `lago.GetPageView("plugin.PageName")`, then chain `WithLayer("stable.key", layer)`.


**HTMX redirects:** use `views.HtmxRedirect(w, r, url, code)` (`views/htmx_redirect.go`) anywhere you would call `http.Redirect` for user-visible navigation. When `HX-Request` is `true`, it sets the `HX-Redirect` header and responds with `200`; otherwise it calls `http.Redirect` with the same `code` you pass (e.g. `http.StatusSeeOther` after POST, `http.StatusMovedPermanently` where appropriate). Do not call `http.Redirect` directly in app or plugin handlers unless you have a rare reason to bypass HTMX. `lago.RedirectView` registers a view whose `RedirectLayer` resolves a URL getter and ends with `views.HtmxRedirect(..., http.StatusMovedPermanently)`.

**Typed CRUD layer** (each owns one concern; order matters):

- `views.LayerList[T]` — paginated list query from URL params; puts `components.ObjectList[T]` in context under `Key`; merges filter/query state into `$get` (and coerces types from the page’s first form when present). On failure it sets `_global` in `getters.ContextKeyError` and calls `next` (no direct error HTTP response).
- `views.LayerDetail[T]` — load one row by path param PK; place **before** update/delete layer that needs the same record. Same error pattern as list on failure.
- `views.LayerCreate[T]` — POST create; sets `$id` on success.
- `views.LayerUpdate[T]` — POST update; expects the record already in context (usually after `LayerDetail`).
- `views.LayerDelete[T]` — POST delete (not HTTP `DELETE`; matches confirmation forms).
- `views.LayerSingleton[T]` — singleton settings load/create on GET/POST.
- `views.LayerJsonImport[T]` — JSON file import.
- `views.MethodLayer` — custom handler for a specific HTTP method.

**Query patching:** attach `views.QueryPatchers[T]` (named `registry.Pair`s) on `LayerList`, `LayerDetail`, or `LayerUpdate`. Prefer the built-in patchers in `views/query_patcher_*.go`: `QueryPatcherPreload[T]`, `QueryPatcherOrderBy[T]`, `QueryPatcherJoinFilter[T, TJoin]` (reads filter values from `$get`). Do not duplicate ad-hoc query logic when these suffice.

**Form patching:** attach `views.FormPatchers` on `LayerCreate` and `LayerUpdate` (`views/form_patchers.go`). `InputManyToMany.Parse` still yields `AssociationIDs`; create/update/singleton layer persists many-to-many via GORM after the row save—do not model those inputs as plain scalar columns.

**Patching views across plugins:** give every layer a stable string key (e.g. `"students.detail"`). Other packages should use `InsertLayerBefore`, `InsertLayerAfter`, or `PatchLayer` against those keys—not fragile positional assumptions.

**Extra context on another plugin’s page** (e.g. related `ObjectList` on a base detail view): do **not** hide DB access inside a component getter. Implement a small type that satisfies `views.Layer`, load data in `Next`, `context.WithValue` the result, and register or patch it onto the base view **after** the layer that provides the parent record (e.g. `InsertLayerAfter("base.detail", "myplugin.extra", myLayer{})`). Point `DataTable` (or similar) at that context key with `getters.Key[...]`.

**Custom forms and errors:** use `view.ParseForm` and `views.ContextWithErrorsAndValues` to re-render with field/global errors; do not recreate removed helpers like `HasErrors` / `RenderWithErrors`.

# Error handling

Error handling is very important. If an error occurs that could make the program behave incorrectly, it is preferable to panic rather than keep it running in a bad state.

Whenever a recoverable error occurs, it should be logged, no matter how unlikely.

All edge cases need to be logged, no edge case should ever be ignored.

Use `log/slog` for recoverable errors and `log.Panicf()` for non-recoverable errors.

# Component Patching

Use `components.InsertChildBefore`, `components.InsertChildAfter`, and `components.ReplaceChild`.

- If a page or form is intended to be extended by another plugin, add stable `Page.Key` values in the base plugin first, then patch against those keys. Do not rely on brittle structural matching when a reusable extension point can be made explicit.

- If a selector route/page is generally useful beyond one addon, add it to the base plugin instead of creating a one-off copy in the addon plugin.

# Generators

- Plugins register data generators via `lago.RegistryGenerator.Register("name", lago.Generator{Create: ..., Remove: ...})` inside their `init()` func.
- Execution order is strictly determined by the `GeneratorOrder` array defined in the deployment's TOML config (e.g., `nirmancampus.toml`).
- `RunGenerators` executes in two phases to respect foreign-key constraints:
  - **Phase 1 (Remove):** Iterates *backwards* through the TOML list, deleting dependent tables before base tables.
  - **Phase 2 (Create):** Iterates *forwards* through the TOML list, creating base tables before dependent ones.
- **Many-to-Many Cleanup:** When writing a generator's `Remove` function, you must manually issue raw SQL to clear any many-to-many join tables (e.g., `db.Exec("DELETE FROM student_assets")`) before deleting the primary model, because GORM/PostgreSQL will not automatically cascade delete rows from many-to-many join tables, resulting in FK violation errors.
