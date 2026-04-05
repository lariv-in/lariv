package getters

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSubmit returns an Alpine @submit.prevent expression; pair with [FormAttr] on FormComponent.Attr. It reads the
// form from $event.target, serializes with htmx.values, and POSTs via htmx.ajax with HX-Boosted
// (target body, outerHTML swap).
func FormSubmit(path Getter[string]) Getter[string] {
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
			"(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;htmx.ajax('POST',%s,{source:f,target:'body',swap:'outerHTML',values:htmx.values(f),headers:{'HX-Boosted':'true'}})})($event)",
			urlLit,
		), nil
	}
}
