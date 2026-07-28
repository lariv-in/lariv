package p_website

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

//go:embed grapesjs_components/*
var grapesJSComponentsFS embed.FS

func mustReadGrapesJSComponentHTML(name string) string {
	b, err := grapesJSComponentsFS.ReadFile("grapesjs_components/" + name)
	if err != nil {
		log.Panicf("p_website: read grapesjs component %q: %v", name, err)
	}
	return strings.TrimSpace(string(b))
}

func grapesJSIsComponentType(typeID string) json.RawMessage {
	return lariv.GrapesJSComponentIsComponentJS(fmt.Sprintf(
		`return !!(el && el.getAttribute && el.getAttribute('data-gjs-type') === %q);`, typeID,
	))
}

func grapesJSSrcURLTrait(name, label string) map[string]any {
	return map[string]any{
		"type":  "p_website.src-url",
		"name":  name,
		"label": label,
	}
}

func grapesJSComponentEntry(typeID string, model map[string]any, view map[string]any) registry.Pair[string, lariv.GrapesJSComponent] {
	comp := lariv.GrapesJSComponent{
		IsComponent: grapesJSIsComponentType(typeID),
		Model:       lariv.GrapesJSComponentMustModelJSON(model),
	}
	if view != nil {
		comp.View = lariv.GrapesJSComponentMustViewJSON(view)
	}
	return registry.Pair[string, lariv.GrapesJSComponent]{Key: typeID, Value: comp}
}

func counterAttrsTraits() []any {
	return []any{
		map[string]any{"type": "number", "name": "data-value", "label": "Value"},
		map[string]any{"type": "number", "name": "data-duration", "label": "Duration (ms)"},
		map[string]any{"type": "text", "name": "data-suffix", "label": "Suffix"},
	}
}

const accordionScript = `
var root = this;
root.addEventListener('click', function (ev) {
  var btn = ev.target.closest('.gjs-accordion-trigger');
  if (!btn || !root.contains(btn)) return;
  var item = btn.closest('.gjs-accordion-item');
  if (!item) return;
  var panel = item.querySelector('.gjs-accordion-panel');
  if (!panel) return;
  var open = !panel.hasAttribute('hidden');
  root.querySelectorAll('.gjs-accordion-panel').forEach(function (p) { p.setAttribute('hidden', ''); });
  if (!open) panel.removeAttribute('hidden');
});
`

const dropdownScript = `
var root = this;
var trigger = root.querySelector('.gjs-dropdown-trigger');
var menu = root.querySelector('.gjs-dropdown-menu');
if (trigger && menu) {
  trigger.addEventListener('click', function () {
    if (menu.hasAttribute('hidden')) menu.removeAttribute('hidden');
    else menu.setAttribute('hidden', '');
  });
}
`

const sliderScript = `
var root = this;
var slides = Array.prototype.slice.call(root.querySelectorAll('.gjs-slider-slide'));
var idx = Math.max(0, slides.findIndex(function (s) { return s.classList.contains('is-active'); }));
function show(i) {
  if (!slides.length) return;
  idx = (i + slides.length) % slides.length;
  slides.forEach(function (s, n) {
    if (n === idx) { s.classList.add('is-active'); s.removeAttribute('hidden'); }
    else { s.classList.remove('is-active'); s.setAttribute('hidden', ''); }
  });
}
show(idx < 0 ? 0 : idx);
var prev = root.querySelector('.gjs-slider-prev');
var next = root.querySelector('.gjs-slider-next');
if (prev) prev.addEventListener('click', function () { show(idx - 1); });
if (next) next.addEventListener('click', function () { show(idx + 1); });
`

const tabsScript = `
var root = this;
var tabs = Array.prototype.slice.call(root.querySelectorAll('.gjs-tab'));
var panels = Array.prototype.slice.call(root.querySelectorAll('.gjs-tab-panel'));
function activate(i) {
  tabs.forEach(function (t, n) { t.classList.toggle('is-active', n === i); });
  panels.forEach(function (p, n) {
    if (n === i) { p.classList.add('is-active'); p.removeAttribute('hidden'); }
    else { p.classList.remove('is-active'); p.setAttribute('hidden', ''); }
  });
}
tabs.forEach(function (tab, i) {
  tab.addEventListener('click', function () { activate(i); });
});
`

const toggleableScript = `
var root = this;
var trigger = root.querySelector('.gjs-toggleable-trigger');
var panel = root.querySelector('.gjs-toggleable-panel');
if (trigger && panel) {
  trigger.addEventListener('click', function () {
    if (panel.hasAttribute('hidden')) panel.removeAttribute('hidden');
    else panel.setAttribute('hidden', '');
  });
}
`

const numberCounterScript = `
var root = this;
var target = parseFloat(root.getAttribute('data-value') || '0') || 0;
var duration = parseInt(root.getAttribute('data-duration') || '1500', 10) || 1500;
var suffix = root.getAttribute('data-suffix') || '';
var el = root.querySelector('.gjs-number-counter-value');
if (!el) return;
var start = null;
function frame(ts) {
  if (start === null) start = ts;
  var p = Math.min(1, (ts - start) / duration);
  el.textContent = Math.round(target * p) + suffix;
  if (p < 1) requestAnimationFrame(frame);
}
requestAnimationFrame(frame);
`

const barCounterScript = `
var root = this;
var target = parseFloat(root.getAttribute('data-value') || '0') || 0;
var duration = parseInt(root.getAttribute('data-duration') || '1200', 10) || 1200;
var suffix = root.getAttribute('data-suffix') || '';
var fill = root.querySelector('.gjs-bar-counter-fill');
var valueEl = root.querySelector('.gjs-bar-counter-value');
var start = null;
function frame(ts) {
  if (start === null) start = ts;
  var p = Math.min(1, (ts - start) / duration);
  var v = Math.round(target * p);
  if (fill) fill.style.width = v + '%';
  if (valueEl) valueEl.textContent = v + suffix;
  if (p < 1) requestAnimationFrame(frame);
}
requestAnimationFrame(frame);
`

const circleCounterScript = `
var root = this;
var target = parseFloat(root.getAttribute('data-value') || '0') || 0;
var duration = parseInt(root.getAttribute('data-duration') || '1200', 10) || 1200;
var suffix = root.getAttribute('data-suffix') || '';
var fg = root.querySelector('.gjs-circle-counter-fg');
var valueEl = root.querySelector('.gjs-circle-counter-value');
var start = null;
function frame(ts) {
  if (start === null) start = ts;
  var p = Math.min(1, (ts - start) / duration);
  var v = Math.round(target * p);
  if (fg) fg.setAttribute('stroke-dasharray', v + ', 100');
  if (valueEl) valueEl.textContent = v + suffix;
  if (p < 1) requestAnimationFrame(frame);
}
requestAnimationFrame(frame);
`

const countdownScript = `
var root = this;
var targetAttr = root.getAttribute('data-target');
var target = targetAttr ? new Date(targetAttr).getTime() : (Date.now() + 7 * 24 * 60 * 60 * 1000);
var daysEl = root.querySelector('.gjs-countdown-days');
var hoursEl = root.querySelector('.gjs-countdown-hours');
var minsEl = root.querySelector('.gjs-countdown-mins');
var secsEl = root.querySelector('.gjs-countdown-secs');
function pad(n) { return (n < 10 ? '0' : '') + n; }
function tick() {
  var diff = Math.max(0, target - Date.now());
  var s = Math.floor(diff / 1000);
  var days = Math.floor(s / 86400); s -= days * 86400;
  var hours = Math.floor(s / 3600); s -= hours * 3600;
  var mins = Math.floor(s / 60); s -= mins * 60;
  if (daysEl) daysEl.textContent = pad(days);
  if (hoursEl) hoursEl.textContent = pad(hours);
  if (minsEl) minsEl.textContent = pad(mins);
  if (secsEl) secsEl.textContent = pad(s);
}
tick();
setInterval(tick, 1000);
`

const headingInit = `
this.on('change:attributes:data-level', function () {
  var level = this.getAttributes()['data-level'] || '2';
  var tag = 'h' + level;
  if (this.get('tagName') !== tag) this.set('tagName', tag);
});
`

func pluginGrapesJSComponents() lariv.PluginFeatures[lariv.GrapesJSComponent] {
	return lariv.PluginFeatures[lariv.GrapesJSComponent]{
		Entries: []registry.Pair[string, lariv.GrapesJSComponent]{
			grapesJSComponentEntry("p_website.accordion", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"droppable":  false,
					"attributes": map[string]any{"data-gjs-type": "p_website.accordion", "class": "gjs-accordion"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
					"script":     accordionScript,
				},
			}, nil),
			grapesJSComponentEntry("p_website.blurb", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.blurb", "class": "gjs-blurb"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.button", map[string]any{
				"defaults": map[string]any{
					"tagName":    "a",
					"editable":   true,
					"attributes": map[string]any{"data-gjs-type": "p_website.button", "class": "gjs-button", "href": "#"},
					"traits": []any{
						map[string]any{"type": "text", "name": "href", "label": "URL"},
						map[string]any{"type": "text", "name": "title", "label": "Title"},
						map[string]any{"type": "select", "name": "target", "label": "Target", "options": []any{
							map[string]any{"id": "", "label": "Same tab"},
							map[string]any{"id": "_blank", "label": "New tab"},
						}},
					},
				},
			}, nil),
			grapesJSComponentEntry("p_website.cta", map[string]any{
				"defaults": map[string]any{
					"tagName":    "section",
					"attributes": map[string]any{"data-gjs-type": "p_website.cta", "class": "gjs-cta"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.code", map[string]any{
				"defaults": map[string]any{
					"tagName":    "pre",
					"attributes": map[string]any{"data-gjs-type": "p_website.code", "class": "gjs-code"},
					"traits": []any{
						map[string]any{"type": "text", "name": "data-language", "label": "Language"},
					},
				},
			}, nil),
			grapesJSComponentEntry("p_website.divider", map[string]any{
				"defaults": map[string]any{
					"tagName":    "hr",
					"void":       true,
					"droppable":  false,
					"attributes": map[string]any{"data-gjs-type": "p_website.divider", "class": "gjs-divider"},
					"traits":     []any{},
				},
			}, nil),
			grapesJSComponentEntry("p_website.dropdown", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.dropdown", "class": "gjs-dropdown"},
					"script":     dropdownScript,
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.gallery", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"droppable":  true,
					"attributes": map[string]any{"data-gjs-type": "p_website.gallery", "class": "gjs-gallery"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			{
				Key: "p_website.heading",
				Value: lariv.GrapesJSComponent{
					Extend:      "text",
					IsComponent: grapesJSIsComponentType("p_website.heading"),
					Model: lariv.GrapesJSComponentMustModelJSON(map[string]any{
						"defaults": map[string]any{
							"tagName":  "h2",
							"editable": true,
							"attributes": map[string]any{
								"data-gjs-type": "p_website.heading",
								"class":         "gjs-heading",
								"data-level":    "2",
							},
							"traits": []any{
								map[string]any{"type": "select", "name": "data-level", "label": "Level", "options": []any{
									map[string]any{"id": "1", "label": "H1"},
									map[string]any{"id": "2", "label": "H2"},
									map[string]any{"id": "3", "label": "H3"},
									map[string]any{"id": "4", "label": "H4"},
									map[string]any{"id": "5", "label": "H5"},
									map[string]any{"id": "6", "label": "H6"},
								}},
							},
						},
						"init": headingInit,
					}),
				},
			},
			grapesJSComponentEntry("p_website.hero", map[string]any{
				"defaults": map[string]any{
					"tagName":    "section",
					"attributes": map[string]any{"data-gjs-type": "p_website.hero", "class": "gjs-hero"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.icon", map[string]any{
				"defaults": map[string]any{
					"tagName":    "span",
					"droppable":  false,
					"attributes": map[string]any{"data-gjs-type": "p_website.icon", "class": "gjs-icon"},
					"traits":     []any{map[string]any{"type": "text", "name": "class", "label": "CSS class"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.icon-list", map[string]any{
				"defaults": map[string]any{
					"tagName":    "ul",
					"droppable":  true,
					"attributes": map[string]any{"data-gjs-type": "p_website.icon-list", "class": "gjs-icon-list"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			{
				Key: "p_website.image",
				Value: lariv.GrapesJSComponent{
					Extend:      "image",
					IsComponent: grapesJSIsComponentType("p_website.image"),
					Model: lariv.GrapesJSComponentMustModelJSON(map[string]any{
						"defaults": map[string]any{
							"attributes": map[string]any{"data-gjs-type": "p_website.image", "class": "gjs-image"},
							"traits": []any{
								grapesJSSrcURLTrait("src", "Source"),
								map[string]any{"type": "text", "name": "alt", "label": "Alt"},
								map[string]any{"type": "text", "name": "title", "label": "Title"},
							},
						},
					}),
				},
			},
			grapesJSComponentEntry("p_website.link", map[string]any{
				"defaults": map[string]any{
					"tagName":    "a",
					"editable":   true,
					"attributes": map[string]any{"data-gjs-type": "p_website.link", "class": "gjs-link", "href": "#"},
					"traits": []any{
						map[string]any{"type": "text", "name": "href", "label": "URL"},
						map[string]any{"type": "text", "name": "title", "label": "Title"},
						map[string]any{"type": "select", "name": "target", "label": "Target", "options": []any{
							map[string]any{"id": "", "label": "Same tab"},
							map[string]any{"id": "_blank", "label": "New tab"},
						}},
					},
				},
			}, nil),
			{
				Key: "p_website.dotlottie",
				Value: lariv.GrapesJSComponent{
					IsComponent: lariv.GrapesJSComponentIsComponentJS(
						`return !!(el && el.tagName && el.tagName.toLowerCase() === 'dotlottie-wc');`,
					),
					Model: lariv.GrapesJSComponentMustModelJSON(map[string]any{
						"defaults": map[string]any{
							"tagName":   "dotlottie-wc",
							"void":      false,
							"droppable": false,
							"attributes": map[string]any{
								"data-gjs-type": "p_website.dotlottie",
								"class":         "gjs-dotlottie",
								"src":           "https://lottie.host/4db68bbd-31f6-4cd8-84eb-189de081159a/IGmMCqhzpt.lottie",
								"autoplay":      "",
								"loop":          "",
								"style":         "width: 300px; height: 300px;",
							},
							"traits": []any{
								grapesJSSrcURLTrait("src", "Animation URL"),
								map[string]any{"type": "checkbox", "name": "autoplay", "label": "Autoplay", "valueTrue": "", "valueFalse": "false"},
								map[string]any{"type": "checkbox", "name": "loop", "label": "Loop", "valueTrue": "", "valueFalse": "false"},
								map[string]any{"type": "text", "name": "style", "label": "Style"},
							},
						},
					}),
					View: lariv.GrapesJSComponentMustViewJSON(map[string]any{
						"onRender": `
var doc = (this.el && this.el.ownerDocument) || document;
if (window.__larivEnsureDotLottie) {
  window.__larivEnsureDotLottie(doc);
}
`,
					}),
				},
			},
			grapesJSComponentEntry("p_website.person", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.person", "class": "gjs-person"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.pricing-tables", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"droppable":  true,
					"attributes": map[string]any{"data-gjs-type": "p_website.pricing-tables", "class": "gjs-pricing"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.slider", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.slider", "class": "gjs-slider"},
					"script":     sliderScript,
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.tabs", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.tabs", "class": "gjs-tabs"},
					"script":     tabsScript,
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.testimonial", map[string]any{
				"defaults": map[string]any{
					"tagName":    "blockquote",
					"attributes": map[string]any{"data-gjs-type": "p_website.testimonial", "class": "gjs-testimonial"},
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			grapesJSComponentEntry("p_website.toggleable", map[string]any{
				"defaults": map[string]any{
					"tagName":    "div",
					"attributes": map[string]any{"data-gjs-type": "p_website.toggleable", "class": "gjs-toggleable"},
					"script":     toggleableScript,
					"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
				},
			}, nil),
			{
				Key: "p_website.text",
				Value: lariv.GrapesJSComponent{
					Extend:      "text",
					IsComponent: grapesJSIsComponentType("p_website.text"),
					Model: lariv.GrapesJSComponentMustModelJSON(map[string]any{
						"defaults": map[string]any{
							"tagName":    "p",
							"editable":   true,
							"attributes": map[string]any{"data-gjs-type": "p_website.text", "class": "gjs-text"},
							"traits":     []any{map[string]any{"type": "text", "name": "id", "label": "ID"}},
						},
					}),
				},
			},
			grapesJSComponentEntry("p_website.bar-counter", map[string]any{
				"defaults": map[string]any{
					"tagName": "div",
					"attributes": map[string]any{
						"data-gjs-type":  "p_website.bar-counter",
						"class":          "gjs-bar-counter",
						"data-value":     "75",
						"data-duration":  "1200",
						"data-suffix":    "%",
					},
					"traits": counterAttrsTraits(),
					"script": barCounterScript,
				},
			}, nil),
			grapesJSComponentEntry("p_website.circle-counter", map[string]any{
				"defaults": map[string]any{
					"tagName": "div",
					"attributes": map[string]any{
						"data-gjs-type": "p_website.circle-counter",
						"class":         "gjs-circle-counter",
						"data-value":    "80",
						"data-duration": "1200",
						"data-suffix":   "%",
					},
					"traits": counterAttrsTraits(),
					"script": circleCounterScript,
				},
			}, nil),
			grapesJSComponentEntry("p_website.countdown-counter", map[string]any{
				"defaults": map[string]any{
					"tagName": "div",
					"attributes": map[string]any{
						"data-gjs-type": "p_website.countdown-counter",
						"class":         "gjs-countdown",
						"data-target":   "",
					},
					"traits": []any{
						map[string]any{"type": "text", "name": "data-target", "label": "Target (ISO date)"},
					},
					"script": countdownScript,
				},
			}, nil),
			grapesJSComponentEntry("p_website.number-counter", map[string]any{
				"defaults": map[string]any{
					"tagName": "div",
					"attributes": map[string]any{
						"data-gjs-type": "p_website.number-counter",
						"class":         "gjs-number-counter",
						"data-value":    "1000",
						"data-duration": "1500",
						"data-suffix":   "+",
					},
					"traits": counterAttrsTraits(),
					"script": numberCounterScript,
				},
			}, nil),
		},
	}
}
