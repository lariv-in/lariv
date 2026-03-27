package lago

import (
	"log"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/views"
)

func GetPageView(pageName string) *views.View {
	_, pageExists := RegistryPage.Get(pageName)
	if !pageExists {
		log.Panicf("Tried to access page: %s, which does not exist in the template registry at this time", pageName)
	}
	return &views.View{
		PageName: pageName,
		PageLookup: func(name string) (components.PageInterface, bool) {
			return RegistryPage.Get(name)
		},
		Handlers: map[string]func(v *views.View) http.Handler{
			http.MethodGet: func(v *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					v.ServeRenderPage(w, r)
				})
			},
		},
	}
}
