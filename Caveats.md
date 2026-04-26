# Caveats — This Codebase

## Start Here

Before adding code, search for existing patterns and extension points:

- New component / getter / pattern → search `components/`, `getters/`, `views/` (`query_patcher*.go`, layers, `crud.go` helpers), `registry/`, plugin `actions*.go`, and similar plugins.
- Reuse, compose, or lightly extend. Avoid parallel types and one-off logic unless existing helpers cannot cover the behavior.
- Models are not patchable via registries like pages/views. Extend another plugin's data with an extension/join model in your plugin plus page/view/query patches. Add fields to base GORM models only when relationship truly belongs in base plugin as first-class.

## Go Hygiene

- Never hand-edit `go.mod` / `go.sum`. Use `go mod init`, `go mod tidy -e`, `go work use`. Can't run those → ask user.
- No repo-local Go caches: `.cache-go`, `.gocache`, `.gomodcache`, custom `GOMODCACHE` / `GOCACHE` / `GOPATH` inside workspace. Use system caches. Sandbox blocks that → ask user; don't pollute repo.
- No `recover()` in background goroutine wrappers (`go func() { defer recover() ... }`). Hides bugs; locks / maps can leave global state wrong; not error handling. Let panic propagate unless real process boundary + documented recovery (rare).
- Package-level typed maps: prefer `syncmap.SyncMap` (`github.com/lariv-in/lago/syncmap`). Global map → pointer so value never copied after use: `var m *syncmap.SyncMap[K,V] = &syncmap.SyncMap[K,V]{}` (see `plugins/p_seer_reddit/runner_worker_pool.go`). Zero value works; pointer = stable identity + "do not copy" rule.
- `time.Time` → respect TZ. `$tz` in context = `*time.Location`.

## Components

### Core Interfaces

- New component → implement `PageInterface` (`components/page.go`).
- Children → `ParentInterface` (`components/parent.go`) for traversal.
- Mutable children → `MutableParentInterface` (`components/parent.go`).
- Input → `InputInterface` (`components/input.go`) so `FormComponent` finds / parses fields.
- Almost always: `&component` before inserting into thing needing `PageInterface`. Else no `MutableParentInterface` → children not patchable.

### Component Patching

`components/parent.go`: `InsertChildBefore`, `InsertChildAfter`, `ReplaceChild`, `RemoveChild`. Recurse `MutableParentInterface`; match concrete type + `Page.Key`.

- Remove node from other plugin → `RemoveChild` with type + key (e.g. `RemoveChild[*components.ButtonLink](scaffold, "users.AuthSignupLink")`). No custom tree walk by URL/label / slice rebuild unless no stable key yet.
- Page/form meant for extension → stable `Page.Key` in base first; patch on keys. No brittle structure matching when explicit extension point possible.
- Selector route useful beyond one addon → put in base plugin, not addon copy.

## Getters

Request-dependent values → `Getter` from `getters/getters.go`. Shared context keys there: `ContextKeyDB`, `ContextKeyError`, `ContextKeyGet`, `ContextKeyIn`.

**Getters are context-only in UI code. Never call `getters.DBFromContext` (or read `"$db"`) inside plugin `getters.go` or in any `Getter` used by pages** — not even “just one” lookup. Data that needs the database is loaded in a **view `Layer`**, `actions.go`, a query patcher, or similar, then **written onto the request context** (a dedicated `context.WithValue` key, `$in` merge, or a model the layer already put under `getters.Key`). The getter then reads that via `getters.Key` or compositions only.

- `getters/` = topic files as siblings, no subpackages: `key.go`, `deref.go`, `format.go`, `any.go`, `association.go`, `association_list.go`, `join_association_list.go`, `association_ids.go`, `foreign_key.go`, `select.go` / `select_multi.go`, `navigate.go`, `parse_int.go` / `parse_uint.go`, etc. Browse or `rg "func "` for combinators.
- Custom getter → first check `getters/` + small compositions:
  - Nullable `*T` field → `getters.Deref(getters.Key[*T]("$in.Field"))`.
  - Format string from multiple getters → `getters.Format("format", getters.Any(g1), ...)`.
  - Route param `id` → `getters.Any(getters.Key[uint]("$id"))`, not custom `uint→string` wrapper.
  - M2M filter in `$get` → `getters.AssociationIDs(getters.ContextKeyGet, "Field")`, not manual unpack.
- Skip pointless getter fns. No plugin-local `func ...() getters.Getter[T]` that only forwards one combinator (`getters.Key`, `lago.RoutePath`, no extra logic). Use combinator at field. Static HTML attrs → `getters.Static(...)`.
- Named `Getter` in plugins is for **compositions that only read context** (permission/role, `getters.Map`, formatting beyond `getters.Format`, etc.) — not for **using `$db`**. Preloads and association loads that touch GORM live in view layers (with `views.QueryPatcher` / `LayerDetail` / a small `Layer` that stores derived values on `context`) or in `actions.go`, never in a getter.
- Empty placeholder `"—"` for blanks → prefer empty `FieldText` unless product wants dash.
- Getter args → narrowest type. `any` almost always wrong.

## Database Context

- Global HTTP attaches `*gorm.DB` per request under `getters.ContextKeyDB` (`"$db"`). `lago.DBLayer` in `lago/layers.go`.
- Don't scatter `ctx.Value("$db").(*gorm.DB)` / `r.Context().Value("$db")`. Use `getters.DBFromContext(ctx)` (`getters/db_context.go`) → `(*gorm.DB, error)` when missing.
- HTTP handlers → `getters.DBFromContext(r.Context())` or thin plugin wrapper that forwards (`filesystemDB`, `exportDB`, ...).
- Tests → `context.WithValue(ctx, getters.ContextKeyDB, db)` to match prod key.

## Selectors And Relations

### Foreign Keys And Many-To-Many

- FK selector: `InputForeignKey.Name`, selector route/page, `GetterSelect(...)` event name must match. `ParentID` input on `DestinationID` selector → wrong event name → input / modal break.
- Same for `InputManyToMany` + `GetterMultiSelect(...)`. Preserve `target_input` across modal open + filter/browse inside modal. Drop it → wrong field name → chips break.
- `InputForeignKey.Getter` → `getters.Association[T](getters.Key[uint]("$in.FieldID"))`. Table from type `T` via GORM `db.Model()`.
- `InputManyToMany.Getter` → preload + `getters.Key[[]T]("$in.Field")`, not custom lookup. Renders from submitted `AssociationIDs`; update/detail still preload so initial + detail have full rows.
- Detail pages: `components.FieldManyToMany[T]` (`field_manytomany.go`) read-only. Same `Getter` + `Display` as matching `InputManyToMany[T]` (`getters.JoinAssociationList[...]` if join table). `Link` optional: `getters.ContextKeyIn` per row like `Display`, e.g. `lago.RoutePath("plugin.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})`. Prefer over `FieldList` + per-row children for plain association lists. Keep `FieldList` when rows not typed model (e.g. `[]map[string]any`, heavy custom row UI).
- Relation only on join model, not base GORM struct → shared `getters.JoinAssociationList[...]` / `getters.AssociationList[...]` + shared query patchers. No ad-hoc plugin lookup.

### Filesystem Selectors

- `filesystem.SelectRoute`, `filesystem.MoveSelectRoute` → directory pickers; dirs selectable.
- `filesystem.MultiSelectRoute` → file picker for asset M2M; files selectable; dir click → browse, not select.

### Environment Selector

- `components.Environment[T]` (`environment.go`) → `<select>` on `environment` JSON cookie; parsed map `$environment` (`map[string]string`) on context.
- `Options` → `Getter` returning `[]registry.Pair[T, string]`: `Key` = cookie id / `<option value>`, `Value` = label. Keys stringified with `fmt.Sprint` for HTML + cookie compare.
- `Default` optional. Cookie missing key → runs; zero `T` = no default, not "selected id".
- Register `&components.Environment[T]{...}` in `[]PageInterface` / registries needing pointers → `MutableParentInterface`, patchable.
- List views / query patchers scoping by user choice → read same cookie, same parse (e.g. numeric id string for `Environment[uint]`).
- Split environment selector helpers by responsibility: default value getter in consuming plugin `getters.go` (context/DB lookup then call action), model-loading default action in plugin that owns queried model (`actions.go`), selected-cookie parsing + list scoping in consuming plugin `querypatchers.go`. Avoid catch-all `session_env.go` files.

## Choice Fields

Fixed string options (dropdown) → one ordered slice in plugin `models.go` next to GORM struct:

```go
var MyFieldChoices = []registry.Pair[string, string]{
	{Key: "stored_value", Value: "Label in UI"},
	// ...
}
```

- `Key` is persisted value (database column, `<option value>`, form POST). `Value` is label shown in selects and read-only views.
- Slice order is UI option order (`InputSelect`, list filters, etc.). Do not add parallel `const` block or `switch` repeating keys; use same string literals everywhere.
- Detail / read-only string label → use `registry.PairValueFromKey(getters.Key[string]("$in.FieldName"), MyFieldChoices)` with `components.FieldText` or any `Getter[string]` that should show `Value`. Empty stored key → `""`; known key → `Value`; unknown key → raw key. Do not reimplement mapping in one-off getter.
- Forms and filters (`InputSelect[string]`): `Choices: getters.Static(MyFieldChoices)`. For `Getter`, use `registry.PairFromGetter`: pass same choices slice and `getters.Key[string]` to current stored key. Use `$in.FieldName` on create/update/detail form contexts; use `$get.FieldName` on list/selection filter forms.
- `registry.PairFromGetter` returns `getters.Getter[registry.Pair[string, string]]`: empty key → zero pair; known key → pair from slice; unknown legacy key → `{Key: s, Value: s}`.
- If form must show specific default option when stored value is empty (e.g. create flow pre-selects "Enrolled" or "Created"), keep small custom getter. `PairFromGetter` does not map empty → chosen default. Examples: `academicrecords` create status, `assignmentsubmissions` submission status, `studentpayments` create payment method.
- Maps when needed: `registry.MapFromPairs(MyFieldChoices)` builds `map[string]string` for ad-hoc lookups; `registry.PairFromMap` works on that map. Prefer slice + `PairFromPairs` for single lookup; no package-level map required.
- Helpers in `registry/registry.go`: `registry.PairFromPairs`, `registry.PairValueFromKey` (read-only label string), `registry.PairFromGetter` (`InputSelect` current pair), `registry.MapFromPairs`, `registry.PairFromMap`, `registry.PairsFromMap`. `registry.KeysFromPairs` returns keys in slice order (e.g. generators or tests). `registry.PairGetter` is different API (key + `Getter[map[K]V]`, strict map lookup); it does not replace `PairFromGetter` for choice slices.
- Generators, form patchers, and tests must use same `Key` string literals as slice so inserts and validation stay aligned. Example: `plugins/p_nirmancampus_academicrecords` (`AcademicRecordStatusChoices`).

## Registry

App-wide patchability → registry from `registry/registry.go`.

- `Register` adds; `Patch` patches existing entry.
- `lago/registry_commands.go` — commands
- `lago/registry_config.go` — `totschool.toml` fields
- `lago/registry_generators.go` — `generate` command
- `lago/registry_layers.go` — global layers (rare)
- `lago/registry_pages.go` — pages; always pointer to `PageInterface`
- `lago/registry_plugins.go` — plugin metadata (`p_dashboard` apps grid)
- `lago/registry_routes.go` — HTTP routes
- `lago/registry_views.go` — views
- `lago/registry_dbinit.go` — post-DB-init fns; automigrations here

## Plugin Layout

### Page Files

Many pages → several files in plugin root (same package), not `pages/` / `models/` subpackage unless strong reason.

- `pages.go` — `init()` → `registerMenuPages`, `registerFilterPages`, etc.; small shared wiring (sidebars). Keep short.
- `pages_<area>.go` — big trees by concern: `pages_form.go`, `pages_detail.go`, `pages_table.go`, `pages_structure.go`, etc.
- Split heuristic: detail / forms / list-filter-table when large; tiny bits → `pages.go` or nearest big file.
- Target ~200-400 lines/file when practical; cohesion > micro-files unless boundary matters.
- `getters.go` — package-local `func ...() getters.Getter[...]` and small helpers only used to build them, e.g. poll-URL or status helpers.
- **Role-gated “Create” on list tables** (`TableButtonCreate.Link` that depends on `$role`): use `getters.Match(getters.Key[string]("$role"), map[...]getters.Getter[string]{...}, getters.Static(err))` **inline in the `register*TablePages` (or similar) function** next to the register call, not a one-off in `getters.go`. The create URL getter is usually `lago.RoutePath` repeated for `admin` / `superuser` (or whatever roles match the button’s `Page.Roles`).
- Keep `pages_*.go` for `RegistryPage.Register` trees and `PageInterface` assembly; do not add new named getter factories there when `getters.go` exists or is appropriate for plugin.

Example: `plugins/p_nirmancampus_programs`.

### Action Files

- `actions.go` + `actions_<area>.go` when large / separate domains. Plugin root package only.
- Here: fns that take / mutate / load / return models (GORM + related). Domain logic: rules, invariants, orchestration, not pure HTTP/UI wiring.
- Default row lookups used by UI selectors (e.g. active/latest session id) are actions when they query models. Put them in model owner plugin's `actions.go` (e.g. `p_nirmancampus_sessions` for `AdmissionSession`) and call them from consuming plugin getters/query patchers.
- Not here: page trees, view layers/handlers, bare `init()`, getters that only read context for components. Views / commands / generators / other actions call in when behavior = domain.
- Mental model: Django model methods: behavior on model(s), many entry points, no duplicated handler logic.

## Routes

Plugin mounts under host `AppUrl` (e.g. `/students/`) → use `addon/<slug>/` after base so subtree doesn't collide with host `{id}` routes.

- Prefix: `HostAppUrl + "addon/" + short-slug + "/"` (e.g. `p_nirmancampus_students.AppUrl + "addon/academicrecords/"`).
- Why: `http.ServeMux` rejects overlapping patterns. Host has `/prefix/{id}/`, `.../edit/`, `.../delete/`. Naive `/prefix/academicrecords/...` can bind `id=academicrecords` or clash with `delete`. Fixed `addon/` segment disambiguates.
- Nested feature not own dashboard tile → prefer `PluginTypeAddon`; link from host menu/UI.

## Views

View = primary HTTP handler for route. `*views.View` (`views/views.go`):

- Which `PageInterface` (`PageName` + `PageLookup`).
- Ordered `Layers` (`views.Layer` → `Next(View, http.Handler) http.Handler`).

Global concerns (`getters.ContextKeyDB`, `$request`, etc.) → `views.GlobalLayer` + app registration, not inside view struct. Routes: `lago.GetPageView("plugin.PageName")`, then `WithLayer("stable.key", layer)`.

### Redirects

Use `views.HtmxRedirect(w, r, url, code)` (`htmx_redirect.go`) instead of raw `http.Redirect` for user navigation. `HX-Request` → `HX-Redirect` + 200; else `http.Redirect` with same `code`. Rare bypass only. `lago.RedirectView` + `RedirectLayer` + `views.HtmxRedirect(..., http.StatusMovedPermanently)`.

### Typed CRUD Layers

One concern each; order matters:

- `LayerList[T]` — paginated list from URL; `components.ObjectList[T]` in context under `Key`; merge filters into `$get`; coerce from page's first form if present. Fail → `_global` in `getters.ContextKeyError`, `next` (no HTTP error response).
- `LayerDetail[T]` — load one row by path PK; before update/delete needing same row. Same error pattern on fail.
- `LayerCreate[T]` — POST create; success sets `$id`.
- `LayerUpdate[T]` — POST update; record usually already in ctx (`LayerDetail` before).
- `LayerDelete[T]` — POST delete (not HTTP DELETE; matches confirm forms).
- `LayerSingleton[T]` — singleton GET/POST load/create.
- `LayerJsonImport[T]` — JSON import.
- `MethodLayer` — custom method handler.

### Query Patching

`views.QueryPatchers[T]` (`registry.Pair`s) on `LayerList` / `LayerDetail` / `LayerUpdate`. Prefer `views/query_patcher_*.go`: `QueryPatcherPreload[T]`, `QueryPatcherOrderBy[T]`, `QueryPatcherJoinFilter[T,TJoin]` (reads `$get`). No duplicate ad-hoc query when these enough.

- `QueryPatcherPreload[T]` (`query_patcher_preload.go`): `Fields []string` = GORM association names / dotted paths (same as `Preload`). Order preserved. `Fields` empty → no preloads that patcher.
- One preload patcher per layer: single `registry.Pair`, one `QueryPatcherPreload[T]{Fields: [...]}` listing all associations for layer. Stable key e.g. `"myplugin.preload"` so other plugins replace/wrap one logical hook.
- Plugin-specific query patchers live in plugin `querypatchers.go`. If filter parses `$environment`, `$get`, route params, or joins/subqueries to scope list, keep parsing next to patcher.

### Form Patching

`views.FormPatchers` on `LayerCreate` / `LayerUpdate` (`form_patchers.go`). `InputManyToMany.Parse` → `AssociationIDs`; create/update/singleton persists M2M via GORM after row save, not plain scalar columns.

- Form patcher map: after `view.ParseForm`, each entry = concrete type from that input's `Parse`.
- Validation: one type per field — one assertion or one owned `case`, not `switch` across `uint` / `*uint` / `int` / etc.
- Repo norms: `InputForeignKey` → `uint`; `InputDuration` → `*time.Duration`.
- Missing / wrong type / invalid (nil duration, zero id when forbidden) → field error; fix input or patcher contract, don't widen types.

### Cross-Plugin View Patches

- Stable string key per layer (`"students.detail"`). Other packages → `InsertLayerBefore` / `InsertLayerAfter` / `PatchLayer` on those keys. No fragile position.
- Extra context on another plugin's page (e.g. related `ObjectList` on base detail): don't hide DB in component getter. Small `views.Layer`, load in `Next`, `context.WithValue`, register/patch after parent-record layer (`InsertLayerAfter("base.detail", "myplugin.extra", ...)`). `DataTable` → `getters.Key[...]` on that key.
- Custom forms / errors: `view.ParseForm` + `views.ContextWithErrorsAndValues`. Don't resurrect removed `HasErrors` / `RenderWithErrors` patterns.

## SQL Identifiers

- PostgreSQL columns `start`, `end` → quote in GORM `Order` / raw SQL: `` Order(`"start" ASC`) ``, `Where(`"start" <= ? AND "end" >= ?`, ...)`. Unquoted `start` invalid in PostgreSQL.

## Error Handling

- Wrong state from error → prefer panic over limping.
- Recoverable error → always log.
- Edge cases → log. Never silent ignore.
- `log/slog` recoverable; `log.Panicf` non-recoverable.

## Generators

- Register: `lago.RegistryGenerator.Register("name", lago.Generator{Create: ..., Remove: ...})` in `init()`.
- Order = deployment TOML `GeneratorOrder` (e.g. `nirmancampus.toml`).
- `RunGenerators`: Phase 1 Remove — reverse TOML order (dependents before bases). Phase 2 Create — forward (bases before dependents).
- M2M cleanup: `Remove` must raw-SQL clear join tables (`db.Exec("DELETE FROM student_assets")`) before deleting primary row. GORM/PostgreSQL won't cascade M2M joins → FK errors otherwise.
