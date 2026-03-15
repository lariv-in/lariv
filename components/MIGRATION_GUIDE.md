# Components Migration Guide

This guide explains how to update a component so it works with the new `PageInterface` and `ParentInterface` requirements.

This is written for beginners. If you are unsure what a method should do, copy one of the working patterns from the existing components and adjust it carefully.

## What changed

Every component that should behave like a page now needs:

- `GetKey() string`
- `GetRoles() []string`

Every component that contains child components and should support tree editing now also needs:

- `GetChildren() []PageInterface`
- `SetChildren([]PageInterface)`

If a component has children, it should implement `ParentInterface`.
If it has no children, it only needs to satisfy `PageInterface`.

## The two interfaces

`PageInterface` requires:

```go
type PageInterface interface {
	Build(context.Context) gomponents.Node
	GetKey() string
	GetRoles() []string
}
```

`ParentInterface` requires:

```go
type ParentInterface interface {
	PageInterface
	GetChildren() []PageInterface
	SetChildren([]PageInterface)
}
```

## The `Page` struct

Most components already embed:

```go
type Page struct {
	Key   string
	Roles []string
}
```

If your component embeds `Page`, then the normal implementations are:

```go
func (e MyComponent) GetKey() string {
	return e.Key
}

func (e MyComponent) GetRoles() []string {
	return e.Roles
}
```

## Rule 1: every component needs `GetKey`

If the component embeds `Page`, use:

```go
func (e MyComponent) GetKey() string {
	return e.Key
}
```

Why:

- `ReplaceChild`
- `InsertChildBefore`
- `InsertChildAfter`

all identify the target child by `GetKey()`.

If `GetKey()` returns the wrong thing, tree editing will not find your component.

## Rule 2: every component needs `GetRoles`

If the component embeds `Page`, use:

```go
func (e MyComponent) GetRoles() []string {
	return e.Roles
}
```

Why:

- role-based rendering depends on this
- if this is missing, the component does not satisfy `PageInterface`

## Rule 3: only parent components need `GetChildren` and `SetChildren`

A "parent component" is anything that stores child components inside itself.

Examples:

- `Children []PageInterface`
- `Columns []TableColumn` where each column has `Children`
- `Items []AccordionItem` where each item has `Children`

If a component does not own child pages, do not add `SetChildren`.

Examples of non-parent components:

- `InputTime`
- `InputCheckbox`
- simple text/field/input components

These should only implement:

- `Build`
- `GetKey`
- `GetRoles`

## The easy case: direct child slice

If your component looks like this:

```go
type ContainerRow struct {
	Page
	Children []PageInterface
	Classes  string
}
```

then use this pattern:

```go
func (e ContainerRow) GetChildren() []PageInterface {
	return e.Children
}

func (e *ContainerRow) SetChildren(children []PageInterface) {
	e.Children = children
}
```

This is the simplest case.

Use it when:

- the component has exactly one child slice
- `GetChildren()` returns that exact slice
- there is no flattening or regrouping logic

## The important caveat: `SetChildren` must usually be a pointer receiver

This is very important.

Use:

```go
func (e *MyComponent) SetChildren(children []PageInterface) {
	e.Children = children
}
```

Do not use:

```go
func (e MyComponent) SetChildren(children []PageInterface) {
	e.Children = children
}
```

Why:

- a value receiver modifies a copy
- a pointer receiver modifies the real struct

If you use a value receiver for `SetChildren`, your replace/insert logic may appear to run but nothing will actually change in the original component tree.

## The harder case: flattened children

Some components do not store children in one direct slice.

Examples:

- `Accordion` stores children in `Items[i].Children`
- table content components store children in `Columns[i].Children`

In these cases, `GetChildren()` often flattens multiple internal slices into one `[]PageInterface`.

Example idea:

```go
func (e TableListContent) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}
```

If `GetChildren()` flattens data like this, then `SetChildren()` must reverse that process.

## How to write `SetChildren` for flattened structures

You must put the flat slice back into the original internal layout.

The current pattern used in this codebase is:

1. Walk the existing internal groups in order.
2. Look at each group's current child count.
3. Take that many items from the incoming flat slice.
4. Assign them back to that group.
5. If there are extra children left over, append them to the last group.

Example pattern:

```go
func (e *MyGroupedComponent) SetChildren(children []PageInterface) {
	offset := 0
	for i := range e.Groups {
		n := len(e.Groups[i].Children)
		end := offset + n
		if end > len(children) {
			end = len(children)
		}
		e.Groups[i].Children = children[offset:end]
		offset = end
		if offset >= len(children) {
			return
		}
	}
	if offset < len(children) && len(e.Groups) > 0 {
		e.Groups[len(e.Groups)-1].Children = append(e.Groups[len(e.Groups)-1].Children, children[offset:]...)
	}
}
```

## Why this grouped `SetChildren` works

Because `GetChildren()` returns children in a predictable order, `SetChildren()` can rebuild the grouped layout using the old group sizes.

This gives a stable rule:

- first chunk goes into the first group
- second chunk goes into the second group
- and so on

This is good enough for tree editing operations like replace and insert.

## Caveat: `SetChildren(GetChildren())` should preserve structure

This is a useful mental test.

If you call:

```go
flat := component.GetChildren()
component.SetChildren(flat)
```

the component should still have the same structure afterward.

If that round trip breaks the grouping, your `SetChildren` implementation is wrong.

## Caveat: extra children need a rule

If an insert operation adds one extra child, where does it go?

For grouped components, you must choose a rule.

The current rule in this codebase is:

- assign children back using existing group sizes
- if anything is left over, append it to the last group

This is not the only possible rule, but it is simple and predictable.

Do not invent a different rule for each component unless there is a very good reason.

Consistency matters more than cleverness here.

## Caveat: fewer children also need to be handled safely

If the new flat slice is shorter than before:

- do not panic
- do not index past the end
- cap the end index at `len(children)`

That is why this line matters:

```go
if end > len(children) {
	end = len(children)
}
```

Without that check, your code can crash.

## Caveat: components with no groups

If your grouped component has zero groups, be careful.

This means code like this is unsafe unless guarded:

```go
e.Groups[len(e.Groups)-1]
```

Always check:

```go
if offset < len(children) && len(e.Groups) > 0 {
	// safe
}
```

## Caveat: `GetChildren` order and `SetChildren` order must match

This is another very important rule.

If `GetChildren()` returns:

1. group A children
2. then group B children
3. then group C children

then `SetChildren()` must rebuild in that same order.

If the orders do not match, replace/insert operations will modify the wrong group.

## When not to add `SetChildren`

Do not add `SetChildren` just because the component mentions another page somewhere.

Only add it if the component truly owns editable child pages.

Examples that usually do not need `SetChildren`:

- a component with `Title PageInterface` only
- a component with one nested page used as metadata, decoration, or label
- components that are leaves in the tree

If the component is not intended to act like a parent in the traversal/editing system, leave it as a plain `PageInterface` implementation.

## Migration checklist for a normal leaf component

If the component has no child pages:

1. Make sure it embeds `Page`, or has equivalent `Key` and `Roles` storage.
2. Add `GetKey() string`.
3. Add `GetRoles() []string`.
4. Do not add `GetChildren()` or `SetChildren()`.

Example:

```go
type InputTime struct {
	Page
	Label    string
	Name     string
	Getter   getters.Getter
	Required bool
	Classes  string
}

func (e InputTime) GetKey() string {
	return e.Key
}

func (e InputTime) GetRoles() []string {
	return e.Roles
}
```

## Migration checklist for a simple parent component

If the component has `Children []PageInterface`:

1. Add `GetKey() string`.
2. Add `GetRoles() []string`.
3. Add `GetChildren() []PageInterface`.
4. Add `SetChildren([]PageInterface)` with a pointer receiver.

Example:

```go
type ContainerColumn struct {
	Page
	Children []PageInterface
	Classes  string
}

func (e ContainerColumn) GetKey() string {
	return e.Key
}

func (e ContainerColumn) GetRoles() []string {
	return e.Roles
}

func (e ContainerColumn) GetChildren() []PageInterface {
	return e.Children
}

func (e *ContainerColumn) SetChildren(children []PageInterface) {
	e.Children = children
}
```

## Migration checklist for a grouped parent component

If the component stores children indirectly:

1. Add `GetKey() string`.
2. Add `GetRoles() []string`.
3. Write `GetChildren()` so it flattens all internal child slices in a stable order.
4. Write `SetChildren()` so it reconstructs those slices in the same order.
5. Use a pointer receiver for `SetChildren()`.
6. Handle both too-few and too-many incoming children safely.

Examples:

- `Accordion`
- `TableGridContent`
- `TableListContent`

## Very common beginner mistakes

### Mistake 1: returning an empty string from `GetKey`

Bad:

```go
func (e MyComponent) GetKey() string {
	return ""
}
```

Why this is bad:

- the component can never be found by key
- replace/insert operations will not target it correctly

Use:

```go
func (e MyComponent) GetKey() string {
	return e.Key
}
```

### Mistake 2: forgetting `GetRoles`

If `PageInterface` requires `GetRoles`, every page component must have it.

Do not assume embedding `Page` automatically satisfies the interface. Promoted fields are not the same as methods.

You still need the method.

### Mistake 3: using a value receiver for `SetChildren`

Bad:

```go
func (e MyComponent) SetChildren(children []PageInterface) {
	e.Children = children
}
```

Why this is bad:

- `e` is a copy
- the caller's component does not change

Use a pointer receiver.

### Mistake 4: flattening in one order and rebuilding in another

If `GetChildren()` and `SetChildren()` disagree about ordering, your tree editing functions will produce weird bugs that are hard to track down.

Keep the order identical.

### Mistake 5: adding parent methods to leaf components

This creates fake parent behavior and makes the code harder to reason about.

If there are no true editable children, do not implement `ParentInterface`.

## Sanity tests you can do mentally

For a leaf component:

- does it have `Build`, `GetKey`, and `GetRoles`?

For a simple parent:

- does `GetChildren()` return the exact child slice?
- does `SetChildren()` replace that exact child slice?

For a grouped parent:

- does `GetChildren()` flatten in a stable order?
- does `SetChildren(GetChildren())` preserve the structure?
- does `SetChildren()` avoid panics when the input slice is too short?
- does it have a clear rule for extra children?

## Short version

If you just want the shortest possible rule:

- every component gets `GetKey()` and `GetRoles()`
- every real parent component gets `GetChildren()` and `SetChildren()`
- `GetKey()` should normally return `e.Key`
- `GetRoles()` should normally return `e.Roles`
- `SetChildren()` should almost always use a pointer receiver
- if `GetChildren()` flattens nested groups, `SetChildren()` must rebuild them in the same order

## Final advice

When in doubt:

1. decide whether the component is a leaf or a parent
2. copy the closest working example
3. keep the implementation boring and predictable
4. do not be clever with `SetChildren`

Boring code is good here. Tree mutation code should be easy to trust.
