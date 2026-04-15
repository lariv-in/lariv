package views

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
)

type multiStepLayerTestPage struct {
	components.Page
	Children []components.PageInterface
}

func (p multiStepLayerTestPage) Build(ctx context.Context) gomponents.Node {
	group := gomponents.Group{}
	for _, child := range p.Children {
		group = append(group, components.Render(child, ctx))
	}
	return group
}

func (p multiStepLayerTestPage) GetKey() string {
	return p.Key
}

func (p multiStepLayerTestPage) GetRoles() []string {
	return p.Roles
}

func (p multiStepLayerTestPage) GetChildren() []components.PageInterface {
	return p.Children
}

func (p *multiStepLayerTestPage) SetChildren(children []components.PageInterface) {
	p.Children = children
}

func TestMultiStepFormLayerAdvancesStage(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "First", Label: "First"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Second", Label: "Second"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"0"},
		"$stage_target": {"1"},
		"First":         {"alpha"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when advancing stage for swap, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="Second"`) {
		t.Fatalf("expected next stage field in body, got %s", body)
	}
	if !strings.Contains(body, `name="First"`) || !strings.Contains(body, `value="alpha"`) {
		t.Fatalf("expected carry-forward hidden field in body, got %s", body)
	}
	if !strings.Contains(body, `name="$stage"`) || !strings.Contains(body, `value="1"`) {
		t.Fatalf("expected next stage marker in body, got %s", body)
	}
}

func TestMultiStepFormLayerRerendersCurrentStageOnFieldError(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputNumber[uint]{Name: "Count", Label: "Count"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Done", Label: "Done"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"0"},
		"$stage_target": {"1"},
		"Count":         {"nope"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 on field error, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="Count"`) {
		t.Fatalf("expected current stage field in body, got %s", body)
	}
	if strings.Contains(body, `name="Done"`) {
		t.Fatalf("expected not to advance to next stage on error, got %s", body)
	}
	if !strings.Contains(body, `name="$stage"`) || !strings.Contains(body, `value="0"`) {
		t.Fatalf("expected current stage marker in body, got %s", body)
	}
	if got := strings.Count(body, `border-error`); got != 1 {
		t.Fatalf("expected only current error stage highlighted, got %d body=%s", got, body)
	}
}

func TestMultiStepFormLayerPassesFinalStageToNext(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "First", Label: "First"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Second", Label: "Second"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})
	view.WithLayer("capture", MethodLayer{
		Method: http.MethodPost,
		Handler: func(_ *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				stage, _ := r.Context().Value("$stage").(int)
				if stage != 1 {
					t.Fatalf("expected final stage in context, got %d", stage)
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("final"))
			})
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage": {"1"},
		"First":  {"alpha"},
		"Second": {"beta"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected final stage to reach downstream handler, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "final" {
		t.Fatalf("expected downstream body, got %q", rec.Body.String())
	}
}

func TestMultiStepFormLayerPassesFinalStageValuesToNext(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "First", Label: "First"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Second", Label: "Second"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})
	view.WithLayer("capture", MethodLayer{
		Method: http.MethodPost,
		Handler: func(_ *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values, _ := r.Context().Value(getters.ContextKeyIn).(map[string]any)
				if values["First"] != "alpha" {
					t.Fatalf("expected carried first value in context, got %#v", values["First"])
				}
				if values["Second"] != "beta" {
					t.Fatalf("expected final stage value in context, got %#v", values["Second"])
				}
				w.WriteHeader(http.StatusCreated)
			})
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage": {"1"},
		"First":  {"alpha"},
		"Second": {"beta"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected final stage to reach downstream handler, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMultiStepFormLayerNavigatesBackToEarlierStage(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "First", Label: "First"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Second", Label: "Second"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"1"},
		"$stage_target": {"0"},
		"First":         {"alpha"},
		"Second":        {"beta"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when swapping back to previous stage, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="First"`) {
		t.Fatalf("expected first stage field in body, got %s", body)
	}
	if strings.Contains(body, `>Second<`) {
		t.Fatalf("expected second stage not to render as active field after back nav, got %s", body)
	}
	if !strings.Contains(body, `name="$stage"`) || !strings.Contains(body, `value="0"`) {
		t.Fatalf("expected previous stage marker in body, got %s", body)
	}
}

func TestMultiStepFormLayerPreservesTargetStageErrorsOnForwardNav(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.First"),
								Children: []components.PageInterface{
									components.InputText{Name: "First", Label: "First"},
								},
							},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.Second"),
								Children: []components.PageInterface{
									components.InputText{Name: "Second", Label: "Second"},
								},
							},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"0"},
		"$stage_target": {"1"},
		"First":         {"alpha"},
		"$error.Second": {"bad second"},
		"$error._form":  {"bad form"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when advancing stage for swap, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "bad second") {
		t.Fatalf("expected target stage error to be preserved on forward nav, got %s", body)
	}
	if !strings.Contains(body, "bad form") {
		t.Fatalf("expected carried form error to persist, got %s", body)
	}
	if strings.Contains(body, `name="$error.First"`) {
		t.Fatalf("expected current stage errors to be cleared after valid forward nav, got %s", body)
	}
	if !strings.Contains(body, `name="$error.Second"`) || !strings.Contains(body, `value="bad second"`) {
		t.Fatalf("expected hidden carry field for preserved target stage error, got %s", body)
	}
	if got := strings.Count(body, `border-error`); got != 2 {
		t.Fatalf("expected global form error to highlight all stages, got %d body=%s", got, body)
	}
}

func TestMultiStepFormLayerClearsCurrentStageErrorsOnBackNav(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.First"),
								Children: []components.PageInterface{
									components.InputText{Name: "First", Label: "First"},
								},
							},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.Second"),
								Children: []components.PageInterface{
									components.InputText{Name: "Second", Label: "Second"},
								},
							},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"1"},
		"$stage_target": {"0"},
		"First":         {"alpha"},
		"Second":        {"beta"},
		"$error.First":  {"bad first"},
		"$error.Second": {"bad second"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when swapping back to previous stage, got %d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "bad first") {
		t.Fatalf("expected prior stage error to survive back nav, got %s", body)
	}
	if strings.Contains(body, "bad second") {
		t.Fatalf("expected current stage error to clear after successful back nav, got %s", body)
	}
	if strings.Contains(body, `name="$error.Second"`) {
		t.Fatalf("expected current stage hidden error carry to be cleared, got %s", body)
	}
	if got := strings.Count(body, `border-error`); got != 1 {
		t.Fatalf("expected only preserved prior-stage error to remain highlighted, got %d body=%s", got, body)
	}
}

func TestMultiStepFormLayerPassesPreservedErrorsToFinalStageNext(t *testing.T) {
	page := &multiStepLayerTestPage{
		Children: []components.PageInterface{
			&components.MultiStepForm{
				Stages: []components.FormInterface{
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "First", Label: "First"},
						},
					},
					components.FormComponent[struct{}]{
						ChildrenInput: []components.PageInterface{
							components.InputText{Name: "Second", Label: "Second"},
						},
					},
				},
			},
		},
	}
	view := &View{
		PageName: "wizard",
		PageLookup: func(name string) (components.PageInterface, bool) {
			return page, name == "wizard"
		},
	}
	view.WithLayer("multistep", MultiStepFormLayer{})
	view.WithLayer("capture", MethodLayer{
		Method: http.MethodPost,
		Handler: func(_ *View) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				errorMap, _ := r.Context().Value(getters.ContextKeyError).(map[string]error)
				if got := errorMap["First"]; got == nil || got.Error() != "bad first" {
					t.Fatalf("expected preserved prior-stage error in final handoff, got %#v", errorMap)
				}
				if _, ok := errorMap["Second"]; ok {
					t.Fatalf("expected current stage errors to be cleared before final handoff, got %#v", errorMap)
				}
				w.WriteHeader(http.StatusCreated)
			})
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"$stage":        {"1"},
		"First":         {"alpha"},
		"Second":        {"beta"},
		"$error.First":  {"bad first"},
		"$error.Second": {"bad second"},
		"$error._form":  {"bad form"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	view.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected final stage to reach downstream handler, got %d body=%s", rec.Code, rec.Body.String())
	}
}
