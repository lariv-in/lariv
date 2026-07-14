package components

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TablePagination represents the list pagination buttons row component for DataTable widgets.
// It compiles numeric navigation controls matching current list offset states, rendering them in a DaisyUI join group.
//
// Use Cases:
//   - Appending page navigations (e.g., [1] [2] ... [24]) underneath data grids and tables.
//
// Example:
//
//	&components.TablePagination[Invoice]{
//	    Data: invoiceDataGetter,
//	}
type TablePagination[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Data represents the dynamic Getter retrieving the paginated ObjectList payload.
	Data getters.Getter[ObjectList[T]]
}

// GetKey returns the unique key identifier for this TablePagination component.
func (e TablePagination[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this TablePagination.
func (e TablePagination[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the TablePagination component into a centered list pagination buttons row.
func (e TablePagination[T]) Build(ctx context.Context) Node {
	if e.Data == nil {
		return nil
	}
	data, err := e.Data(ctx)
	if err != nil {
		return nil
	}
	number := data.Number
	numPages := data.NumPages

	if numPages <= 1 {
		return nil
	}

	req, ok := ctx.Value("$request").(*http.Request)
	if !ok {
		return nil // Cannot reconstruct URL without request
	}

	var pages []Node

	n := int(number)
	np := int(numPages)
	windowSize := 5
	startPage := max(n-windowSize/2, 1)
	endPage := startPage + windowSize - 1
	if endPage > np {
		endPage = np
		startPage = max(endPage-windowSize+1, 1)
	}

	if startPage > 1 {
		pages = append(pages, e.pageButton(req, 1, number == 1),
			Button(Disabled(), Class("join-item btn btn-sm"), Text("...")))
	}

	for p := startPage; p <= endPage; p++ {
		pages = append(pages, e.pageButton(req, p, uint(p) == number))
	}

	if endPage < np {
		pages = append(pages, Button(Disabled(), Class("join-item btn btn-sm"), Text("...")),
			e.pageButton(req, np, number == numPages))
	}

	return Div(
		Class("flex flex-col justify-center items-center gap-2 p-4"),
		Div(
			Class("join"),
			Group(pages),
		),
	)
}

// pageButton constructs a single navigation button Node linked to page index p.
func (e TablePagination[T]) pageButton(req *http.Request, p int, active bool) Node {
	u := *req.URL
	q := u.Query()
	q.Set("page", strconv.Itoa(p))
	u.RawQuery = q.Encode()

	classes := "join-item btn btn-sm"
	if active {
		classes += " btn-active"
	}

	return A(
		Href(u.String()),
		Class(classes),
		Text(strconv.Itoa(p)),
	)
}
