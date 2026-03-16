# View Migration Guide

This guide explains how to migrate a `views.go` file from the old untyped getter format to the new typed generic getter format.

It is written for humans, not compiler enthusiasts.

## Goal

The goal of the migration is:

- stop passing loosely typed getters around
- make route params explicit
- catch type mistakes earlier
- reduce runtime/reflection surprises

Old code often used:

```go
map[string]getters.Getter{"id": getters.GetterKey("$id")}
```

New code should use:

```go
map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
}
```

## The Big Idea

There are 3 important rules:

1. `GetterKey` must now be typed.
2. Route argument maps must now be `map[string]getters.Getter[any]`.
3. If a typed getter is being passed into a place that expects `Getter[any]`, wrap it with `getters.GetterAny(...)`.

## What Usually Changes

In a typical `views.go` file, you will mostly touch:

- `lago.GetterRoutePath(...)`
- `lago.NewRedirectView(...)`
- any `map[string]getters.Getter{...}`
- any `getters.GetterKey(...)` inside route args

Most middleware and view registration structure stays the same.

## Step 1: Find Old Route Arg Maps

Search for patterns like:

```go
map[string]getters.Getter{
```

These old maps need to become:

```go
map[string]getters.Getter[any]{
```

## Step 2: Type Every `GetterKey`

Old code:

```go
getters.GetterKey("$id")
getters.GetterKey("user.ID")
getters.GetterKey("role.ID")
```

New code:

```go
getters.GetterKey[uint]("$id")
getters.GetterKey[uint]("user.ID")
getters.GetterKey[uint]("role.ID")
```

Pick the real type from context.

Common examples:

- database IDs: `uint`
- strings like names/emails/path params stored as strings: `string`
- flags: `bool`

## Step 3: Wrap Route Args With `GetterAny`

`lago.GetterRoutePath(...)` expects:

```go
map[string]getters.Getter[any]
```

But `getters.GetterKey[uint]("$id")` returns:

```go
getters.Getter[uint]
```

Those are not the same type.

So wrap the typed getter:

```go
getters.GetterAny(getters.GetterKey[uint]("$id"))
```

## Step 4: Migrate `GetterRoutePath(...)`

### Old

```go
views.CreateView[User](
	lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter{
		"id": getters.GetterKey("$id"),
	}),
)
```

### New

```go
views.CreateView[User](
	lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
	}),
)
```

That is the main migration pattern.

## Step 5: Migrate `NewRedirectView(...)`

If your view file has redirects with args, do the same thing there.

### Old

```go
lago.NewRedirectView("users.DetailRoute", map[string]getters.Getter{
	"id": getters.GetterKey("$id"),
})
```

### New

```go
lago.NewRedirectView("users.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
})
```

If `NewRedirectView` still accepts `map[string]getters.Getter`, you must update its signature too.

## Step 6: Use the Correct Type for `$id`

Do not blindly use `uint` everywhere.

Look at what you actually stored in context.

### Case A: `$id` is a `uint`

Example:

```go
ctx := context.WithValue(r.Context(), "$id", user.ID)
```

Then use:

```go
getters.GetterAny(getters.GetterKey[uint]("$id"))
```

### Case B: `$id` is a `string`

Example:

```go
ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", user.ID))
```

Then use:

```go
getters.GetterAny(getters.GetterKey[string]("$id"))
```

This matters. If you store a string and read it as `uint`, the getter will fail.

## Step 7: Do Not Change What Does Not Need Changing

These usually stay the same:

- `views.ListView[T](...)`
- `views.DetailView[T](...)`
- `views.UpdateView[T](...)`
- middleware wrapping
- page names like `lago.GetPageView("users.UserTable")`

The migration is mostly about getter types, not view architecture.

## Before And After Example

### Old

```go
lago.RegistryView.Register("users.CreateView",
	AuthenticationMiddleware(
		RoleAuthorizationMiddleware([]string{""})(
			views.CreateView[User](lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter{
				"id": getters.GetterKey("$id"),
			}))(
				lago.GetPageView("users.UserCreateForm")))))
```

### New

```go
lago.RegistryView.Register("users.CreateView",
	AuthenticationMiddleware(
		RoleAuthorizationMiddleware([]string{""})(
			views.CreateView[User](lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}))(
				lago.GetPageView("users.UserCreateForm")))))
```

## Special Case: RedirectView / Framework Helpers

If your view file is migrated but helper code is still old, you may hit compiler errors.

Example:

```go
cannot use map[string]getters.Getter[any] as map[string]invalid type
```

That usually means a helper still has an old signature.

For example, this old helper:

```go
type RedirectView struct {
	Args map[string]getters.Getter
}

func NewRedirectView(routeKey string, args ...map[string]getters.Getter) RedirectView
```

must become:

```go
type RedirectView struct {
	Args map[string]getters.Getter[any]
}

func NewRedirectView(routeKey string, args ...map[string]getters.Getter[any]) RedirectView
```

If one side is migrated and the other side is not, the compiler will complain.

## Common Mistakes

### Mistake 1: Forgetting `[any]` on the map

Wrong:

```go
map[string]getters.Getter{
```

Right:

```go
map[string]getters.Getter[any]{
```

### Mistake 2: Forgetting the type parameter on `GetterKey`

Wrong:

```go
getters.GetterKey("$id")
```

Right:

```go
getters.GetterKey[uint]("$id")
```

## Mistake 3: Forgetting `GetterAny`

Wrong:

```go
"id": getters.GetterKey[uint]("$id")
```

Right:

```go
"id": getters.GetterAny(getters.GetterKey[uint]("$id"))
```

## Mistake 4: Using the wrong type

Wrong:

```go
ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", user.ID))
"id": getters.GetterAny(getters.GetterKey[uint]("$id"))
```

Right:

```go
ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", user.ID))
"id": getters.GetterAny(getters.GetterKey[string]("$id"))
```

## Simple Migration Checklist

For each `views.go` file:

- replace `map[string]getters.Getter` with `map[string]getters.Getter[any]` where used for route args
- add a concrete type to every `GetterKey`
- wrap route arg getters in `getters.GetterAny(...)`
- verify `$id` and other context values use the same type you read back
- update helper/framework code if it still accepts old getter map types
- run the compiler and fix the remaining typed getter mismatches

## Fast Mental Model

Use this rule of thumb:

- reading from context: `GetterKey[RealType](...)`
- passing to route args: `GetterAny(GetterKey[RealType](...))`
- route arg maps: `map[string]getters.Getter[any]`

## One-Liner Conversion Recipe

If you see this:

```go
map[string]getters.Getter{"id": getters.GetterKey("$id")}
```

turn it into this:

```go
map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}
```

Then double-check whether `$id` is really a `uint` or actually a `string`.
