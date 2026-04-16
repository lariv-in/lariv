package views

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
)

type MultiStepFormLayer struct{}

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

func multiStepRenderRequest(r *http.Request) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodGet
	return clone
}

type multiStepSwapResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *multiStepSwapResponseWriter) WriteHeader(statusCode int) {
	w.wroteHeader = true
	if statusCode >= 200 && statusCode < 300 {
		w.ResponseWriter.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *multiStepSwapResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	return w.ResponseWriter.Write(b)
}

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
