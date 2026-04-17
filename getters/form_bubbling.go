package getters

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"strings"

	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// FormBubbling returns attributes for [components.FormComponent].Attr: HTTP POST plus Alpine
// @submit.prevent that stops native submit and dispatches a bubbling CustomEvent "lago-form-submit"
// from the form with detail { form: <form element>, name: string }. Parents (e.g. [components.ButtonModalForm],
// or [components.FormListenBoostedPost]) handle the request.
func FormBubbling(name Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		if name == nil {
			return nil, fmt.Errorf("FormBubbling: Name is nil")
		}
		resolvedName, err := name(ctx)
		if err != nil {
			return nil, err
		}
		nLit, err := json.Marshal(resolvedName)
		if err != nil {
			return nil, err
		}
		return gomponents.Group{
			ghtml.Method("POST"),
			gomponents.Attr("hx-boost", "false"),
			gomponents.Attr("@submit.prevent", fmt.Sprintf(
				`(function(evt){evt.preventDefault();var f=(evt.currentTarget&&evt.currentTarget.tagName==='FORM')?evt.currentTarget:(evt.target&&evt.target.closest&&evt.target.closest('form'));if(!f)return;f.dispatchEvent(new CustomEvent('lago-form-submit',{bubbles:true,detail:{form:f,name:%s}}))})($event)`,
				nLit,
			)),
		}, nil
	}
}

// FormBubblingWithDataPostURL is [FormBubbling] plus data-lago-post-url on the form (merges ?name= into postURLBase) so row-scoped POST targets work when many modal openers share one dialog id.
func FormBubblingWithDataPostURL(name Getter[string], postURLBase Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		bub, err := FormBubbling(name)(ctx)
		if err != nil {
			return nil, err
		}
		if postURLBase == nil {
			return bub, nil
		}
		u, err := postURLBase(ctx)
		if err != nil {
			return nil, err
		}
		u = strings.TrimSpace(u)
		if u == "" {
			return bub, nil
		}
		resolvedName, err := name(ctx)
		if err != nil {
			return nil, err
		}
		if parsed, err := neturl.Parse(u); err == nil {
			q := parsed.Query()
			q.Set("name", resolvedName)
			parsed.RawQuery = q.Encode()
			u = parsed.String()
		}
		return gomponents.Group{bub, gomponents.Attr("data-lago-post-url", u)}, nil
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
