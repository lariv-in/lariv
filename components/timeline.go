package components

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
)

// Timeline represents an interactive vertical event timeline listing component.
// It displays a chronological list of activities, messages, or changes mapped from dynamic Getter payloads,
// featuring item onClick navigations, optional filter menus, create action triggers, and vertical connector lines.
//
// Use Cases:
//   - Showing transactional histories, audit trials, message feeds, or workflow timeline charts.
//
// Example:
//
//	&components.Timeline[Event]{
//	    Title:     "Activity Log",
//	    Data:      eventDataGetter,
//	    OnClick:   getters.RowAttrNavigate(lariv.RoutePath("events.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
//	    CreateUrl: lariv.RoutePath("events.CreateRoute", nil),
//	}
type Timeline[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// UID represents the unique HTML element wrapper ID (defaults to "timeline-container").
	UID string
	// Title represents the heading label text displayed above the timeline.
	Title string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Data represents the dynamic Getter retrieving the paginated ObjectList timeline payload.
	Data getters.Getter[ObjectList[T]]
	// OnClick represents the dynamic getter returning target navigation URLs on item clicks.
	OnClick getters.Getter[string]
	// FilterComponent represents an optional filter dropdown menu component.
	FilterComponent PageInterface
	// CreateUrl is the dynamic function retrieving the creation button target path.
	CreateUrl getters.Getter[string]
	// Children represents the slice of sub-components rendering inside individual timeline card nodes.
	Children []PageInterface
}

type timelineItem struct {
	Href    string
	Content template.HTML
}

// Build compiles the Timeline component into chronological cards lists, side guidelines and pagination panels.
func (e Timeline[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	var data []T
	if e.Data != nil {
		list, err := e.Data(ctx)
		if err != nil {
			slog.Error("Timeline Data getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		data = list.Items
	}

	uid := e.UID
	if uid == "" {
		uid = "timeline-container"
	}

	var createHTML template.HTML
	if e.CreateUrl != nil {
		createURL, err := e.CreateUrl(ctx)
		if err == nil && createURL != "" {
			h, err := RenderHTML(ButtonLink{
				Link:    getters.Static(createURL),
				Icon:    "plus",
				Classes: "btn-square btn-outline btn-sm",
			}, cat, ctx)
			if err != nil {
				return err
			}
			createHTML = h
		}
	}

	var filterHTML template.HTML
	if e.FilterComponent != nil {
		icon, err := RenderHTML(Icon{Name: "funnel"}, cat, ctx)
		if err != nil {
			return err
		}
		panel, err := RenderHTML(e.FilterComponent, cat, ctx)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := Execute(&buf, "timeline_filter", struct {
			Icon  template.HTML
			Panel template.HTML
		}{Icon: icon, Panel: panel}); err != nil {
			return err
		}
		filterHTML = template.HTML(buf.String())
	}

	var actions template.HTML
	if filterHTML != "" || createHTML != "" {
		actions = filterHTML + createHTML
	}
	showHeader := e.Title != "" || filterHTML != "" || createHTML != ""

	items := make([]timelineItem, 0, len(data))
	for _, item := range data {
		rowMap := getters.MapFromStruct(any(item))
		itemCtx := context.WithValue(ctx, "$row", rowMap)

		content, err := RenderChildren(cat, itemCtx, e.Children)
		if err != nil {
			return err
		}

		ti := timelineItem{Content: content}
		if e.OnClick != nil {
			url, err := e.OnClick(itemCtx)
			if err != nil {
				slog.Error("Timeline OnClick getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
			}
			ti.Href = url
		}
		items = append(items, ti)
	}

	pagination, err := RenderHTML(TablePagination[T]{Data: e.Data}, cat, ctx)
	if err != nil {
		return err
	}

	return Execute(w, "timeline", struct {
		UID        string
		Classes    string
		ShowHeader bool
		Title      string
		Actions    template.HTML
		ShowLine   bool
		Empty      bool
		Items      []timelineItem
		Pagination template.HTML
	}{
		UID:        uid,
		Classes:    e.Classes,
		ShowHeader: showHeader,
		Title:      e.Title,
		Actions:    actions,
		ShowLine:   len(data) > 0,
		Empty:      len(data) == 0,
		Items:      items,
		Pagination: pagination,
	})
}

// GetKey returns the unique key identifier for this Timeline.
func (e Timeline[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this Timeline.
func (e Timeline[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e Timeline[T]) GetChildren() []PageInterface {
	var children []PageInterface
	if e.FilterComponent != nil {
		children = append(children, e.FilterComponent)
	}
	children = append(children, e.Children...)
	return children
}

// SetChildren replaces the slice of nested sub-components.
func (e *Timeline[T]) SetChildren(children []PageInterface) {
	offset := 0
	if e.FilterComponent != nil && len(children) > 0 {
		e.FilterComponent = children[0]
		offset = 1
	}
	e.Children = children[offset:]
}
