package getters

import (
	"context"
	"encoding/json"
	"fmt"

	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// FormBubbling returns attributes for [components.FormComponent].Attr: HTTP POST plus Alpine
// @submit.prevent that stops native submit and dispatches a bubbling CustomEvent "lago-form-submit"
// from the form with detail { form: <form element>, name?: string }. Parents (e.g. [components.ButtonModalForm],
// or [components.FormListenBoostedPost]) handle the request.
//
// Pass nil or a pointer to "" to omit detail.name; otherwise listeners can filter by detail.name.
func FormBubbling(name *string) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		if name == nil || *name == "" {
			return gomponents.Group{
				ghtml.Method("POST"),
				gomponents.Attr("@submit.prevent", `(function(evt){evt.preventDefault();var f=evt.target&&evt.target.closest&&evt.target.closest('form');if(!f)return;f.dispatchEvent(new CustomEvent('lago-form-submit',{bubbles:true,detail:{form:f}}))})($event)`),
			}, nil
		}
		nLit, err := json.Marshal(*name)
		if err != nil {
			return nil, err
		}
		return gomponents.Group{
			ghtml.Method("POST"),
			gomponents.Attr("@submit.prevent", fmt.Sprintf(
				`(function(evt){evt.preventDefault();var f=evt.target&&evt.target.closest&&evt.target.closest('form');if(!f)return;f.dispatchEvent(new CustomEvent('lago-form-submit',{bubbles:true,detail:{form:f,name:%s}}))})($event)`,
				nLit,
			)),
		}, nil
	}
}

// FormBoostedGet returns attributes for GET filter forms: Method GET plus @submit.prevent that
// issues a boosted GET via htmx.ajax (outerHTML swap). When the form is inside a modal
// (dialog.modal), the swap target is that dialog so the list modal refreshes in place;
// otherwise the target is body for full-page list filters.
func FormBoostedGet(path Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		url, err := IfOr(path, ctx, "")
		if err != nil {
			return nil, err
		}
		urlLit, err := json.Marshal(url)
		if err != nil {
			return nil, err
		}
		return gomponents.Group{
			ghtml.Method("GET"),
			gomponents.Attr("@submit.prevent", fmt.Sprintf(
				`(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;var m=f.closest('dialog.modal');var o={source:f,swap:'outerHTML',values:htmx.values(f),headers:{'HX-Boosted':'true'}};o.target=m||'body';htmx.ajax('GET',%s,o)})($event)`,
				urlLit,
			)),
		}, nil
	}
}
