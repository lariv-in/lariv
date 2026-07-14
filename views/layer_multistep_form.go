package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
)

// MultiStepFormLayer coordinates stage transitions, input validations, error caching, and HTMX swaps for multi-stage wizards.
// It intercepts HTTP POST operations targeting [components.MultiStepForm] page wrappers, increments or decrements the request context "$stage" index,
// merges carried validation errors, and enforces HTTP StatusUnprocessableEntity (422) redirects to swap out form stages.
//
// Use Cases:
//   - Driving multi-stage checkout forms, step-by-step registration processes, or linear onboarding flows.
//
// Example:
//
//	views.View{
//	    Layers: []views.Layer{
//	        views.MultiStepFormLayer{},
//	    },
//	}
type MultiStepFormLayer struct{}

// Next wraps the downstream HTTP request handlers routing multi-stage form transitions.
func (m MultiStepFormLayer) Next(view View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		form, ok := viewMultiStepForm(view)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodPost {
			stage := 0
			ctx := context.WithValue(r.Context(), "$stage", stage)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, err := view.ParseForm(w, r)
		if err != nil {
			slog.Error("views: multi-step: parse form", "error", err)
			return
		}

		stage := form.ParseStage(r)
		targetStage := min(form.ParseTargetStage(r, stage), stage+1)
		ctx := context.WithValue(r.Context(), "$stage", stage)
		carriedErrors := components.ParseMultiStepErrors(r)
		mergedErrors := mergeMultiStepFormErrors(carriedErrors, form.StageInputNames(stage), fieldErrors)
		if len(fieldErrors) != 0 {
			for field, ferr := range fieldErrors {
				slog.Error("views: multi-step: field error", "field", field, "error", ferr)
			}
			ctx = ContextWithErrorsAndValues(ctx, values, mergedErrors)
			next.ServeHTTP(w, multiStepRenderRequest(r.WithContext(ctx)))
			return
		}

		lastStage := form.StageCount() - 1
		if targetStage < stage {
			ctx = ContextWithErrorsAndValues(ctx, values, mergedErrors)
			ctx = context.WithValue(ctx, "$stage", targetStage)
			next.ServeHTTP(&multiStepSwapResponseWriter{ResponseWriter: w}, multiStepRenderRequest(r.WithContext(ctx)))
			return
		}

		if stage < lastStage {
			ctx = ContextWithErrorsAndValues(ctx, values, mergedErrors)
			ctx = context.WithValue(ctx, "$stage", min(targetStage, stage+1))
			next.ServeHTTP(&multiStepSwapResponseWriter{ResponseWriter: w}, multiStepRenderRequest(r.WithContext(ctx)))
			return
		}

		ctx = ContextWithErrorsAndValues(ctx, values, mergedErrors)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mergeMultiStepFormErrors merges cached errors with newly occurred validation errors while purging expired errors belonging to the current step.
func mergeMultiStepFormErrors(carriedErrors map[string]error, stageFields map[string]struct{}, freshErrors map[string]error) map[string]error {
	merged := map[string]error{}
	for key, err := range carriedErrors {
		if err == nil {
			continue
		}
		if _, ok := stageFields[key]; ok {
			continue
		}
		merged[key] = err
	}
	for key, err := range freshErrors {
		if err == nil {
			delete(merged, key)
			continue
		}
		merged[key] = err
	}
	return merged
}

// multiStepRenderRequest clones the incoming request changing the HTTP method to GET to execute safe view refreshes.
func multiStepRenderRequest(r *http.Request) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodGet
	return clone
}

// multiStepSwapResponseWriter enforces StatusUnprocessableEntity (422) for successful intermediate transitions,
// triggering HTMX swap logic rather than browser document reloads.
type multiStepSwapResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

// WriteHeader writes the custom 422 HTTP status header when receiving 2xx statuses.
func (w *multiStepSwapResponseWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	if statusCode >= 200 && statusCode < 300 {
		w.ResponseWriter.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes data payloads to the client stream enforcing WriteHeader.
func (w *multiStepSwapResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	return w.ResponseWriter.Write(b)
}

// viewMultiStepForm scans the View page hierarchy returning the target MultiStepForm structure instance.
func viewMultiStepForm(view View) (components.MultiStepForm, bool) {
	page, ok := view.GetPage()
	if !ok {
		return components.MultiStepForm{}, false
	}
	switch typed := page.(type) {
	case components.MultiStepForm:
		return typed, true
	case *components.MultiStepForm:
		return *typed, true
	}

	parent, ok := page.(components.ParentInterface)
	if !ok {
		return components.MultiStepForm{}, false
	}
	if forms := components.FindChildren[*components.MultiStepForm](parent); len(forms) > 0 {
		return *forms[0], true
	}
	if forms := components.FindChildren[components.MultiStepForm](parent); len(forms) > 0 {
		return forms[0], true
	}
	return components.MultiStepForm{}, false
}
