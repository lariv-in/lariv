package getters

import (
	"context"
	"encoding/json"
	"fmt"
)

// FormSubmitCloseModal returns an Alpine @submit.prevent expression; use with [FormAttr] on FormComponent.Attr.
// It resolves the form from $event.target, POSTs via fetch with HX-Request and HX-Boosted (so
// views.HtmxRedirect returns 200 + HX-Redirect). The enclosing dialog.modal (see components.Modal)
// is replaced with outerHTML from the response body when there is no redirect, so validation and
// other error responses replace the modal content.
//
// When the response includes HX-Redirect (the success path from views.HtmxRedirect on create/update/etc.),
// only the dialog is removed — the page is not navigated; refresh the surrounding UI separately if needed.
// Whenever the dialog is removed after a successful POST, document dispatches a bubbling CustomEvent
// named "lago:modal-closed" with detail { dialog, redirectURL? } (redirectURL when HX-Redirect was set).
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
			"(function(evt){var t=evt&&evt.target;var f=t&&t.closest&&t.closest('form');if(!f)return;var m=f.closest('dialog.modal');if(!m)return;var u=%s;function closeModal(d){document.dispatchEvent(new CustomEvent('lago:modal-closed',{bubbles:true,detail:Object.assign({dialog:m},d)}));m.remove();}fetch(u,{method:'POST',body:new FormData(f),credentials:'same-origin',headers:{'HX-Request':'true','HX-Boosted':'true'},redirect:'manual'}).then(function(r){var x=r.headers.get('HX-Redirect');if(x){closeModal({redirectURL:x});return;}if(r.type==='opaqueredirect'||(r.status>=300&&r.status<400)){closeModal({httpStatus:r.status});return;}return r.text().then(function(h){m.outerHTML=h;});});})($event)",
			urlLit,
		), nil
	}
}
