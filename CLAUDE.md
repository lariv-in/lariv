# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Hot-reload development server (uses .air.toml)
air

# Build manually
go build -o ./tmp/main ./deployments/totschool_lago

# Run
./tmp/main

# Run tests (components package has the only tests)
cd components && go test ./...

# Run a single test
cd components && go test -run TestInputImplementations
```

## Architecture

This is a Go web framework called **lago** structured as a `go.work` monorepo with these modules:

| Module | Role |
|---|---|
| `lago/` | Core framework: registries, routing, middleware, DB init |
| `registry/` | Generic `Registry[T]` type used by all extension points |
| `getters/` | `Getter` type (`func(context.Context) any`) and helpers |
| `components/` | Declarative UI components built with `gomponents` |
| `views/` | `View` struct — maps HTTP methods to handler functions |
| `plugins/p_users/` | Auth plugin: JWT cookies, User/Role models, login/signup pages |
| `plugins/p_dashboard/` | Dashboard plugin: apps grid, topbar buttons |
| `plugins/p_otp/` | OTP plugin: SMS/email OTP login, preferences management |
| `deployments/totschool_lago/` | Application entry point (main package) |

### Registry Pattern

The central pattern is a generic `Registry[T]` in `registry/`. Every extension point is a registry:

- `lago.RegistryRoute` — URL path → `Route{Path, Handler}`
- `lago.RegistryPage` — name → `PageInterface` (declarative UI trees)
- `lago.RegistryView` — name → `http.Handler`
- `lago.RegistryMiddleware` — name → `Middleware`
- `lago.RegistryPlugins` — name → `Plugin` (app metadata for the dashboard)

Registries support `Register(name, item)` and `Patch(name, func(T) T)` — patch lets plugins modify items registered by other plugins without direct dependencies.

### Plugin Activation via `init()`

All plugins register themselves in `init()` functions. The application activates plugins by blank-importing them in `deployments/totschool_lago/main.go`:

```go
import (
    _ "github.com/lariv-in/p_users"
    _ "github.com/lariv-in/p_dashboard"
    _ "github.com/lariv-in/p_otp"
)
```

Each plugin's `init()` calls register on the appropriate registries.

### Page → View → Route pipeline

1. **Pages** are declarative component trees registered in `RegistryPage` (e.g. `"users.LoginPage"`).
2. **Views** wrap a page and add per-HTTP-method handlers, registered in `RegistryView` (e.g. `"users.LoginView"`). `lago.GetPageView(pageName)` creates a view with a default GET handler that renders the page.
3. **Routes** map URL paths to `DynamicView` handlers that look up the view by name at request time, registered in `RegistryRoute`.

### Getter Pattern

`Getter` is `func(context.Context) any` — defined in the `getters/` module. Used in components to lazily resolve data from context at render time. Key context values:

- `"$db"` — `*gorm.DB` injected by `MiddlewareDb`
- `"$user"` — authenticated `User` injected by `AuthMiddleware`
- `"$error.<field>"` — form validation errors (e.g. `"$error.email"`)
- `"$in"` — `map[string]any` of form values for pre-population (accessed via `GetterKey("$in.fieldname")`)

Helper constructors: `GetterStatic(v)`, `GetterKey("dot.path")`, `GetterNil()`, `GetterFormat(fmt, getters...)`, `GetterQueryEscape(getter)`.

### DB Initialization

Plugins register GORM migrations via `lago.OnDBInit(func(*gorm.DB) *gorm.DB)` in their `init()`. The DB (SQLite, file `test.db`) is initialized in `lago.Start()` before middlewares are applied.

### Shell Scaffolds

Pre-composed layout wrappers in `components/` that implement the `Shell` interface for HTMX boosted partial body rendering:
- `ShellAuthScaffold` — centered auth card
- `ShellTopbarScaffold` — page with top navigation bar
- `ShellScaffold` — topbar + sidebar layout

### Form Handling

`FormComponent.ParseForm(r)` iterates child components implementing `InputInterface`, calling each input's `Parse()` to validate and clean the value. Returns `(values map[string]any, errors map[string]error, err error)`. Input types: `InputText`, `InputEmail`, `InputPassword`, `InputPhone`, `InputCheckbox`.

### View Helper Methods

`View` has helper methods for common handler patterns:
- `v.ParseForm(w, r)` — finds and parses the first form in the view's page
- `v.RenderWithErrors(w, r, fieldErrors, values)` — re-renders with errors and `$in` values in context
- `views.HasErrors(errs)` — checks if any field error is non-nil

### CRUD Views (Generics)

`views/crud.go` provides generic CRUD view wrappers: `ListView[T]`, `DetailView[T]`, `CreateView[T]`, `UpdateView[T]`, `DeleteView[T]`, `SingletonView[T]`. They accept a model instance for type inference (e.g. `views.CreateView(User{}, "/users/%v/")`).

`SingletonView[T](model, successUrlGetter)` handles singleton config forms — loads via `FirstOrCreate` into `$in` context for GET, parses + updates on POST. The `successUrl` is a `Getter` (use `lago.RoutePathGetter("route.Name")` to pull from the route registry).

### `GetterKey` resolves `$in` as a map, not flat keys

`GetterKey("$in.fieldname")` splits on `"."` and first looks up `ctx.Value("$in")` expecting a `map[string]any`, then navigates into `"fieldname"`. Setting flat context keys like `context.WithValue(ctx, "$in.fieldname", val)` will NOT work — you must set `"$in"` as a single `map[string]any`.

### Handlers must be registered for the methods they handle

`lago.GetPageView(name)` only registers a default GET handler. If your handler function checks `r.Method == http.MethodGet` internally, it must also be registered for GET via `WithMethod(http.MethodGet, handler)`, otherwise the default handler runs instead and your code is dead.

## Common Pitfalls

### Route conflicts with Go 1.22 ServeMux
Nested resource paths like `/users/roles/{id}/` conflict with `/users/{id}/delete/` because neither pattern is more specific. Give sub-resources their own top-level path (e.g. `/roles/` instead of `/users/roles/`).

### Form field names must match DB column names
`InputForeignKey` and other inputs submit values under their `Name` field. This must match the GORM column name (snake_case of the struct field). E.g. for a `RoleID int` struct field, use `Name: "role_id"`, not `Name: "role"`.

### `InputForeignKey` selection event name must match input Name
`GetterSelect(name, ...)` dispatches an event with the given name. The `InputForeignKey` listens for events matching its own `Name` field. These must be identical — e.g. if the input has `Name: "role_id"`, use `GetterSelect("role_id", ...)`.

### `MapFromStruct` flattens embedded structs
`getters.MapFromStruct` promotes fields from anonymous embedded structs (like `gorm.Model`) to the top level. So `ID`, `CreatedAt`, etc. are accessed directly (e.g. `$row.ID`), not nested under `$row.Model.ID`.

### Use `RoutePathGetter` not `RegistryRoute.Getter` for URLs
`RegistryRoute.Getter(name)` returns the full `Route` struct. For URL strings (in component attrs, topbar buttons, etc.), use `lago.RoutePathGetter(name)` which returns just the path string.

### URL-encode identifiers in query params
Phone numbers with `+` (e.g. `+91...`) must be URL-encoded when passed as query params. Use `url.QueryEscape()` in handlers or `getters.GetterQueryEscape(getter)` in page definitions, otherwise `+` is decoded as a space.

### Alpine `x-data` scope matters for `Alpine.$data()`
`Alpine.$data(el)` only works on elements that have `x-data`. The theme data lives on `<body>`, so use `Alpine.$data(document.body)`, not `document.documentElement`.

### HTMX uses `hx-*` attributes, not Turbo
This app uses HTMX, not Turbo Drive. For non-GET requests use `hx-post`, `hx-put`, `hx-delete` (not `data-turbo-method`). For targeting use `hx-target` (not `data-turbo-frame`).

### HTMX relative selectors
- `closest <sel>` — searches **ancestors** (up the DOM)
- `next <sel>` — searches **next siblings** (forward in DOM)
- `find <sel>` — searches **descendants** (down the DOM)

Use `next` for sibling elements, not `closest`.
