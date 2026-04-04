package getters

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSubmitCloseModal returns an Alpine @submit.prevent expression (see components.FormComponent).
// It resolves the form from $event.target, POSTs via fetch with HX-Request and HX-Boosted (so
// views.HtmxRedirect returns 200 + HX-Redirect). The enclosing dialog.modal (see components.Modal)
// is replaced with outerHTML from the response body when there is no redirect, so validation and
// other error responses replace the modal content.
//
// When the response includes HX-Redirect (the success path from views.HtmxRedirect on create/update/etc.),
// only the dialog is removed — the page is not navigated; refresh the surrounding UI separately if needed.
func FormSubmitCloseModal(path Getter[string]) Getter[string] {
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
			"(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;var m=f.closest('dialog.modal');if(!m)return;var u=%s;fetch(u,{method:'POST',body:new FormData(f),credentials:'same-origin',headers:{'HX-Request':'true','HX-Boosted':'true'},redirect:'manual'}).then(function(r){if(r.headers.get('HX-Redirect')){m.remove();return;}if(r.type==='opaqueredirect'||(r.status>=300&&r.status<400)){m.remove();return;}return r.text().then(function(h){m.outerHTML=h;});});})($event)",
			urlLit,
		), nil
	}
}
