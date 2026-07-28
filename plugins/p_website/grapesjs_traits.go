package p_website

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

func pluginGrapesJSTraits() lariv.PluginFeatures[lariv.GrapesJSTrait] {
	return lariv.PluginFeatures[lariv.GrapesJSTrait]{
		Entries: []registry.Pair[string, lariv.GrapesJSTrait]{
			{
				Key: "p_website.src-url",
				Value: lariv.GrapesJSTrait{
					EventCapture: []string{"input", "change"},
					CreateInput: lariv.GrapesJSTraitCreateInputJS(`
const el = document.createElement('input');
el.type = 'url';
el.placeholder = (trait && trait.get && trait.get('placeholder')) || 'https://…';
el.style.width = '100%';
return el;
`),
					OnEvent: lariv.GrapesJSTraitOnEventJS(`
const name = (trait && trait.get && trait.get('name')) || 'src';
const value = elInput && elInput.value != null ? elInput.value : '';
if (trait && trait.get && trait.get('changeProp')) {
  component.set(name, value);
} else {
  const attrs = {};
  attrs[name] = value;
  component.addAttributes(attrs);
}
`),
					OnUpdate: lariv.GrapesJSTraitOnUpdateJS(`
const name = (trait && trait.get && trait.get('name')) || 'src';
let value = '';
if (trait && trait.get && trait.get('changeProp')) {
  value = component.get(name) || '';
} else {
  const attrs = component.getAttributes() || {};
  value = attrs[name] || '';
}
if (elInput) elInput.value = value;
`),
				},
			},
		},
	}
}
