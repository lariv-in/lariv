# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Hot-reload development server (uses .air.toml)
air

# Build manually
go build -o ./tmp/main ./totschool_lago

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
| `components/` | Declarative UI components built with `gomponents` |
| `views/` | `View` struct — maps HTTP methods to handler functions |
| `p_users/` | Auth plugin: JWT cookies, User/Role models, login/signup pages |
| `p_dashboard/` | Dashboard plugin: apps grid, topbar buttons |
| `totschool_lago/` | Application entry point (main package) |

### Registry Pattern

The central pattern is a generic `Registry[T]` in `lago/registry.go`. Every extension point is a registry:

- `lago.RegistryRoute` — URL path → `Route{Path, Handler}`
- `lago.RegistryPage` — name → `PageInterface` (declarative UI trees)
- `lago.RegistryView` — name → `http.Handler`
- `lago.RegistryMiddleware` — name → `Middleware`
- `lago.RegistryPlugins` — name → `Plugin` (app metadata for the dashboard)

Registries support `Register(name, item)` and `Patch(name, func(T) T)` — patch lets plugins modify items registered by other plugins without direct dependencies.

### Plugin Activation via `init()`

All plugins register themselves in `init()` functions. The application activates plugins by blank-importing them in `totschool_lago/main.go`:

```go
import (
    _ "github.com/lariv-in/p_users"
    _ "github.com/lariv-in/p_dashboard"
)
```

Each plugin's `init()` calls register on the appropriate registries.

### Page → View → Route pipeline

1. **Pages** are declarative component trees registered in `RegistryPage` (e.g. `"users.LoginPage"`).
2. **Views** wrap a page and add per-HTTP-method handlers, registered in `RegistryView` (e.g. `"users.LoginView"`). `lago.GetPageView(pageName)` creates a view with a default GET handler that renders the page.
3. **Routes** map URL paths to `DynamicView` handlers that look up the view by name at request time, registered in `RegistryRoute`.

### Getter Pattern

`Getter` is `func(context.Context) any` — used in components to lazily resolve data from context at render time. Key context values:

- `"$db"` — `*gorm.DB` injected by `MiddlewareDb`
- `"$user"` — authenticated `User` injected by `AuthMiddleware`
- `"$error.<field>"` — form validation errors (e.g. `"$error.email"`)
- `"$in.<field>"` — previously submitted form values for re-population

Helper constructors: `GetterStatic(v)`, `GetterKey("dot.path")`, `GetterNil()`, `GetterFormat(fmt, getters...)`.

### DB Initialization

Plugins register GORM migrations via `lago.OnDbInit(func(*gorm.DB) *gorm.DB)` in their `init()`. The DB (SQLite, file `test.db`) is initialized in `lago.Start()` before middlewares are applied.

### Shell Scaffolds

Pre-composed layout wrappers in `components/` that implement the `Shell` interface for HTMX boosted partial body rendering:
- `ShellAuthScaffold` — centered auth card
- `ShellTopbarScaffold` — page with top navigation bar
- `ShellScaffold` — topbar + sidebar layout

### Form Handling

`FormComponent.ParseForm(r)` iterates child components implementing `InputInterface`, calling each input's `Parse()` to validate and clean the value. Returns `(values map[string]any, errors map[string]error, err error)`. Input types: `InputText`, `InputEmail`, `InputPassword`, `InputPhone`, `InputCheckbox`.

### CRUD Views (Generics)

`views/crud.go` provides generic CRUD view wrappers: `ListView[T]`, `DetailView[T]`, `CreateView[T]`, `UpdateView[T]`, `DeleteView[T]`. They accept a model instance for type inference (e.g. `views.CreateView(User{}, "/users/%v/")`). Form values are mapped to struct fields via `applyValues` (snake_case → PascalCase + type conversion).

## Common Pitfalls

### Route conflicts with Go 1.22 ServeMux
Nested resource paths like `/users/roles/{id}/` conflict with `/users/{id}/delete/` because neither pattern is more specific. Give sub-resources their own top-level path (e.g. `/roles/` instead of `/users/roles/`).

### Form field names must match DB column names
`InputForeignKey` and other inputs submit values under their `Name` field. This must match the GORM column name (snake_case of the struct field). E.g. for a `RoleID int` struct field, use `Name: "role_id"`, not `Name: "role"`.

### `InputForeignKey` selection event name must match input Name
`GetterSelect(name, ...)` dispatches an event with the given name. The `InputForeignKey` listens for events matching its own `Name` field. These must be identical — e.g. if the input has `Name: "role_id"`, use `GetterSelect("role_id", ...)`.

### `MapFromStruct` flattens embedded structs
`components.MapFromStruct` promotes fields from anonymous embedded structs (like `gorm.Model`) to the top level. So `ID`, `CreatedAt`, etc. are accessed directly (e.g. `$row.ID`), not nested under `$row.Model.ID`.

### Use `RoutePathGetter` not `RegistryRoute.Getter` for URLs
`RegistryRoute.Getter(name)` returns the full `Route` struct. For URL strings (in component attrs, topbar buttons, etc.), use `lago.RoutePathGetter(name)` which returns just the path string.

### Alpine `x-data` scope matters for `Alpine.$data()`
`Alpine.$data(el)` only works on elements that have `x-data`. The theme data lives on `<body>`, so use `Alpine.$data(document.body)`, not `document.documentElement`.

### HTMX uses `hx-*` attributes, not Turbo
This app uses HTMX, not Turbo Drive. For non-GET requests use `hx-post`, `hx-put`, `hx-delete` (not `data-turbo-method`). For targeting use `hx-target` (not `data-turbo-frame`).

### HTMX relative selectors
- `closest <sel>` — searches **ancestors** (up the DOM)
- `next <sel>` — searches **next siblings** (forward in DOM)
- `find <sel>` — searches **descendants** (down the DOM)

Use `next` for sibling elements, not `closest`.
