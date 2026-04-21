package p_seer_gdelt

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
)

const (
	gdeltSearchStateKey = "seer_gdelt.search_state"
	gdeltResultsKey     = "seer_gdelt.results"
)

type gdeltSearchState struct {
	Searched bool
	Request  GDELTSearchRequest
}

type gdeltSearchLayer struct{}

func (gdeltSearchLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		search, searched, fieldErrors := parseGDELTSearchRequest(r)
		ctx := context.WithValue(r.Context(), gdeltSearchStateKey, gdeltSearchState{
			Searched: searched,
			Request:  search,
		})
		if len(fieldErrors) > 0 {
			ctx = views.ContextWithErrorsAndValues(ctx, nil, fieldErrors)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if !searched {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			slog.Error("p_seer_gdelt: missing db in context", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		results, err := FetchAndStoreGDELTEvents(ctx, db, search)
		if err != nil {
			slog.Error("p_seer_gdelt: search failed", "error", err)
			ctx = views.ContextWithErrorsAndValues(ctx, nil, map[string]error{
				"_form": err,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, gdeltResultsKey, components.ObjectList[Event]{
			Items:    results,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(results)),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseGDELTSearchRequest(r *http.Request) (GDELTSearchRequest, bool, map[string]error) {
	q := r.URL.Query()
	search := GDELTSearchRequest{
		Query:         strings.TrimSpace(q.Get("Query")),
		Domain:        strings.TrimSpace(q.Get("Domain")),
		ActionCountry: strings.TrimSpace(q.Get("ActionCountry")),
		Sort:          strings.TrimSpace(q.Get("Sort")),
		MaxRecords:    limitGDELTMaxRecords(0),
	}
	fieldErrors := map[string]error{}

	var searched bool
	if search.Query != "" || search.Domain != "" || search.ActionCountry != "" ||
		strings.TrimSpace(q.Get("StartDate")) != "" || strings.TrimSpace(q.Get("EndDate")) != "" {
		searched = true
	}
	if raw := strings.TrimSpace(q.Get("MinMentions")); raw != "" {
		searched = true
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			fieldErrors["MinMentions"] = fmt.Errorf("minimum mentions must be a whole number")
		} else {
			search.MinMentions = uint(n)
		}
	}
	if raw := strings.TrimSpace(q.Get("MaxRecords")); raw != "" {
		searched = true
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			fieldErrors["MaxRecords"] = fmt.Errorf("max records must be a whole number")
		} else if n == 0 || n > uint64(maxGDELTMaxRecords) {
			fieldErrors["MaxRecords"] = fmt.Errorf("max records must be between 1 and %d", maxGDELTMaxRecords)
		} else {
			search.MaxRecords = uint(n)
		}
	}
	if search.Sort != "" && gdeltSortOrDefault(search.Sort) != search.Sort {
		fieldErrors["Sort"] = fmt.Errorf("invalid sort option")
	}
	if raw := strings.TrimSpace(q.Get("StartDate")); raw != "" {
		searched = true
		start, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			fieldErrors["StartDate"] = fmt.Errorf("start date must use YYYY-MM-DD")
		} else {
			search.StartDate = &start
		}
	}
	if raw := strings.TrimSpace(q.Get("EndDate")); raw != "" {
		searched = true
		endDate, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			fieldErrors["EndDate"] = fmt.Errorf("end date must use YYYY-MM-DD")
		} else {
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			search.EndDate = &endDate
		}
	}
	if search.StartDate != nil && search.EndDate != nil && search.StartDate.After(*search.EndDate) {
		fieldErrors["EndDate"] = fmt.Errorf("end date must be on or after start date")
	}
	if searched && strings.TrimSpace(search.Query) == "" && strings.TrimSpace(search.Domain) == "" &&
		strings.TrimSpace(search.ActionCountry) == "" && search.StartDate == nil && search.EndDate == nil {
		fieldErrors["Query"] = fmt.Errorf("add keyword, domain, action country, or date filter")
	}
	return search, searched, fieldErrors
}

func gdeltSearchStateFromContext(ctx context.Context) gdeltSearchState {
	state, _ := ctx.Value(gdeltSearchStateKey).(gdeltSearchState)
	return state
}

func gdeltResultsGetter(ctx context.Context) (components.ObjectList[Event], error) {
	results, _ := ctx.Value(gdeltResultsKey).(components.ObjectList[Event])
	return results, nil
}
