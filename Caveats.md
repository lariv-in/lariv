# Caveats when working on this codebase

- In nearly all cases, we should take the address of components before inserting into something that requires PageInterface. Not doing so means that the value will not implement MutableParentInterface and its children won't be patchable.

- When creating new components, The new component should implement at least PageInterface from components/page.go

- If the component has children, then it should implement ParentInterface from components/parent.go to allow for traversing its children.
- If the component allows for modifying its children, then it should implement MutableParentInterface from components/parent.go
- If the component is a input, then it needs to implement InputInterface from components/input.go so that the FormComponent can detect it and parse its fields

- Whenever anything requires a value that can depend on the request, it should use a Getter from @getters/getter.go.

- When defining getter arguments, the type should be the most restrictive type possible, 'any' type is almost always a bad idea.

# Registry

Any thing that needs to be pactchable on an app wide scale should be done via a registry from registry/registry.go

Registries use Register method to add to the registry and Patch method to register a patch for an already existing element in the registry.

Existing registries: 
   - lago/registry_commands.go - for adding custom commands
   - lago/registry_config.go - for adding config fields to totschool.toml
   - lago/registry_generators.go - for adding generators that will run when generate command is run
   - lago/registry_middlewares.go - for adding global middlewares, generally not needed
   - lago/registry_pages.go - for adding pages, always insert a pointer to a PageInterface implementer
   - lago/registry_plugins.go - for adding plugin information, primarily used by p_dashboard/components/apps_grid.go
   - lago/registry_routes.go - for adding http routes
   - lago/registry_views.go - for adding views (see the views section below)
   - lago/regsitry_dbinit.go - for adding functions that will run after database is initialised, run automigrate of models here.

# Views

A view is the primary http handler here. A view keeps track of what PageInterface to display, what middlewares to run for the current route, except for the global middlewares. Custom functions for patching the query, Custom functions for handling form data after being parsed, before being used by other functions.

Currently, there are the following views factories in views/crud.go:
    - CreateView
    - ListView
    - DetailView
    - UpdateView
    - DeleteView
    - SingletonView

List view is needed whenever we need a ObjectList of a model
Detail view is needed whenever we need a singl instance of a model

Almost always, DetailView needs to be chained with Update and Delete view

# Error handling

Error handling is very important. If a error occurs that might give make the program behave in a incorrect manner, it is prefereable to make it panic, rather than keeping it running.

Whenever a recoverable error occurs, then it should be logged, no matter how unlikely.

All edge cases need to be logged, no edge case should ever be ignored.

Use "log/slog" for recoverable errors and log.Panicf() for non recoverable errors


# Component Patching

use components.InsertChildBefore, components.InserChildAfter and components.ReplaceChild
