package p_website

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

// Pinned GrapesJS CDN version for the builder page.
const grapesJSCDNVersion = "0.22.6"

type grapesJSHead struct {
	components.Page
}

func (e grapesJSHead) GetKey() string     { return e.Key }
func (e grapesJSHead) GetRoles() []string { return e.Roles }

func (e grapesJSHead) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	base := "https://unpkg.com/grapesjs@" + grapesJSCDNVersion + "/dist/"
	_, err := io.WriteString(w, `<link rel="stylesheet" href="`+base+`css/grapes.min.css">
<script src="`+base+`grapes.min.js"></script>
<style>
  html, body { height: 100%; margin: 0; background: var(--gjs-primary-color); }
  .gjs-builder-wrap { display: flex; flex-direction: column; height: 100vh; background: var(--gjs-primary-color); }
  .gjs-builder-bar { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem 1rem; border-bottom: 1px solid rgba(255,255,255,0.12); background: var(--gjs-primary-color); flex-shrink: 0; color: #fff; }
  .gjs-builder-bar-actions { margin-left: auto; display: flex; align-items: center; gap: 0.75rem; }
  .gjs-builder-save-status { font-size: 0.875rem; opacity: 0.7; min-width: 4rem; }
  #gjs { flex: 1; min-height: 0; }
</style>
`)
	return err
}

type grapesJSBuilderBody struct {
	components.Page
}

func (e grapesJSBuilderBody) GetKey() string     { return e.Key }
func (e grapesJSBuilderBody) GetRoles() []string { return e.Roles }

type grapesJSBlockPayload struct {
	ID string `json:"id"`
	lariv.GrapesJSBlock
}

type grapesJSComponentPayload struct {
	ID string `json:"id"`
	lariv.GrapesJSComponent
}

type grapesJSTraitPayload struct {
	ID string `json:"id"`
	lariv.GrapesJSTrait
}

type grapesJSThemePayload struct {
	ID string `json:"id"`
	lariv.GrapesJSTheme
}

func grapesJSBlocksJSON(blocks *[]registry.Pair[string, lariv.GrapesJSBlock]) ([]byte, error) {
	if blocks == nil {
		return []byte("[]"), nil
	}
	out := make([]grapesJSBlockPayload, 0, len(*blocks))
	for _, pair := range *blocks {
		out = append(out, grapesJSBlockPayload{
			ID:            pair.Key,
			GrapesJSBlock: pair.Value,
		})
	}
	return json.Marshal(out)
}

func grapesJSComponentsJSON(components *[]registry.Pair[string, lariv.GrapesJSComponent]) ([]byte, error) {
	if components == nil {
		return []byte("[]"), nil
	}
	out := make([]grapesJSComponentPayload, 0, len(*components))
	for _, pair := range *components {
		out = append(out, grapesJSComponentPayload{
			ID:                pair.Key,
			GrapesJSComponent: pair.Value,
		})
	}
	return json.Marshal(out)
}

func grapesJSTraitsJSON(traits *[]registry.Pair[string, lariv.GrapesJSTrait]) ([]byte, error) {
	if traits == nil {
		return []byte("[]"), nil
	}
	out := make([]grapesJSTraitPayload, 0, len(*traits))
	for _, pair := range *traits {
		out = append(out, grapesJSTraitPayload{
			ID:            pair.Key,
			GrapesJSTrait: pair.Value,
		})
	}
	return json.Marshal(out)
}

func grapesJSThemesJSON(themes *[]registry.Pair[string, lariv.GrapesJSTheme]) ([]byte, error) {
	if themes == nil {
		return []byte("[]"), nil
	}
	out := make([]grapesJSThemePayload, 0, len(*themes))
	for _, pair := range *themes {
		out = append(out, grapesJSThemePayload{
			ID:            pair.Key,
			GrapesJSTheme: pair.Value,
		})
	}
	return json.Marshal(out)
}

func (e grapesJSBuilderBody) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	route, err := getters.Key[DBRoute]("dbroute")(ctx)
	if err != nil {
		return err
	}

	detailURL, err := lariv.RoutePath("p_website.RoutesDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(route.ID)),
	})(ctx)
	if err != nil {
		return err
	}
	loadURL, err := lariv.RoutePath("p_website.RoutesBuilderProjectRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(route.ID)),
	})(ctx)
	if err != nil {
		return err
	}
	storeURL := loadURL
	uploadURL, err := lariv.RoutePath("p_website.BuilderAssetUploadRoute", nil)(ctx)
	if err != nil {
		return err
	}
	themeURL, err := lariv.RoutePath("p_website.RoutesBuilderThemeRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(route.ID)),
	})(ctx)
	if err != nil {
		return err
	}

	blocksJSON := []byte("[]")
	componentsJSON := []byte("[]")
	traitsJSON := []byte("[]")
	themesJSON := []byte("[]")
	if app, ok := lariv.AppFromContext(ctx); ok && app != nil {
		if encoded, encErr := grapesJSBlocksJSON(app.GrapesJSBlocks.AllStable()); encErr != nil {
			slog.Error("grapesjs builder: encode blocks", "error", encErr)
		} else {
			blocksJSON = encoded
		}
		if encoded, encErr := grapesJSComponentsJSON(app.GrapesJSComponents.AllStable()); encErr != nil {
			slog.Error("grapesjs builder: encode components", "error", encErr)
		} else {
			componentsJSON = encoded
		}
		if encoded, encErr := grapesJSTraitsJSON(app.GrapesJSTraits.AllStable()); encErr != nil {
			slog.Error("grapesjs builder: encode traits", "error", encErr)
		} else {
			traitsJSON = encoded
		}
		if encoded, encErr := grapesJSThemesJSON(app.GrapesJSThemes.AllStable()); encErr != nil {
			slog.Error("grapesjs builder: encode themes", "error", encErr)
		} else {
			themesJSON = encoded
		}
	}

	currentTheme, err := json.Marshal(route.Theme)
	if err != nil {
		return err
	}

	script := fmt.Sprintf(`
(function() {
  const loadURL = %q;
  const storeURL = %q;
  const uploadURL = %q;
  const themeURL = %q;
  const registeredBlocks = %s;
  const registeredComponents = %s;
  const registeredTraits = %s;
  const registeredThemes = %s;
  let currentThemeId = %s;
  const dotLottieCDN = %q;
  const dotLottieAttr = %q;

  function themeById(id) {
    if (!id) return null;
    for (var i = 0; i < (registeredThemes || []).length; i++) {
      if (registeredThemes[i].id === id) return registeredThemes[i];
    }
    return null;
  }

  function applyThemeToCanvas(editor, themeId) {
    var frame = editor && editor.Canvas && editor.Canvas.getFrameEl && editor.Canvas.getFrameEl();
    var doc = frame && frame.contentDocument;
    if (!doc || !doc.head) return;
    Array.prototype.slice.call(doc.querySelectorAll('[data-lariv-theme]')).forEach(function (n) {
      n.parentNode && n.parentNode.removeChild(n);
    });
    var theme = themeById(themeId);
    if (!theme) return;
    (theme.stylesheets || []).forEach(function (href) {
      if (!href) return;
      var link = doc.createElement('link');
      link.rel = 'stylesheet';
      link.href = href;
      link.setAttribute('data-lariv-theme', themeId);
      doc.head.appendChild(link);
    });
    if (theme.css) {
      var style = doc.createElement('style');
      style.setAttribute('data-lariv-theme', themeId);
      style.textContent = theme.css;
      doc.head.appendChild(style);
    }
  }

  function normalizeBlockProps(props) {
    if (typeof props.onClick === 'string') {
      const body = props.onClick;
      props.onClick = function (block, editor) {
        return new Function('block', 'editor', body)(block, editor);
      };
    }
    return props;
  }

  function reviveTraitProps(props) {
    ['createInput', 'createLabel', 'onEvent', 'onUpdate'].forEach(function (key) {
      if (typeof props[key] === 'string') {
        const body = props[key];
        props[key] = function (p) {
          p = p || {};
          return new Function('trait', 'elInput', 'component', 'event', 'label', body)(
            p.trait, p.elInput, p.component, p.event, p.label
          );
        };
      }
    });
    if (typeof props.templateInput === 'string' && props.templateInput.indexOf('<') !== 0) {
      const body = props.templateInput;
      props.templateInput = function (p) {
        p = p || {};
        return new Function('trait', body)(p.trait);
      };
    }
    return props;
  }

  function reviveObjectMethods(obj, keys) {
    if (!obj || typeof obj !== 'object') return;
    keys.forEach(function (key) {
      if (typeof obj[key] === 'string') {
        const body = obj[key];
        obj[key] = function () {
          return new Function(body).call(this);
        };
      }
    });
  }

  function normalizeComponentProps(props) {
    if (typeof props.isComponent === 'string') {
      const body = props.isComponent;
      props.isComponent = function (el) {
        return new Function('el', body)(el);
      };
    }
    if (props.model && typeof props.model === 'object') {
      reviveObjectMethods(props.model, ['init', 'updated', 'removed']);
    }
    if (props.view && typeof props.view === 'object') {
      reviveObjectMethods(props.view, ['init', 'onRender', 'onRemove']);
      if (typeof props.view.onRender === 'function') {
        const userOnRender = props.view.onRender;
        props.view.onRender = function () {
          return userOnRender.apply(this, arguments);
        };
      }
    }
    return props;
  }

  window.__larivEnsureDotLottie = function (doc) {
    doc = doc || document;
    if (!doc || !doc.head) return;
    if (doc.querySelector('script[' + dotLottieAttr + ']')) return;
    if (doc.defaultView && doc.defaultView.customElements && doc.defaultView.customElements.get('dotlottie-wc')) return;
    const s = doc.createElement('script');
    s.type = 'module';
    s.src = dotLottieCDN;
    s.setAttribute(dotLottieAttr, '');
    doc.head.appendChild(s);
  };

  const editor = grapesjs.init({
    container: '#gjs',
    height: '100%%',
    width: 'auto',
    fromElement: false,
    assetManager: {
      upload: uploadURL,
      uploadName: 'files',
      multiUpload: true,
      autoAdd: true,
      embedAsBase64: false,
      credentials: 'include'
    },
    storageManager: {
      type: 'remote',
      autosave: true,
      autoload: true,
      stepsBeforeSave: 3,
      options: {
        remote: {
          urlLoad: loadURL,
          urlStore: storeURL,
          onStore: (data, ed) => ({
            data: data,
            html: ed.getHtml(),
            css: ed.getCss()
          }),
          onLoad: (result) => {
            if (result && result.data) {
              return result.data;
            }
            if (result && result.html) {
              return {
                pages: [{
                  name: 'Page',
                  component: result.html
                }]
              };
            }
            return {
              pages: [{
                name: 'Page',
                component: '<h1>New page</h1>'
              }]
            };
          }
        }
      }
    }
  });

  const tm = editor.TraitManager;
  (registeredTraits || []).forEach(function (trait) {
    const id = trait.id;
    const props = Object.assign({}, trait);
    delete props.id;
    tm.addType(id, reviveTraitProps(props));
  });

  const dc = editor.DomComponents;
  (registeredComponents || []).forEach(function (comp) {
    const id = comp.id;
    const props = Object.assign({}, comp);
    delete props.id;
    dc.addType(id, normalizeComponentProps(props));
  });

  const bm = editor.BlockManager;
  (registeredBlocks || []).forEach(function (block) {
    const id = block.id;
    const props = Object.assign({}, block);
    delete props.id;
    bm.add(id, normalizeBlockProps(props));
  });

  function syncThemeSelect() {
    var sel = document.getElementById('gjs-theme-select');
    if (!sel) return;
    sel.innerHTML = '';
    var none = document.createElement('option');
    none.value = '';
    none.textContent = 'None';
    sel.appendChild(none);
    (registeredThemes || []).forEach(function (theme) {
      var opt = document.createElement('option');
      opt.value = theme.id;
      opt.textContent = theme.label || theme.id;
      sel.appendChild(opt);
    });
    sel.value = currentThemeId || '';
  }

  syncThemeSelect();
  editor.on('load', function () {
    applyThemeToCanvas(editor, currentThemeId);
  });
  applyThemeToCanvas(editor, currentThemeId);

  var themeSelect = document.getElementById('gjs-theme-select');
  if (themeSelect) {
    themeSelect.addEventListener('change', function () {
      var next = themeSelect.value || '';
      themeSelect.disabled = true;
      fetch(themeURL, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ theme: next })
      }).then(function (res) {
        if (!res.ok) throw new Error('theme save failed');
        currentThemeId = next;
        applyThemeToCanvas(editor, currentThemeId);
      }).catch(function (err) {
        console.error('Failed to save theme', err);
        themeSelect.value = currentThemeId || '';
      }).finally(function () {
        themeSelect.disabled = false;
      });
    });
  }

  const saveBtn = document.getElementById('gjs-save-btn');
  const saveStatus = document.getElementById('gjs-save-status');
  function setSaveStatus(text) {
    if (saveStatus) saveStatus.textContent = text || '';
  }
  if (saveBtn) {
    saveBtn.addEventListener('click', function () {
      saveBtn.disabled = true;
      setSaveStatus('Saving…');
      Promise.resolve(editor.store())
        .then(function () { setSaveStatus('Saved'); })
        .catch(function (err) {
          console.error('GrapesJS store failed', err);
          setSaveStatus('Save failed');
        })
        .finally(function () { saveBtn.disabled = false; });
    });
  }

  window.__gjsEditor = editor;
})();
`, loadURL, storeURL, uploadURL, themeURL, blocksJSON, componentsJSON, traitsJSON, themesJSON, currentTheme, dotLottieCDNURL, dotLottieScriptAttr)

	_, err = fmt.Fprintf(w, `
<div class="gjs-builder-wrap">
  <div class="gjs-builder-bar">
    <a class="btn btn-sm btn-outline" href="%s" hx-boost="false">← Back to route</a>
    <span class="text-sm opacity-70">Editing %s</span>
    <div class="gjs-builder-bar-actions">
      <label class="text-sm opacity-80" for="gjs-theme-select">Theme</label>
      <select id="gjs-theme-select" class="select select-sm select-bordered bg-transparent text-white border-white/30"></select>
      <span id="gjs-save-status" class="gjs-builder-save-status" aria-live="polite"></span>
      <button type="button" id="gjs-save-btn" class="btn btn-sm btn-primary">Save</button>
    </div>
  </div>
  <div id="gjs"></div>
</div>
<script>%s</script>
`, html.EscapeString(detailURL), html.EscapeString(route.Path), script)
	return err
}
