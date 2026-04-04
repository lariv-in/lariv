package getters

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSubmitGet returns an Alpine @submit.prevent expression for GET filter forms: it resolves field
// values with htmx.values and issues a boosted GET via htmx.ajax (target body, outerHTML swap).
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
			"(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;htmx.ajax('GET',%s,{source:f,target:'body',swap:'outerHTML',values:htmx.values(f),headers:{'HX-Boosted':'true'}})})($event)",
			urlLit,
		), nil
	}
}
