package getters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"maragu.dev/gomponents"
)

// ModalRefreshList returns a Getter suitable for [components.ButtonModal] / [components.ButtonModalForm].Attr. It emits an
// x-init handler that registers a single document listener for "lago:modal-closed" (dispatched when
// [components.ButtonModalForm] closes the dialog after a successful POST). On each event, it GETs listURL and swaps innerHTML into the first element
// matching tableSelector (typically "#your-table-uid"). The response is parsed with DOMParser and
// only that element’s innerHTML is copied, so a full-page GET (including a nested <dialog.modal>)
// does not insert an entire second modal inside the existing table container.
// Duplicate url|selector pairs only register once per page load.
//
// Refresh uses fetch (not htmx.ajax) so the browser history / address bar is not updated to the
// list URL (htmx would push /students/addon/pick-user/ etc. onto history).
//
// If listURL is nil or resolves to the empty string, the refresh URL is taken from the request in
// context as "$request" (same idea as [components.TablePagination]: path + query for the view that
// is rendering this button). That is required when the button lives inside a modal: the browser URL
// is still the parent page, but the list belongs to the modal’s route. If "$request" is absent,
// the client falls back to window.location.pathname + window.location.search.
func ModalRefreshList(listURL, tableSelector Getter[string]) Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		if tableSelector == nil {
			return nil, fmt.Errorf("getters.ModalRefreshList: tableSelector getter is nil")
		}
		sel, err := tableSelector(ctx)
		if err != nil {
			return nil, err
		}
		if sel == "" {
			return nil, fmt.Errorf("getters.ModalRefreshList: tableSelector is empty")
		}

		var urlStr string
		if listURL != nil {
			urlStr, err = listURL(ctx)
			if err != nil {
				return nil, err
			}
		}
		if urlStr == "" {
			if r, ok := ctx.Value("$request").(*http.Request); ok && r != nil && r.URL != nil {
				urlStr = r.URL.RequestURI()
			}
		}

		uLit, err := json.Marshal(urlStr)
		if err != nil {
			return nil, err
		}
		sLit, err := json.Marshal(sel)
		if err != nil {
			return nil, err
		}

		script := fmt.Sprintf(
			"(function(){var u=%s;var s=%s;var k='lago:mr:'+u+'|'+s;if(!window.__lagoModalRefreshKeys)window.__lagoModalRefreshKeys=new Set();if(window.__lagoModalRefreshKeys.has(k))return;window.__lagoModalRefreshKeys.add(k);document.addEventListener('lago:modal-closed',function(){var url=u||window.location.pathname+window.location.search;fetch(url,{credentials:'same-origin',headers:{'HX-Request':'true','HX-Boosted':'true','Accept':'text/html'}}).then(function(r){if(!r.ok)return null;return r.text()}).then(function(h){if(h==null)return;var el=document.querySelector(s);if(!el)return;var doc=new DOMParser().parseFromString(h,'text/html');var fresh=doc.querySelector(s);if(fresh){el.innerHTML=fresh.innerHTML;return;}el.innerHTML=h;});});})()",
			uLit,
			sLit,
		)
		return gomponents.Attr("x-init", script), nil
	}
}
