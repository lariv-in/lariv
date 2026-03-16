package components

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TablePagination[T any] struct {
	Page
	Data getters.Getter[ObjectList[T]]
}

func (e TablePagination[T]) GetKey() string {
	return e.Key
}

func (e TablePagination[T]) GetRoles() []string {
	return e.Roles
}

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

	// Calculate window (similar to table_page_range in python but simpler logic)
	// Just showing a window block
	windowSize := 5
	startPage := max(number-windowSize/2, 1)
	endPage := startPage + windowSize - 1
	if endPage > numPages {
		endPage = numPages
		startPage = max(endPage-windowSize+1, 1)
	}

	if startPage > 1 {
		pages = append(pages, e.pageButton(req, 1, number == 1),
			Button(Disabled(), Class("join-item btn btn-sm"), Text("...")))
	}

	for p := startPage; p <= endPage; p++ {
		pages = append(pages, e.pageButton(req, p, number == p))
	}

	if endPage < numPages {
		pages = append(pages, Button(Disabled(), Class("join-item btn btn-sm"), Text("...")),
			e.pageButton(req, numPages, number == numPages))
	}

	return Div(Class("flex flex-col justify-center items-center gap-2 p-4"),
		Div(Class("join"),
			Group(pages),
		),
	)
}

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
