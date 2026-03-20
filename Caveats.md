# Caveats When Working On This Codebase

- NEVER write go.mod or go.sum manually, use go mod init, go mod tidy -e, go work use for project management
- In nearly all cases, take the address of components before inserting them into something that requires `PageInterface`. Otherwise, the value will not implement `MutableParentInterface` and its children will not be patchable.

- When creating new components, they should implement at least `PageInterface` from `components/page.go`.

- If a component has children, it should implement `ParentInterface` from `components/parent.go` so its children can be traversed.
- If a component allows modifying its children, it should implement `MutableParentInterface` from `components/parent.go`.
- If a component is an input, it needs to implement `InputInterface` from `components/input.go` so that `FormComponent` can detect it and parse its fields.

- Whenever something requires a value that can depend on the request, it should use a `Getter` from `getters/getter.go`.

- Before writing a custom getter, always check if an existing getter can't cover the use case:
   - Use `getters.GetterDeref(getters.GetterKey[*T]("$in.Field"))` for nullable pointer fields instead of writing custom wrapper functions.
   - Use `getters.GetterFormat("format", getters.GetterAny(getter1), ...)` to combine multiple getters into a formatted string instead of custom inline functions.
   - For route params like `id`, prefer `getters.GetterAny(getters.GetterKey[uint]("$id"))` instead of writing custom `uint -> string` wrapper getters.
   - For many-to-many filter state stored in `$get`, prefer `getters.GetterContextAssociationIDs(getters.ContextKeyGet, "Field")` instead of manually unpacking `AssociationIDs`.

- When defining getter arguments, use the most restrictive type possible. `any` is almost always a bad idea.

- For foreign key selectors, the `InputForeignKey.Name`, the selector route/page it opens, and the `GetterSelect(...)` event name all need to match. If a `ParentID` input opens a selector table built for `DestinationID`, the selection event will be dispatched with the wrong name and the input will not update or close its modal.

- The same name-matching rule applies to `InputManyToMany` and `GetterMultiSelect(...)`. Many-to-many selectors also need to preserve `target_input` across the initial modal open and any filter/browse requests inside the modal. If `target_input` is dropped, the selector will dispatch the wrong field name and the chips will not update.

- For `InputForeignKey.Getter`, use `getters.GetterAssociation[T](getters.GetterKey[uint]("$in.FieldID"))`. It infers the table name from the type `T` via GORM's `db.Model()`.

- For `InputManyToMany.Getter`, prefer preloaded associations plus `getters.GetterKey[[]T]("$in.Field")` instead of custom lookup getters. `InputManyToMany` re-renders from submitted `AssociationIDs`, but update/detail views should still preload the association so initial render and detail pages have the full related objects available.

- If the relation is intentionally not declared on the base GORM model and is instead represented by a separate join model, prefer shared getters such as `getters.GetterJoinAssociationList[...]` / `getters.GetterAssociationList[...]` plus shared query patchers instead of ad-hoc plugin-local lookup code.

- Models are not patchable through registries the way pages and views are. If a plugin needs to extend another plugin's data model, prefer a separate extension/join model owned by the new plugin plus page/view/query patches around it. Only add fields directly to the base GORM model when that relationship truly belongs in the base plugin and is intended to be a first-class part of that model.

- The filesystem selector routes have two different behaviors:
   - `filesystem.SelectRoute` and `filesystem.MoveSelectRoute` are directory pickers; directories are selectable.
   - `filesystem.MultiSelectRoute` is a file picker for asset-style many-to-many fields; files are selectable, but clicking a directory should browse into it instead of selecting it.

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
   - `lago/registry_middlewares.go` for adding global middlewares, generally not needed
   - `lago/registry_pages.go` for adding pages; always insert a pointer to a `PageInterface` implementer
   - `lago/registry_plugins.go` for adding plugin information, primarily used by `p_dashboard/components/apps_grid.go`
   - `lago/registry_routes.go` for adding HTTP routes
   - `lago/registry_views.go` for adding views (see the views section below)
   - `lago/regsitry_dbinit.go` for adding functions that run after the database is initialized; run model automigrations here

# Views

A view is the primary HTTP handler here. A view keeps track of:
   - which `PageInterface` to display
   - which middlewares to run for the current route, excluding global middlewares
   - custom functions for patching the query
   - custom functions for handling form data after parsing and before it is used elsewhere

Currently, the following view factories exist in `views/crud.go`:
   - `CreateView`
   - `ListView`
   - `DetailView`
   - `UpdateView`
   - `DeleteView`
   - `SingletonView`

- `ListView` is used whenever we need an `ObjectList` of a model.
- `DetailView` is used whenever we need a single instance of a model.

- In most cases, `DetailView` needs to be chained with `UpdateView` and `DeleteView`.

- For common query patching behavior, prefer shared helpers in `views/query_patcher.go` over custom per-plugin functions:
   - Use `views.QueryPatcherPreload("AssociationName")` for preloads.
   - Use `views.QueryPatcherOrderBy("field ASC|DESC")` for ordering.
   - Use `views.QueryPatcherJoinFilter[...]` when filtering a base list view through a separate join model.

- Generic CRUD many-to-many support now exists in `views/crud.go`. `InputManyToMany.Parse` returns a typed `AssociationIDs` payload, and `CreateView` / `UpdateView` / `SingletonView` persist it through GORM associations after the base row save. Do not try to make many-to-many form inputs look like ordinary scalar columns.

# Error handling

Error handling is very important. If an error occurs that could make the program behave incorrectly, it is preferable to panic rather than keep it running in a bad state.

Whenever a recoverable error occurs, it should be logged, no matter how unlikely.

All edge cases need to be logged, no edge case should ever be ignored.

Use `log/slog` for recoverable errors and `log.Panicf()` for non-recoverable errors.

# Component Patching

Use `components.InsertChildBefore`, `components.InsertChildAfter`, and `components.ReplaceChild`.

- If a page or form is intended to be extended by another plugin, add stable `Page.Key` values in the base plugin first, then patch against those keys. Do not rely on brittle structural matching when a reusable extension point can be made explicit.

- If a selector route/page is generally useful beyond one addon, add it to the base plugin instead of creating a one-off copy in the addon plugin.
