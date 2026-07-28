package components

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lariv/getters"
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

type tablePaginationPage struct {
	Ellipsis bool
	URL      string
	PushURL  string
	Classes  string
	Label    string
}

// Build compiles the TablePagination component into a centered list pagination buttons row.
func (e TablePagination[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
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

	var pages []tablePaginationPage

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
			tablePaginationPage{Ellipsis: true})
	}

	for p := startPage; p <= endPage; p++ {
		pages = append(pages, e.pageButton(req, p, uint(p) == number))
	}

	if endPage < np {
		pages = append(pages, tablePaginationPage{Ellipsis: true},
			e.pageButton(req, np, number == numPages))
	}

	return Execute(w, "table_pagination", struct {
		Pages []tablePaginationPage
	}{Pages: pages})
}

// pageButton constructs a single navigation button linked to page index p.
func (e TablePagination[T]) pageButton(req *http.Request, p int, active bool) tablePaginationPage {
	u := *req.URL
	q := u.Query()
	q.Set("page", strconv.Itoa(p))
	u.RawQuery = q.Encode()

	classes := "join-item btn btn-sm"
	if active {
		classes += " btn-active"
	}

	pushURL := "true"
	if strings.Contains(req.URL.Path, "select") {
		pushURL = "false"
	}

	return tablePaginationPage{
		URL:     u.String(),
		PushURL: pushURL,
		Classes: classes,
		Label:   strconv.Itoa(p),
	}
}
