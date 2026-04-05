package getters

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSubmitGet returns an Alpine @submit.prevent expression for GET filter forms; use with [FormAttr] on FormComponent.Attr. It resolves field
// values with htmx.values and issues a boosted GET via htmx.ajax (outerHTML swap). When the form is
// inside a modal (dialog.modal), the swap target is that dialog so the list modal is
// refreshed in place; otherwise the target is body for full-page list filters.
func FormSubmitGet(path Getter[string]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		url, err := IfOr(path, ctx, "")
		if err != nil {
			return "", err
		}
		urlLit, err := json.Marshal(url)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;var m=f.closest('dialog.modal');var o={source:f,swap:'outerHTML',values:htmx.values(f),headers:{'HX-Boosted':'true'}};o.target=m||'body';htmx.ajax('GET',%s,o)})($event)",
			urlLit,
		), nil
	}
}
