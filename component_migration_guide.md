# Component Migration Guide

This guide explains how to migrate component files from the old loose getter style to the new typed generic getter style.

It is written for normal people.

Not for compiler goblins.

## Why We Are Doing This

Old components often used:

```go
Getter getters.Getter
```

That means the component accepted "some getter that returns something".

That sounds flexible, but it causes problems:

- the component does not know what type it is supposed to render
- the compiler cannot help much
- code often falls back to `any`, reflection, or type assertions
- mistakes show up late at runtime

The new goal is:

- each component should declare the real type it expects
- the compiler should catch mistakes early
- rendering code should stop guessing

## The Main Rule

A component should use the most specific getter type it can.

Examples:

- text input: `getters.Getter[string]`
- checkbox field: `getters.Getter[bool]`
- datetime field: `getters.Getter[time.Time]`
- route URL: `getters.Getter[string]`
- table data: `getters.Getter[ObjectList[T]]`

If a component really needs a dynamic catch-all value, only then should it use `Getter[any]`.

That should be rare.

## The Migration Mindset

When migrating a component, ask:

1. What value type is this component supposed to render?
2. Is the current code using reflection or type assertions because the getter is too loose?
3. Can I replace that with a typed getter and normal Go code?

If the answer is yes, do it.

## Rule 1: Type The `Getter` Field

### Old

```go
type InputText struct {
	Page
	Getter getters.Getter
}
```

### New

```go
type InputText struct {
	Page
	Getter getters.Getter[string]
}
```

That is the first and most important change.

## Rule 2: Stop Using `any` In Builder Functions

Old component code often looked like this:

```go
getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
	return Value(fmt.Sprintf("%s", value))
})
```

That is old-style code.

The getter was untyped, so the callback had to accept `any`.

After migration, the callback should use the real type:

```go
getters.GetterIf(e.Getter, ctx, func(_ context.Context, value string) (Node, error) {
	return Value(value), nil
})
```

Notice 2 things:

- `value` is now `string`
- the callback returns `(Node, error)` because that is what `GetterIf` expects

## Rule 3: Handle Getter Errors Properly

Do not assume `GetterIf` or `Getter(...)` always succeeds.

For render components (especially fields), do **not** silently swallow getter errors.

If a getter fails:

1. log the error
2. render the error with `ContainerError` (`ErrorContainer` in discussions)

### Good pattern

```go
var valueNode Node = Value("")
if e.Getter != nil {
	if node, err := getters.GetterIf(e.Getter, ctx, func(_ context.Context, value string) (Node, error) {
		return Value(value), nil
	}); err == nil {
		valueNode = node
	} else {
		slog.Error("InputText getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
	}
}
```

Then use `valueNode` safely.

This is much better than stuffing a maybe-broken getter call directly into HTML.

## Rule 4: Replace Type Assertions With Typed Getter Calls

Old field code often looked like this:

```go
v, ok := e.Getter(ctx).(time.Time)
if !ok {
	return Group{}
}
```

That is a red flag.

It means the component does not trust its own getter type.

### New

```go
if e.Getter == nil {
	return Group{}
}
v, err := e.Getter(ctx)
if err != nil {
	slog.Error("FieldDatetime getter failed", "error", err, "key", e.Key)
	return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
}
```

Now the component asks for a `time.Time`, so there is no need for `.(time.Time)`.

## Example 1: `InputText`

### Old shape

```go
type InputText struct {
	Page
	Label  string
	Name   string
	Getter getters.Getter
}
```

### New shape

```go
type InputText struct {
	Page
	Label  string
	Name   string
	Getter getters.Getter[string]
}
```

### Old build style

```go
Input(
	Name(e.Name),
	getters.GetterIf(e.Getter, ctx, func(ctx context.Context, value any) Node {
		return Value(fmt.Sprintf("%s", value))
	}),
)
```

### New build style

```go
var valueNode Node = Value("")
if e.Getter != nil {
	if node, err := getters.GetterIf(e.Getter, ctx, func(_ context.Context, value string) (Node, error) {
		return Value(value), nil
	}); err == nil {
		valueNode = node
	}
}

Input(
	Name(e.Name),
	valueNode,
)
```

Why this is better:

- the component says it needs a string
- the builder receives a string
- no `fmt.Sprintf("%s", value)` hack
- no `any`

## Example 2: `FieldDatetime`

### Old shape

```go
type FieldDatetime struct {
	Page
	Getter getters.Getter
}
```

### Old build style

```go
v, ok := e.Getter(ctx).(time.Time)
if !ok {
	return Group{}
}
```

### New shape

```go
type FieldDatetime struct {
	Page
	Getter getters.Getter[time.Time]
}
```

### New build style

```go
if e.Getter == nil {
	return Group{}
}
v, err := e.Getter(ctx)
if err != nil {
	return Group{}
}
```

Why this is better:

- the type is declared once
- the compiler enforces it
- runtime type guessing disappears

## Example 3: `DataTable`

This is the more advanced case.

Simple components usually only need:

- one typed getter field
- a typed build function

But table components are special because they render lists of rows.

That means the table itself should usually become generic.

### Old shape

```go
type DataTable struct {
	Data      getters.Getter
	CreateUrl getters.Getter
	OnClick   getters.Getter
}
```

This is too loose.

The component has no idea:

- what row type it is rendering
- what the data shape is
- whether `CreateUrl` is a string
- whether `OnClick` is a string expression

### New shape

```go
type DataTable[T any] struct {
	Data      getters.Getter[ObjectList[T]]
	CreateUrl getters.Getter[string]
	OnClick   getters.Getter[string]
}
```

This is much better.

Now the table knows:

- its rows are `T`
- its list shape is `ObjectList[T]`
- its click expression is a string
- its create URL is a string

## Rule 5: If The Component Renders A Collection, Consider Making The Component Generic

A table is not the only example.

Any component that works with a list or model of some type may need:

```go
type MyComponent[T any] struct { ... }
```

Typical signs:

- the component reads `.Items`
- the component loops over rows
- the component builds `$row` context
- the component uses reflection to inspect a model

If you see those signs, generics are probably the right direction.

## Rule 6: Remove Reflection When A Typed Struct Already Exists

Old table code often did something like:

```go
v := reflect.ValueOf(data)
itemsField := v.FieldByName("Items")
```

That is usually a clue that the getter type is too loose.

If the real shape is:

```go
ObjectList[T]
```

then just use:

```go
for _, row := range data.Items {
	...
}
```

Likewise for pagination:

```go
number := data.Number
numPages := data.NumPages
```

If you already know the struct type, do not use reflection to rediscover it.

## Rule 7: Tighten Related Helper Fields Too

When you migrate one component, look at all its related fields.

For example, in `DataTable`:

- `Data` should not be loose
- `CreateUrl` should be `Getter[string]`
- `OnClick` should be `Getter[string]`
- `Displays` must also use the new typed signatures

Do not migrate only one field and leave the rest in old mode if they are part of the same data flow.

## Rule 8: Migrate Helper Components Together When Needed

Some components cannot be migrated alone.

Example:

- `DataTable[T]`
- `TableListContent[T]`
- `TableGridContent[T]`
- `TablePagination[T]`

These components share the same data shape.

If only one is migrated, the others will still force you back into weak typing.

So migrate them together.

## Rule 9: Update Call Sites After Making A Component Generic

If you change:

```go
type DataTable[T any] struct
```

then all usages must now specify a type:

### Old

```go
components.DataTable{
	Data: getters.GetterKey[components.ObjectList[User]]("users"),
}
```

### New

```go
components.DataTable[User]{
	Data: getters.GetterKey[components.ObjectList[User]]("users"),
}
```

If you forget this, Go will complain:

```go
cannot use generic type components.DataTable[T any] without instantiation
```

That error is normal.

It just means you made the component generic correctly, but the callers are still old.

## Rule 10: Prefer Concrete Types Over `Getter[any]`

Use `Getter[any]` only when the component genuinely cannot know the value type.

Most of the time, components do know.

Examples:

- `InputText` knows it needs a `string`
- `FieldDatetime` knows it needs a `time.Time`
- `FieldCheckbox` knows it needs a `bool`
- `ButtonLink` knows it needs a string URL

Do not use `Getter[any]` out of habit.

## Common Migration Patterns

### Pattern A: Text-like components

Examples:

- `InputText`
- `InputEmail`
- `InputPhone`
- `FieldText`
- `FieldTitle`
- `FieldSubtitle`

Usually migrate to:

```go
Getter getters.Getter[string]
```

### Pattern B: Boolean components

Examples:

- `FieldCheckbox`
- `InputCheckbox`
- `InputTernary`

Usually migrate to:

```go
Getter getters.Getter[bool]
```

Sometimes the parser may still return `nil` for "unset", but the getter used for rendering should still be typed as narrowly as the component expects.

### Pattern C: Time/date components

Examples:

- `FieldDatetime`
- `FieldDate`
- `FieldTime`
- `InputDatetime`
- `InputDate`
- `InputTime`

Usually migrate to:

- `Getter[time.Time]` if the component works from a `time.Time`
- or `Getter[string]` if the component intentionally works with a preformatted string

Pick the real rendering contract.

### Pattern D: URL/action components

Examples:

- `ButtonLink`
- `ButtonPost`
- `ButtonModal`
- `DeleteConfirmation`

Usually migrate URL fields to:

```go
Getter getters.Getter[string]
```

### Pattern E: List/model components

Examples:

- `DataTable`
- `Timeline`
- maybe `FieldList`

Usually these should become generic:

```go
type MyComponent[T any] struct {
	Data getters.Getter[ObjectList[T]]
}
```

or another precise container type if `ObjectList[T]` is not the right one.

## Common Mistakes

## Mistake 1: Leaving `Getter` untyped

Wrong:

```go
Getter getters.Getter
```

Right:

```go
Getter getters.Getter[string]
```

or whatever the real type is.

## Mistake 2: Migrating the field but not the build logic

Wrong:

```go
Getter getters.Getter[string]

v, ok := e.Getter(ctx).(string)
```

If the getter is already typed, do this instead:

```go
v, err := e.Getter(ctx)
```

## Mistake 3: Keeping `any` in callback builders

Wrong:

```go
func(ctx context.Context, value any) Node
```

Right:

```go
func(_ context.Context, value string) (Node, error)
```

## Mistake 4: Forgetting to update related helper fields

Example:

- migrating `Data` to `Getter[ObjectList[T]]`
- but leaving `OnClick` as plain `Getter`

That is only a half migration.

## Mistake 5: Making the component generic but not its helpers

Example:

- `DataTable[T]` becomes generic
- but `TableListContent` still expects plain `Getter`

That pushes weak typing back into the system.

## Mistake 6: Forgetting to update call sites

Example:

```go
components.DataTable{
```

should become:

```go
components.DataTable[User]{
```

## Mistake 7: Using reflection when you already know the type

If you know the value is `ObjectList[T]`, `time.Time`, `string`, or `bool`, do not use reflection.

Just use the typed value directly.

## Mistake 8: Silently swallowing getter errors

Wrong:

```go
v, err := e.Getter(ctx)
if err != nil {
	return Group{}
}
```

Right:

```go
v, err := e.Getter(ctx)
if err != nil {
	slog.Error("FieldDate getter failed", "error", err, "key", e.Key)
	return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
}
```

## Step-By-Step Migration Recipe

For each component file:

1. Find every `getters.Getter` field.
2. Decide the real value type for that field.
3. Replace `getters.Getter` with `getters.Getter[RealType]`.
4. Update `Build(...)` to use typed getter calls.
5. On getter error, log and render with `ContainerError` (do not fail silently).
6. Remove type assertions like `.(time.Time)` or `.(string)` when no longer needed.
7. Remove reflection if the struct type is now known.
8. If the component renders rows or lists, consider making the component generic.
9. Migrate dependent helper components if they share the same data path.
10. Update call sites.
11. Run lints and fix any remaining mismatches.

## Quick Decision Guide

Ask this:

- Does this component render one text value?
  Use `Getter[string]`.

- Does this component render one boolean?
  Use `Getter[bool]`.

- Does this component render one timestamp?
  Use `Getter[time.Time]`.

- Does this component render a URL or click expression?
  Use `Getter[string]`.

- Does this component render a list of records?
  Make it generic and use `Getter[ObjectList[T]]` or another exact container type.

## Tiny Cheat Sheet

If you see this:

```go
Getter getters.Getter
```

replace it with the real type:

```go
Getter getters.Getter[string]
```

If you see this:

```go
v, ok := e.Getter(ctx).(time.Time)
```

replace it with:

```go
v, err := e.Getter(ctx)
```

If you see this:

```go
func(ctx context.Context, value any) Node
```

replace it with:

```go
func(_ context.Context, value string) (Node, error)
```

If you see this:

```go
type DataTable struct
```

consider:

```go
type DataTable[T any] struct
```

## Final Rule Of Thumb

A migrated component should answer this question clearly:

"What exact type do I expect from this getter?"

If the answer is still "uh, some value probably", the migration is not done yet.
