package forms

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerPublicPage() {
	lago.RegistryPage.Register("forms.PublicSubmitPage", &components.ShellBase{
		Page: components.Page{Key: "forms.PublicSubmitPage"},
		Children: []components.PageInterface{
			&components.LayoutCard{
				Page: components.Page{Key: "forms.PublicSubmitCard"},
				Children: []components.PageInterface{
					&PublicSubmitForm{
						Page:      components.Page{Key: "forms.PublicSubmitFormBody"},
						ActionURL: lago.RoutePath("forms.PublicFormRoute", map[string]getters.Getter[any]{"slug": getters.Any(getters.Key[string](ContextKeyPublicLoadedForm + ".Slug"))}),
					},
				},
			},
		},
	})
}
