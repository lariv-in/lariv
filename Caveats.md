# Caveats When Working On This Codebase

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

- When defining getter arguments, use the most restrictive type possible. `any` is almost always a bad idea.

- For foreign key selectors, the `InputForeignKey.Name`, the selector route/page it opens, and the `GetterSelect(...)` event name all need to match. If a `ParentID` input opens a selector table built for `DestinationID`, the selection event will be dispatched with the wrong name and the input will not update or close its modal.

- For `InputForeignKey.Getter`, use `getters.GetterAssociation[T](getters.GetterKey[uint]("$in.FieldID"))`. It infers the table name from the type `T` via GORM's `db.Model()`.

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

# Error handling

Error handling is very important. If an error occurs that could make the program behave incorrectly, it is preferable to panic rather than keep it running in a bad state.

Whenever a recoverable error occurs, it should be logged, no matter how unlikely.

All edge cases need to be logged, no edge case should ever be ignored.

Use `log/slog` for recoverable errors and `log.Panicf()` for non-recoverable errors.

# Component Patching

Use `components.InsertChildBefore`, `components.InsertChildAfter`, and `components.ReplaceChild`.
