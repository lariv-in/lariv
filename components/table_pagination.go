package components

import (
	"context"
	"net/http"
	"reflect"
	"strconv"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TablePagination struct {
	Data Getter
}

func (e TablePagination) Build(ctx context.Context) Node {
	data := IfOrGetter(e.Data, ctx, nil)
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	// Extract Number and NumPages safely
	var number, numPages int
	numField := v.FieldByName("Number")
	if numField.IsValid() && numField.CanInt() {
		number = int(numField.Int())
	} else {
		return nil
	}

	numPagesField := v.FieldByName("NumPages")
	if numPagesField.IsValid() && numPagesField.CanInt() {
		numPages = int(numPagesField.Int())
	}

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
	startPage := number - windowSize/2
	if startPage < 1 {
		startPage = 1
	}
	endPage := startPage + windowSize - 1
	if endPage > numPages {
		endPage = numPages
		startPage = endPage - windowSize + 1
		if startPage < 1 {
			startPage = 1
		}
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

func (e TablePagination) pageButton(req *http.Request, p int, active bool) Node {
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
