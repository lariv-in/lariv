package p_website

import (
	"embed"
	"log"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

//go:embed grapesjs_blocks/*
var grapesJSBlocksFS embed.FS

func mustReadGrapesJSBlockHTML(name string) string {
	b, err := grapesJSBlocksFS.ReadFile("grapesjs_blocks/" + name)
	if err != nil {
		log.Panicf("p_website: read grapesjs block %q: %v", name, err)
	}
	return strings.TrimSpace(string(b))
}

func grapesJSBlockFromComponentHTML(key, label, category, htmlFile string) registry.Pair[string, lariv.GrapesJSBlock] {
	return registry.Pair[string, lariv.GrapesJSBlock]{
		Key: key,
		Value: lariv.GrapesJSBlock{
			Label:    label,
			Category: lariv.GrapesJSBlockCategoryString(category),
			Content:  lariv.GrapesJSBlockContentString(mustReadGrapesJSComponentHTML(htmlFile)),
		},
	}
}

func pluginGrapesJSBlocks() lariv.PluginFeatures[lariv.GrapesJSBlock] {
	layout := lariv.GrapesJSBlockCategoryString("Layout")
	return lariv.PluginFeatures[lariv.GrapesJSBlock]{
		Entries: []registry.Pair[string, lariv.GrapesJSBlock]{
			{
				Key: "p_website.section",
				Value: lariv.GrapesJSBlock{
					Label:    "Section",
					Category: layout,
					Content:  lariv.GrapesJSBlockContentString(mustReadGrapesJSBlockHTML("section.html")),
				},
			},
			{
				Key: "p_website.2-columns",
				Value: lariv.GrapesJSBlock{
					Label:    "2 Columns",
					Category: layout,
					Content:  lariv.GrapesJSBlockContentString(mustReadGrapesJSBlockHTML("2-columns.html")),
				},
			},
			{
				Key: "p_website.3-columns",
				Value: lariv.GrapesJSBlock{
					Label:    "3 Columns",
					Category: layout,
					Content:  lariv.GrapesJSBlockContentString(mustReadGrapesJSBlockHTML("3-columns.html")),
				},
			},
			{
				Key: "p_website.card",
				Value: lariv.GrapesJSBlock{
					Label:    "Card",
					Category: layout,
					Content:  lariv.GrapesJSBlockContentString(mustReadGrapesJSBlockHTML("card.html")),
				},
			},
			grapesJSBlockFromComponentHTML("p_website.accordion", "Accordion", "Interactive", "accordion.html"),
			grapesJSBlockFromComponentHTML("p_website.blurb", "Blurb", "Basic", "blurb.html"),
			grapesJSBlockFromComponentHTML("p_website.button", "Button", "Basic", "button.html"),
			grapesJSBlockFromComponentHTML("p_website.cta", "CTA", "Basic", "cta.html"),
			grapesJSBlockFromComponentHTML("p_website.code", "Code", "Basic", "code.html"),
			grapesJSBlockFromComponentHTML("p_website.divider", "Divider", "Basic", "divider.html"),
			grapesJSBlockFromComponentHTML("p_website.dropdown", "Dropdown", "Interactive", "dropdown.html"),
			grapesJSBlockFromComponentHTML("p_website.gallery", "Gallery", "Media", "gallery.html"),
			grapesJSBlockFromComponentHTML("p_website.heading", "Heading", "Basic", "heading.html"),
			grapesJSBlockFromComponentHTML("p_website.hero", "Hero", "Layout", "hero.html"),
			grapesJSBlockFromComponentHTML("p_website.icon", "Icon", "Media", "icon.html"),
			grapesJSBlockFromComponentHTML("p_website.icon-list", "Icon list", "Basic", "icon-list.html"),
			grapesJSBlockFromComponentHTML("p_website.image", "Image", "Media", "image.html"),
			grapesJSBlockFromComponentHTML("p_website.link", "Link", "Basic", "link.html"),
			grapesJSBlockFromComponentHTML("p_website.dotlottie", "DotLottie", "Media", "dotlottie.html"),
			grapesJSBlockFromComponentHTML("p_website.person", "Person", "Basic", "person.html"),
			grapesJSBlockFromComponentHTML("p_website.pricing-tables", "Pricing tables", "Basic", "pricing-tables.html"),
			grapesJSBlockFromComponentHTML("p_website.slider", "Slider", "Interactive", "slider.html"),
			grapesJSBlockFromComponentHTML("p_website.tabs", "Tabs", "Interactive", "tabs.html"),
			grapesJSBlockFromComponentHTML("p_website.testimonial", "Testimonial", "Basic", "testimonial.html"),
			grapesJSBlockFromComponentHTML("p_website.toggleable", "Toggleable", "Interactive", "toggleable.html"),
			grapesJSBlockFromComponentHTML("p_website.text", "Text", "Basic", "text.html"),
			grapesJSBlockFromComponentHTML("p_website.bar-counter", "Bar counter", "Counters", "bar-counter.html"),
			grapesJSBlockFromComponentHTML("p_website.circle-counter", "Circle counter", "Counters", "circle-counter.html"),
			grapesJSBlockFromComponentHTML("p_website.countdown-counter", "Countdown", "Counters", "countdown-counter.html"),
			grapesJSBlockFromComponentHTML("p_website.number-counter", "Number counter", "Counters", "number-counter.html"),
		},
	}
}
