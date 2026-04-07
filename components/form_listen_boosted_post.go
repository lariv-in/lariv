package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FormListenBoostedPost is a parent wrapper for forms that use [getters.FormBubbling] on [FormComponent].Attr.
// It listens for the bubbling "lago-form-submit" event and POSTs via htmx.ajax with HX-Boosted (body,
// outerHTML swap). Use for full-page flows. While a POST is in flight, further submits for the same form
// are ignored so rapid clicks do not create duplicate records.
type FormListenBoostedPost struct {
	Page
	Name      getters.Getter[string]
	ActionURL getters.Getter[string]
	Children  []PageInterface
}

func (e FormListenBoostedPost) GetKey() string {
	return e.Key
}

func (e FormListenBoostedPost) GetRoles() []string {
	return e.Roles
}

func (e FormListenBoostedPost) GetChildren() []PageInterface {
	return e.Children
}

func (e *FormListenBoostedPost) SetChildren(children []PageInterface) {
	e.Children = children
}

func (e FormListenBoostedPost) Build(ctx context.Context) gomponents.Node {
	if e.Name == nil {
		return ContainerError{Error: getters.Static(fmt.Errorf("FormListenBoostedPost: Name is nil"))}.Build(ctx)
	}
	if e.ActionURL == nil {
		return ContainerError{Error: getters.Static(fmt.Errorf("FormListenBoostedPost: ActionURL is nil"))}.Build(ctx)
	}
	name, err := e.Name(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	url, err := e.ActionURL(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	nameLit, err := json.Marshal(name)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	urlLit, err := json.Marshal(url)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	expr := fmt.Sprintf(`(function(evt){
  var d = evt.detail || {};
  if (d.name !== %s) return;
  var f = d.form;
  if (!f) return;
  var root = evt.currentTarget;
  if (!root || !root.contains || !root.contains(f)) return;
  var u = %s;
  var targetPath;
  try {
    targetPath = (new URL(u, window.location.href)).pathname;
  } catch (_) {
    targetPath = '';
  }
  var formAction = f.getAttribute && f.getAttribute('action');
  if (!formAction || formAction === '') {
    formAction = window.location.href;
  }
  var formPath;
  try {
    formPath = (new URL(formAction, window.location.href)).pathname;
  } catch (_) {
    formPath = '';
  }
  if (targetPath !== '' && formPath !== '' && targetPath !== formPath) return;
  evt.stopPropagation();
  if (f.dataset.lagoPostPending) return;
  f.dataset.lagoPostPending = '1';
  htmx.ajax('POST', u, {
    source: f,
    target: 'body',
    swap: 'outerHTML',
    values: htmx.values(f),
    headers: { 'HX-Boosted': 'true' },
  }).finally(function () {
    delete f.dataset.lagoPostPending;
  });
})($event)`,
		nameLit,
		urlLit,
	)
	var childNodes []gomponents.Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, ctx))
	}
	return Div(Class("contents"),
		gomponents.Attr("@lago-form-submit", expr),
		gomponents.Group(childNodes),
	)
}
