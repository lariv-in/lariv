# Views API Migration Guide (For Dummies)

This is the "what do I actually replace this with?" guide for moving from the old views API to the new middleware-based API.

Use this document as the migration playbook. The old `Caveats.md` views section still talks about removed factories, so do not copy that old pattern back in.

## The New Mental Model

Old API:

- A view had special helpers like `ListView`, `DetailView`, `CreateView`, `UpdateView`, `DeleteView`, `SingletonView`, `JsonImport`, `WithMethod`, `WithQueryPatcher`, `WithFormPatcher`, `WithRenderMiddleware`, `Handlers`, and `lago.NewRedirectView`.

New API:

- A route maps to a plain `*views.View`.
- A `views.View` is just:
  - a page name
  - a page lookup function
  - an ordered list of middlewares
- Every piece of behavior lives in a middleware.
- Middleware order is the behavior order.

Start almost every route like this:

```go
lago.GetPageView("plugin.SomePage")
```

Then append middlewares in the order you want them to run.

## The Big Rule

Ask one question:

"Which middleware owns this step?"

Use:

- `views.MiddlewareList[T]` for paginated list queries
- `views.MiddlewareDetail[T]` for loading one row by path param
- `views.MiddlewareCreate[T]` for POST create
- `views.MiddlewareUpdate[T]` for POST update
- `views.MiddlewareDelete[T]` for POST delete
- `views.MiddlewareSingleton[T]` for singleton settings pages
- `views.MiddlewareJsonImport[T]` for JSON upload/import
- `views.MethodMiddleware` for custom GET/POST handlers
- a custom `views.Middleware` for extra context loading / render-time enrichment

## Old To New Cheat Sheet

- `views.ListView[T]("items")(page).WithQueryPatcher(...)`
  -> `page.WithMiddleware("plugin.list", views.MiddlewareList[T]{...})`

- `views.DetailView[T]("item", "id")(page).WithQueryPatcher(...)`
  -> `page.WithMiddleware("plugin.detail", views.MiddlewareDetail[T]{...})`

- `views.CreateView[T](successURL)(page).WithFormPatcher(...)`
  -> `page.WithMiddleware("plugin.create", views.MiddlewareCreate[T]{...})`

- `views.UpdateView[T]("id", successURL)(page)`
  -> usually `MiddlewareDetail[T]` then `MiddlewareUpdate[T]`

- `views.DeleteView[T]("id", successURL)(page)`
  -> usually `MiddlewareDetail[T]` then `MiddlewareDelete[T]`

- `views.SingletonView[T](successURL)(page)`
  -> `page.WithMiddleware("plugin.singleton", views.MiddlewareSingleton[T]{...})`

- `views.JsonImport[T](fileField, successURL)(page)`
  -> `page.WithMiddleware("plugin.import", views.MiddlewareJsonImport[T]{...})`

- `WithMethod(http.MethodPost, handler)`
  -> `WithMiddleware("plugin.action", views.MethodMiddleware{Method: http.MethodPost, Handler: handler})`

- `WithQueryPatcher(...)`
  -> put `QueryPatchers` inside `MiddlewareList`, `MiddlewareDetail`, or `MiddlewareUpdate`

- `WithFormPatcher(...)`
  -> put `FormPatchers` inside `MiddlewareCreate` or `MiddlewareUpdate`

- `WithRenderMiddleware(...)`
  -> replace with an ordinary custom `views.Middleware`

- `lago.NewRedirectView(...)`
  -> `lago.RedirectView(...)` or `lago.Redirect(...)`

- `Handlers: map[string]func(*views.View) http.Handler{...}`
  -> one or more `views.MethodMiddleware`

## Core Design Conventions

- Never add new legacy API usage. Do not introduce `ListView`, `DetailView`, `CreateView`, `UpdateView`, `DeleteView`, `SingletonView`, `JsonImport`, `WithMethod`, `WithQueryPatcher`, `WithFormPatcher`, `WithRenderMiddleware`, `Handlers`, or `lago.NewRedirectView`.

- All per-view middleware must implement `views.Middleware`. Do not use raw `func(http.Handler) http.Handler`.

- All global HTTP middleware must implement `views.GlobalMiddleware`. Do not use ad-hoc global middleware funcs either.

- Give every middleware a stable key like `"users.detail"` or `"forms.path_params"`. These keys are what other plugins patch against.

- When patching another plugin's view, use `InsertMiddlewareBefore` or `InsertMiddlewareAfter`. Do not depend on brittle ordering assumptions if you can anchor to a stable middleware key.

- Middleware order matters. A good default order is:
  - auth / role middleware
  - path/context helpers
  - list/detail data loading
  - extra context enrichment
  - create/update/delete/custom action middleware

- `MiddlewareDetail[T]` is the owner of "load one row from path param". If update/delete/custom detail logic needs the record, put `MiddlewareDetail[T]` first and reuse the same context key.

- `MiddlewareDelete[T]` runs on `POST`, not `DELETE`. This is intentional so it matches the existing delete confirmation form flow.

- Redirects belong in middleware or handlers, not in fake legacy redirect views.

- Prefer built-in typed query patchers before writing custom ones:
  - `views.QueryPatcherPreload[T]{Field: "..."}`
  - `views.QueryPatcherOrderBy[T]{Order: "..."}`
  - `views.QueryPatcherJoinFilter[T, TJoin]{...}`

- Use `getters.Static(...)` for fixed context keys, path param names, and other constant getter values.

- Use `lago.RoutePath(...)` for redirect URLs, not hand-built route strings.

- If a custom handler needs to re-render a page with form errors, inline that logic with `views.ContextWithErrorsAndValues(...)`. Do not recreate `HasErrors` or `RenderWithErrors`.

## Basic Migration Recipes

### 1. Plain Page View

```go
lago.RegistryView.Register("dashboard.AppsView",
	lago.GetPageView("dashboard.AppsPage").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}))
```

### 2. List View

```go
lago.RegistryView.Register("contacts.ListView",
	lago.GetPageView("contacts.ContactTable").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("contacts.list", views.MiddlewareList[Contact]{
			Key: getters.Static("contacts"),
			QueryPatchers: views.QueryPatchers[Contact]{
				{Key: "contacts.order_name", Value: views.QueryPatcherOrderBy[Contact]{Order: "name ASC"}},
			},
		}))
```

### 3. Detail View

```go
lago.RegistryView.Register("contacts.DetailView",
	lago.GetPageView("contacts.ContactDetail").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
			Key:          getters.Static("contact"),
			PathParamKey: getters.Static("id"),
		}))
```

### 4. Create View

```go
lago.RegistryView.Register("contacts.CreateView",
	lago.GetPageView("contacts.ContactCreateForm").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("contacts.create", views.MiddlewareCreate[Contact]{
			SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$id")),
			}),
		}))
```

Notes:

- `MiddlewareCreate[T]` sets `$id` after a successful create.
- Use that `$id` in the success route.

### 5. Update View

```go
lago.RegistryView.Register("contacts.UpdateView",
	lago.GetPageView("contacts.ContactUpdateForm").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
			Key:          getters.Static("contact"),
			PathParamKey: getters.Static("id"),
		}).
		WithMiddleware("contacts.update", views.MiddlewareUpdate[Contact]{
			Key: getters.Static("contact"),
			SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$id")),
			}),
		}))
```

### 6. Delete View

```go
lago.RegistryView.Register("contacts.DeleteView",
	lago.GetPageView("contacts.ContactDeleteForm").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
			Key:          getters.Static("contact"),
			PathParamKey: getters.Static("id"),
		}).
		WithMiddleware("contacts.delete", views.MiddlewareDelete[Contact]{
			Key:        getters.Static("contact"),
			SuccessURL: lago.RoutePath("contacts.DefaultRoute", nil),
		}))
```

Important:

- The delete confirmation form submits `POST`.
- `MiddlewareDelete[T]` expects the record to already be in context under `Key`.

### 7. Singleton View

```go
lago.RegistryView.Register("otp.OTPPreferencesView",
	lago.GetPageView("otp.OTPPreferencesForm").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("users.role", p_users.RoleAuthorizationMiddleware{Roles: []string{"superuser"}}).
		WithMiddleware("otp.preferences", views.MiddlewareSingleton[OTPPreferences]{
			SuccessURL: lago.RoutePath("otp.OTPPreferencesRoute", nil),
		}))
```

### 8. JSON Import View

```go
lago.RegistryView.Register("nirmancampus_website.ImportantLinksImportView",
	lago.GetPageView("nirmancampus_website.ImportantLinksImportForm").
		WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
		WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware).
		WithMiddleware("important_links_admin.import", views.MiddlewareJsonImport[ImportantLink]{
			FileField:  "ImportFile",
			SuccessURL: lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
		}))
```

### 9. Custom GET/POST Route

```go
lago.RegistryView.Register("users.LoginView",
	lago.GetPageView("users.LoginPage").
		WithMiddleware("users.login", views.MethodMiddleware{
			Method:  http.MethodPost,
			Handler: loginHandler,
		}))
```

If the same page needs both GET and POST behavior, add two `MethodMiddleware` entries in order.

### 10. Redirect Route

Pure redirect route:

```go
lago.RegistryView.Register("base.HomeView",
	lago.RedirectView(lago.RoutePath("users.LoginRoute", nil)))
```

Redirect inside a custom handler:

```go
url, _ := getters.IfOr(lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.Any(getters.Static(user.ID)),
}), r.Context(), "")
lago.Redirect(w, r, url)
```

Rules:

- Do not use `lago.NewRedirectView`.
- If the whole route is just a redirect, prefer `lago.RedirectView(...)`.
- If only one branch redirects, call `lago.Redirect(...)` in the handler/middleware.

### 11. Path Params Helper

Old:

```go
views.PathMiddleware("form_id", "id")
```

New:

```go
views.PathMiddleware{Names: []string{"form_id", "id"}}
```

This only copies raw path values into `$path`. It does not load a DB record.

## Query Patcher Conventions

Query patchers are typed now.

Use a struct or exported zero-value variable:

```go
type courseScopeByRole struct{}

func (courseScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Course]) gorm.ChainInterface[Course] {
	return query.Where("created_by_id = ?", r.Context().Value("$user").(User).ID)
}

var CourseScopeByRole views.QueryPatcher[Course] = courseScopeByRole{}
```

Do not return raw functions as `views.QueryPatcher`.

If the behavior is simple, use the built-ins:

```go
views.QueryPatcherPreload[Student]{Field: "User"}
views.QueryPatcherOrderBy[Student]{Order: "name ASC"}
```

## Form Patcher Conventions

Form patchers also use an interface now.

```go
type createdByFormPatcher struct{}

func (createdByFormPatcher) Patch(_ views.View, r *http.Request, values map[string]any, errs map[string]error) (map[string]any, map[string]error) {
	values["CreatedByID"] = r.Context().Value("$user").(p_users.User).ID
	return values, errs
}
```

Attach them inside the owning middleware:

```go
views.MiddlewareCreate[Thing]{
	FormPatchers: views.FormPatchers{
		{Key: "thing.created_by", Value: createdByFormPatcher{}},
	},
}
```

Do not call `WithFormPatcher`.

## Replacing `WithRenderMiddleware`

There is no special render middleware anymore.

Just write a normal `views.Middleware`:

```go
type attachExtraContext struct{}

func (attachExtraContext) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "extra", "value")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

Then add or patch it like any other middleware:

```go
lago.RegistryView.Patch("students.DetailView", func(v *views.View) *views.View {
	return v.InsertMiddlewareAfter("students.detail", "myplugin.extra", attachExtraContext{})
})
```

Use this pattern when a plugin needs to show related data on another plugin's detail page.

## Replacing `Handlers` / `WithMethod`

Do not build a custom `View{Handlers: ...}`.

Instead:

- start from `lago.GetPageView("...")`
- add normal middlewares
- add one or more `views.MethodMiddleware`

Example:

```go
lago.GetPageView("users.SelfChangePasswordForm").
	WithMiddleware("users.auth", AuthenticationMiddleware{}).
	WithMiddleware("users.self_detail", authenticatedUserDetailMiddleware{}).
	WithMiddleware("users.self_change_password", views.MethodMiddleware{
		Method:  http.MethodPost,
		Handler: selfChangePasswordHandler,
	})
```

## Error Handling Pattern In Custom Handlers

Old helpers are gone:

- no `v.HasErrors(...)`
- no `v.RenderWithErrors(...)`

Use this pattern:

```go
values, fieldErrors, err := v.ParseForm(w, r)
if err != nil {
	ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
	v.RenderPage(w, r.WithContext(ctx))
	return
}

if len(fieldErrors) != 0 {
	ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
	v.RenderPage(w, r.WithContext(ctx))
	return
}
```

If you add more field errors later, re-render the same way.

## Special Cases And Gotchas

- `MiddlewareDelete[T]` uses `POST`. Ignore any old comment or old mental model that says DELETE.

- `MiddlewareDetail[T]` stores the loaded record under whatever `Key` you give it. `MiddlewareUpdate[T]`, `MiddlewareDelete[T]`, and custom context middleware must use the same key.

- `MiddlewareList[T]` automatically stores parsed query/filter state in `$get`.

- `MiddlewareCreate[T]` stores the new record ID in `$id`.

- `MiddlewareJsonImport[T]` stores the imported count in `$count`.

- `MiddlewareSingleton[T]` loads/creates the singleton on GET and places it in `$in`.

- If you need extra context for a detail page, put that middleware after the detail middleware, not before it.

- If you need role/auth before a create/update/delete, put those middlewares before the CRUD middleware.

- If your page uses many-to-many inputs, do not manually persist join tables unless the flow is truly custom. `MiddlewareCreate`, `MiddlewareUpdate`, and `MiddlewareSingleton` already handle `components.AssociationIDs`.

- If you are patching another plugin's view, stable middleware keys are more important than clever code. Name them once, then patch against those names forever.

- `views.PathMiddleware` is now a struct literal, not a constructor call.

- `p_users.AuthenticationMiddleware`, `OptionalAuthMiddleware`, and `RoleAuthorizationMiddleware` are struct-based middleware now. Use struct literals like `AuthenticationMiddleware{}` and `RoleAuthorizationMiddleware{Roles: []string{"admin"}}`.

- Global HTTP middleware also uses structs now, for example `views.AttachRequestMiddleware{}` and `lago.DBMiddleware{DB: db}`.

## Recommended Migration Order

When rewriting one old view file:

1. Replace the outer factory with `lago.GetPageView("...")`.
2. Add auth/role/path middleware first.
3. Add `MiddlewareList` or `MiddlewareDetail`.
4. Add `MiddlewareCreate`, `MiddlewareUpdate`, `MiddlewareDelete`, `MiddlewareSingleton`, or `MiddlewareJsonImport` if needed.
5. Convert old query patchers to typed `views.QueryPatcher[T]`.
6. Convert old form patchers to `views.FormPatcher`.
7. Replace `WithMethod`/`Handlers` with `views.MethodMiddleware`.
8. Replace `WithRenderMiddleware` with a normal middleware.
9. Replace redirects with `lago.RedirectView` or `lago.Redirect`.
10. Inline error rendering with `views.ContextWithErrorsAndValues`.

## Final "Do Not Do This" List

- Do not add `views.ListView`, `views.DetailView`, `views.CreateView`, `views.UpdateView`, `views.DeleteView`, `views.SingletonView`, or `views.JsonImport`.
- Do not add `WithMethod`, `WithQueryPatcher`, `WithFormPatcher`, or `WithRenderMiddleware`.
- Do not add `Handlers`.
- Do not add `lago.NewRedirectView`.
- Do not add raw `func(http.Handler) http.Handler` middleware.
- Do not reintroduce `HasErrors` or `RenderWithErrors`.
- Do not make delete routes depend on HTTP DELETE.

If you follow the rules above, the migration is usually just "turn special helpers into ordered middlewares".
