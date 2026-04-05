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
// outerHTML swap). Use for full-page flows;
type FormListenBoostedPost struct {
	Page
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
	if e.ActionURL == nil {
		return ContainerError{Error: getters.Static(fmt.Errorf("FormListenBoostedPost: ActionURL is nil"))}.Build(ctx)
	}
	url, err := e.ActionURL(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	urlLit, err := json.Marshal(url)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	expr := fmt.Sprintf(
		`(function(evt){evt.stopPropagation();var d=evt.detail||{};var f=d.form;if(!f)return;htmx.ajax('POST',%s,{source:f,target:'body',swap:'outerHTML',values:htmx.values(f),headers:{'HX-Boosted':'true'}})})($event)`,
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
