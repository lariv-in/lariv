package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FormListenBoostedPost represents a parent container that listens for bubbled form submission events and posts them via HTMX.
// It intercepts "lago-form-submit" Alpine.js custom events from child components (specifically those configured with [getters.FormBubbling] on [FormComponent].Attr).
// When triggered, it initiates a POST request using `htmx.ajax` with `HX-Boosted: true`, swapping the entire body `outerHTML` for full-page flows.
// It features double-submit protection by ignoring new click events while a POST request is in flight.
//
// Use Cases:
//   - Safely handling full-page form submissions without standard full page reload flickers.
//   - Disabling submit buttons automatically until the server responds, avoiding duplicate record creations.
//
// Example:
//
//	&components.FormListenBoostedPost{
//	    Name:      getters.Static("createUserForm"),
//	    ActionURL: lago.RoutePath("admin.UserCreate", nil),
//	    Children: []components.PageInterface{
//	        &components.FormComponent[User]{
//	            Attr: getters.FormBubbling(),
//	            // ... inputs and submit buttons ...
//	        },
//	    },
//	}
type FormListenBoostedPost struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Name is a Getter resolving to the unique name identifier of the child form it should intercept.
	Name getters.Getter[string]
	// ActionURL is a Getter resolving to the target URL endpoint for the AJAX POST request.
	ActionURL getters.Getter[string]
	// Children represents the form components nested inside the listener scope.
	Children []PageInterface
}

// GetKey returns the unique key identifier for this FormListenBoostedPost component.
func (e FormListenBoostedPost) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FormListenBoostedPost.
func (e FormListenBoostedPost) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside this listener scope.
func (e FormListenBoostedPost) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside this listener scope.
func (e *FormListenBoostedPost) SetChildren(children []PageInterface) {
	e.Children = children
}

// Build compiles the FormListenBoostedPost component into a Div Node listening for the Alpine.js form submission event.
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
	expr := fmt.Sprintf(
		`(function(evt){
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
	return Div(
		Class("contents"),
		gomponents.Attr("@lago-form-submit", expr),
		gomponents.Group(childNodes),
	)
}
