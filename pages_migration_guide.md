# Pages Migration Guide

This guide explains how to migrate a `pages.go` file from the old loose getter style to the new typed generic getter style.

It is written for humans.

The goal is to make page definitions boring, predictable, and compiler-friendly.

## Why This Migration Exists

Old page files often used:

```go
getters.GetterKey("course.Name")
map[string]getters.Getter{"id": getters.GetterKey("course.ID")}
components.DataTable{ ... }
```

That style is weakly typed.

It causes problems like:

- the compiler cannot infer the real type
- route args become loosely typed blobs
- table data shape is unclear
- errors show up later than they should

The new style makes types explicit:

```go
getters.GetterKey[string]("course.Name")
map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("course.ID"))}
components.DataTable[Course]{ ... }
```

## The Main Idea

A migrated `pages.go` file does 4 things:

1. every `GetterKey(...)` gets a real type
2. every route arg map becomes `map[string]getters.Getter[any]`
3. route args use `getters.GetterAny(...)`
4. generic components like `DataTable` get a concrete type parameter

That is most of the work.

## Quick Rule Of Thumb

If you see this:

```go
getters.GetterKey("something")
```

you now need this:

```go
getters.GetterKey[RealType]("something")
```

If you see this:

```go
map[string]getters.Getter{"id": getters.GetterKey("something.ID")}
```

you now need this:

```go
map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("something.ID")),
}
```

## What Usually Changes In A `pages.go`

Most page files contain these sections:

- menus
- filters
- forms
- tables
- detail pages
- delete confirmation
- selection tables

Each section has a common migration pattern.

## Step 1: Migrate Menu Titles And Menu Links

### Old

```go
Title: getters.GetterFormat("Course: %s", getters.GetterKey("course.Name"))
```

### New

```go
Title: getters.GetterFormat("Course: %s", getters.GetterAny(getters.GetterKey[string]("course.Name")))
```

Why?

- `GetterFormat` expects `Getter[any]`
- `GetterKey[string](...)` returns `Getter[string]`
- so you wrap it with `GetterAny(...)`

### Old route link

```go
Url: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{
	"id": getters.GetterKey("course.ID"),
})
```

### New route link

```go
Url: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("course.ID")),
})
```

## Step 2: Migrate Filters

Filter forms are usually easy.

Text fields become `GetterKey[string](...)`.

Boolean/ternary fields become `GetterKey[bool](...)`.

### Old

```go
components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey("$get.Name")}
components.InputTernary{Getter: getters.GetterKey("$get.IsActive")}
```

### New

```go
components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")}
components.InputTernary{Getter: getters.GetterKey[bool]("$get.IsActive")}
```

## Step 3: Migrate Form Fields

Every input getter should be typed.

Examples:

- text input: `GetterKey[string](...)`
- checkbox/ternary: `GetterKey[bool](...)`
- foreign key id: often `GetterKey[uint](...)`

### Old

```go
components.InputText{Label: "Course Name", Name: "Name", Getter: getters.GetterKey("$in.Name")}
components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey("$in.Code")}
components.InputTernary{Getter: getters.GetterKey("$in.IsActive")}
```

### New

```go
components.InputText{Label: "Course Name", Name: "Name", Getter: getters.GetterKey[string]("$in.Name")}
components.InputText{Label: "Code", Name: "Code", Getter: getters.GetterKey[string]("$in.Code")}
components.InputTernary{Getter: getters.GetterKey[bool]("$in.IsActive")}
```

## Step 4: Add `ContainerError` Around Inputs

This is not strictly about getters, but it is part of the new page format.

The newer page style wraps inputs in `components.ContainerError`.

### Old

```go
components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$in.Name")}
```

### New

```go
components.ContainerError{
	Error: getters.GetterKey[error]("$error.Name"),
	Children: []components.PageInterface{
		components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$in.Name")},
	},
}
```

This gives you:

- field-level validation errors
- consistent form rendering
- the same structure as migrated pages like `p_users`

## Step 5: Split Form Fields From Form Actions

Old pages often put the submit button inside the field helper.

### Old

```go
func courseFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			components.InputText{...},
			components.InputTextarea{...},
			components.ButtonSubmit{Label: "Save Course"},
		},
	}
}
```

### New

Keep the helper for fields only:

```go
func courseFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			components.ContainerError{...},
			components.ContainerError{...},
		},
	}
}
```

Then add actions in the form itself:

```go
ChildrenAction: []components.PageInterface{
	components.ButtonSubmit{Label: "Save Course"},
}
```

Why this is better:

- fields are reusable
- actions are clearly separated
- form layout is more consistent

## Step 6: Migrate FormComponent Getters

If the form loads an existing model, type the model getter too.

### Old

```go
Getter: getters.GetterKey("course")
```

### New

```go
Getter: getters.GetterKey[Course]("course")
```

This is especially common in update forms and detail pages.

## Step 7: Migrate Form URLs

If the form URL has route params, migrate the route map.

### Old

```go
Url: lago.GetterRoutePath("courses.UpdateRoute", map[string]getters.Getter{
	"id": getters.GetterKey("$in.ID"),
})
```

### New

```go
Url: lago.GetterRoutePath("courses.UpdateRoute", map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
})
```

## Step 8: Migrate DataTable

This is the biggest visible page change.

If `DataTable` is now generic, you must instantiate it with the row type.

### Old

```go
components.DataTable{
	Data: getters.GetterKey("courses"),
}
```

### New

```go
components.DataTable[Course]{
	Data: getters.GetterKey[components.ObjectList[Course]]("courses"),
}
```

That is the new standard pattern.

## Step 9: Migrate Table Click Routes

### Old

```go
OnClick: getters.GetterNavigateGetter(
	lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{
		"id": getters.GetterKey("$row.ID"),
	}),
)
```

### New

```go
OnClick: getters.GetterNavigateGetter(
	lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
	}),
)
```

## Step 10: Migrate Table Column Getters

All cell field getters should be typed.

### Old

```go
components.FieldText{Getter: getters.GetterKey("$row.Name")}
components.FieldCheckbox{Getter: getters.GetterKey("$row.IsActive")}
```

### New

```go
components.FieldText{Getter: getters.GetterKey[string]("$row.Name")}
components.FieldCheckbox{Getter: getters.GetterKey[bool]("$row.IsActive")}
```

## Step 11: Migrate Detail Pages

Detail pages usually need:

- typed page getter for the model
- typed field getters for all displayed values

### Old

```go
components.Detail{
	Getter: getters.GetterKey("course"),
	Children: []components.PageInterface{
		components.FieldTitle{Getter: getters.GetterKey("$in.Name")},
		components.FieldSubtitle{Getter: getters.GetterKey("$in.Code")},
	},
}
```

### New

```go
components.Detail{
	Getter: getters.GetterKey[Course]("course"),
	Children: []components.PageInterface{
		components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
		components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Code")},
	},
}
```

## Step 12: Migrate DeleteConfirmation URLs

### Old

```go
CancelUrl: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{
	"id": getters.GetterKey("course.ID"),
})
```

### New

```go
CancelUrl: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("course.ID")),
})
```

## Step 13: Migrate Selection Tables

Selection tables follow the same `DataTable[T]` pattern.

### Old

```go
components.DataTable{
	Data: getters.GetterKey("courses"),
	OnClick: getters.GetterSelect("course", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
}
```

### New

```go
components.DataTable[Course]{
	Data: getters.GetterKey[components.ObjectList[Course]]("courses"),
	OnClick: getters.GetterSelect("course", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
}
```

The same idea applies to `GetterMultiSelect(...)`.

## How To Pick The Right Getter Type

Use the real domain type.

Common choices:

- names, labels, descriptions, titles, emails, phone values: `string`
- booleans like `IsActive`, `IsSuperuser`: `bool`
- ids: usually `uint`
- model object: `Course`, `User`, `Role`, etc.
- paginated list: `components.ObjectList[T]`

## Common Migration Pattern Summary

### Text values

```go
getters.GetterKey[string]("$in.Name")
```

### Boolean values

```go
getters.GetterKey[bool]("$in.IsActive")
```

### Model value

```go
getters.GetterKey[Course]("course")
```

### Table list value

```go
getters.GetterKey[components.ObjectList[Course]]("courses")
```

### Route arg id

```go
map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
}
```

### Format getter input

```go
getters.GetterFormat("Course: %s", getters.GetterAny(getters.GetterKey[string]("course.Name")))
```

## Common Mistakes

## Mistake 1: Leaving `GetterKey(...)` untyped

Wrong:

```go
getters.GetterKey("$row.Name")
```

Right:

```go
getters.GetterKey[string]("$row.Name")
```

## Mistake 2: Forgetting `GetterAny(...)` in route maps

Wrong:

```go
map[string]getters.Getter[any]{
	"id": getters.GetterKey[uint]("$row.ID"),
}
```

Right:

```go
map[string]getters.Getter[any]{
	"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
}
```

## Mistake 3: Forgetting the type parameter on `DataTable`

Wrong:

```go
components.DataTable{
```

Right:

```go
components.DataTable[Course]{
```

## Mistake 4: Using the wrong list getter type

Wrong:

```go
Data: getters.GetterKey[Course]("courses")
```

Right:

```go
Data: getters.GetterKey[components.ObjectList[Course]]("courses")
```

Remember:

- `"course"` is one object
- `"courses"` is usually a paginated list from `ListView`

## Mistake 5: Leaving submit buttons inside field helper functions

Old code often worked, but the new format is cleaner if:

- field helper returns inputs only
- `ChildrenAction` contains submit buttons

## Mistake 6: Migrating table data but not table columns

If the table becomes:

```go
components.DataTable[Course]
```

but column fields still use untyped getters, you have only done half the work.

## Mistake 7: Forgetting `ContainerError`

You can technically type the getters and still skip error containers.

But the newer page format expects them for proper field-level error rendering.

## Step-By-Step Recipe

For each `pages.go` file:

1. Find all `GetterKey(...)` calls.
2. Add a concrete type to each one.
3. Find all `map[string]getters.Getter{...}` route arg maps.
4. Convert them to `map[string]getters.Getter[any]{...}`.
5. Wrap every route arg getter with `getters.GetterAny(...)`.
6. Update `GetterFormat(...)` inputs using `GetterAny(...)` when needed.
7. Add `ContainerError` around form inputs.
8. Move submit buttons from field helper functions into `ChildrenAction`.
9. Update model getters like `GetterKey("course")` to `GetterKey[Course]("course")`.
10. Update `DataTable` to `DataTable[ModelType]`.
11. Update table data getters to `GetterKey[components.ObjectList[ModelType]](...)`.
12. Update table column getters to typed getters.
13. Update selection tables the same way.
14. Run lints and fix the leftovers.

## Before And After Mini Example

### Old

```go
components.DataTable{
	Data:      getters.GetterKey("courses"),
	OnClick:   getters.GetterNavigateGetter(lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$row.ID")})),
	CreateUrl: lago.GetterRoutePath("courses.CreateRoute", nil),
	Columns: []components.TableColumn{
		{Label: "Name", Key: "Name", Children: []components.PageInterface{
			components.FieldText{Getter: getters.GetterKey("$row.Name")},
		}},
	},
}
```

### New

```go
components.DataTable[Course]{
	Data:      getters.GetterKey[components.ObjectList[Course]]("courses"),
	OnClick:   getters.GetterNavigateGetter(lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
	CreateUrl: lago.GetterRoutePath("courses.CreateRoute", nil),
	Columns: []components.TableColumn{
		{Label: "Name", Key: "Name", Children: []components.PageInterface{
			components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
		}},
	},
}
```

## Fast Mental Model

When migrating a `pages.go`, think like this:

- page fields and labels use typed getters
- route params use `Getter[any]`
- route param values come from `GetterAny(...)`
- list pages use `DataTable[Model]`
- list data comes from `components.ObjectList[Model]`

## Final Rule Of Thumb

If a page definition still contains:

- untyped `GetterKey(...)`
- `map[string]getters.Getter`
- non-generic `DataTable`

then the migration is not finished yet.
