package views

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lariv-in/lago/components"
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
		"$stage": {"0"},
		"First":  {"alpha"},
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
		"$stage": {"0"},
		"Count":  {"nope"},
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
