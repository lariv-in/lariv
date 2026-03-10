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

### Layout Scaffolds

Pre-composed layout wrappers in `components/`:
- `LayoutAuthScaffold` — centered auth card
- `LayoutTopbarScaffold` — page with top navigation bar
- `LayoutScaffold` — topbar + sidebar layout

### Form Handling

`FormComponent.ParseForm(r)` iterates child components implementing `InputInterface`, calling each input's `Parse()` to validate and clean the value. Returns `(values map[string]any, errors map[string]error, err error)`. Input types: `InputText`, `InputEmail`, `InputPassword`, `InputPhone`, `InputCheckbox`.
