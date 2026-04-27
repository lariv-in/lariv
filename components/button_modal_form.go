package components

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ButtonModalForm is like [ButtonModal] but registers a local listener for bubbling
// "lago-form-submit" from [getters.FormBubbling]. Only events whose form sits in
// dialog.modal with id [ModalUID] are handled. The form is POSTed to [FormPostURL].
// On HTTP redirect (see [views.HtmxRedirect] without HX-Request: normal 3xx + Location),
// the dialog is removed, "lago:modal-closed" is dispatched, and the browser navigates.
// On 2xx success without redirect, the dialog is closed. On other statuses (e.g. 422),
// the dialog is replaced by the response body HTML.
// The POST URL always carries a "name" query param (same as Name) so the request URL query
// is populated on POST; FormBubbling(getters.Key("$get.name")) then renders the same
// registry name after validation errors (422) as on the initial GET modal open.
type ButtonModalForm struct {
	Page
	Label       string
	Url         getters.Getter[string]
	Name        getters.Getter[string]
	FormPostURL getters.Getter[string]
	ModalUID    string
	Icon        string
	IconClasses string
	Classes     string
	Attr        getters.Getter[Node]
}

func (e ButtonModalForm) GetKey() string     { return e.Key }
func (e ButtonModalForm) GetRoles() []string { return e.Roles }

func (e ButtonModalForm) Build(ctx context.Context) Node {
	if e.Name == nil {
		return ContainerError{Error: getters.Static(fmt.Errorf("ButtonModalForm: Name is nil"))}.Build(ctx)
	}
	href := ""
	if e.Url != nil {
		if v, err := e.Url(ctx); err == nil {
			href = v
		}
	}
	name, err := e.Name(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	postURL := ""
	if e.FormPostURL != nil {
		if v, err := e.FormPostURL(ctx); err == nil {
			postURL = v
		}
	}
	if postURL == "" || e.ModalUID == "" {
		return ContainerError{Error: getters.Static(fmt.Errorf("ButtonModalForm: FormPostURL and ModalUID are required"))}.Build(ctx)
	}

	if postParsed, err := neturl.Parse(postURL); err == nil {
		pq := postParsed.Query()
		pq.Set("name", name)
		postParsed.RawQuery = pq.Encode()
		postURL = postParsed.String()
	}

	if href != "" {
		parsedURL, err := neturl.Parse(href)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		query := parsedURL.Query()
		query.Set("name", name)
		parsedURL.RawQuery = query.Encode()
		href = parsedURL.String()
	}

	nameLit, err := json.Marshal(name)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	postLit, err := json.Marshal(postURL)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	uidLit, err := json.Marshal(e.ModalUID)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	// Alpine @lago-form-submit: POST the bubbling form via htmx.ajax, then close or swap the dialog.
	// %s/%s are JSON string literals for modal id and POST URL (see json.Marshal above).
	script := fmt.Sprintf(
		`(function(evt){
  var d = evt.detail || {};
  var f = d.form;
  if (!f || d.name !== %s) return;
  var m = f.closest('dialog.modal');
  if (!m || m.id !== %s) return;
  evt.stopPropagation();
  if (f.dataset.lagoPostPending) return;
  f.dataset.lagoPostPending = '1';
  var u = (f.getAttribute('data-lago-post-url')||'').trim();
  if (!u) { u = %s; }
  var body = document.body;
  function closeModal(x) {
    document.dispatchEvent(new CustomEvent('lago:modal-closed', { bubbles: true, detail: Object.assign({ dialog: m }, x) }));
    m.remove();
  }
  var cleanup = function () {
    body.removeEventListener('htmx:beforeSwap', onBeforeSwap);
    delete f.dataset.lagoPostPending;
  };
  var onBeforeSwap = function (e) {
    var detail = e.detail || {};
    if (detail.elt !== m) return;
    var xhr = detail.xhr;
    if (!xhr) return;
    var hxLoc = xhr.getResponseHeader('HX-Redirect');
    if (hxLoc) {
      detail.shouldSwap = false;
      closeModal({ httpStatus: xhr.status, location: hxLoc });
      window.location.assign(hxLoc);
      return;
    }
    if (xhr.status >= 200 && xhr.status < 300) {
      detail.shouldSwap = false;
      closeModal({ httpStatus: xhr.status });
      return;
    }
    detail.shouldSwap = true;
    detail.target = m;
  };
  body.addEventListener('htmx:beforeSwap', onBeforeSwap);
  htmx.ajax('POST', u, {
    source: f,
    target: m,
    swap: 'outerHTML',
    values: htmx.values(f),
    headers: { 'HX-Boosted': 'true' }
  }).finally(cleanup);
})($event)`,
		string(nameLit),
		string(uidLit),
		string(postLit),
	)

	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	buttonClasses := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	buttonAttrs := []Node{
		Type("button"),
		Class(buttonClasses),
		Attr("hx-get", href),
		Attr("hx-target", HTMXTargetBodyModal),
		Attr("hx-swap", HTMXSwapBodyModal),
		Attr("hx-push-url", "false"),
	}
	if e.Attr != nil {
		extra, err := e.Attr(ctx)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if extra != nil {
			buttonAttrs = append(buttonAttrs, extra)
		}
	}
	buttonAttrs = append(buttonAttrs, content)

	return Div(Class("fk-modal-host"),
		Attr("@lago-form-submit.window.stop", script),
		Button(Group(buttonAttrs)),
	)
}
