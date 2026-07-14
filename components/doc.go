// Package components provides reusable and interactive UI layout components, fields, inputs, and page scaffolds
// for the Lago application framework, constructed using maragu.dev/gomponents.
//
// The package revolves around several core interfaces that define components, nested structures, document shells, and forms.
//
// # PageInterface
//
// Representing the base component type. It is used to build single fields (e.g. FieldText), form inputs (e.g. InputText), small buttons, or custom page displays.
//
//	type CustomBadge struct {
//		components.Page
//		Text string
//	}
//
//	func (b CustomBadge) Build(ctx context.Context) gomponents.Node {
//		return html.Span(html.Class("badge"), gomponents.Text(b.Text))
//	}
//
//	func (b CustomBadge) GetKey() string     { return b.Key }
//	func (b CustomBadge) GetRoles() []string { return b.Roles }
//
// # ParentInterface & MutableParentInterface
//
// Interfaces representing components that contain nested sub-components. They are used to create layout containers (e.g. LayoutCard), menu panels, or page scaffolds containing customizable content slots.
//
//	func ApplyAdminActions(parent components.MutableParentInterface) {
//		components.InsertChildAfter(parent, "save-button-key", func(existing *components.ButtonSubmit) components.PageInterface {
//			return &components.ButtonLink{
//				Label: getters.Static("Cancel"),
//				Link:  lago.RoutePath("admin.Dashboard", nil),
//			}
//		})
//	}
//
// # Shell
//
// Represents the global root HTML document template scaffolding. It is used to create unified layout structures containing standard metadata headers, body styles, alerts, and navigation links.
//
//	var MainShell components.Shell = &components.ShellScaffold{
//		Sidebar:  []components.PageInterface{AppMenu},
//		Children: []components.PageInterface{DashboardContent},
//	}
//
// # FormInterface
//
// Defines components representing web forms that wrap parameters inputs, handle validations, and parse submitted values. They are used to create database editing wizards, credentials forms, or query filters.
//
//	var UserForm components.FormInterface = &components.FormComponent[User]{
//		ActionURL: lago.RoutePath("users.Create", nil),
//		Children:  []components.PageInterface{EmailInput, RoleSelector},
//	}
package components
