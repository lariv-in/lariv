package lago

import (
	"log"
	"net/http"

	"github.com/lariv-in/views"
)

func GetPageView(pageName string) views.View {
	_, pageExists := RegistryPage.Get(pageName)
	if !pageExists {
		log.Panicf("Tried to access page: %s, which does not exist in the template registry at this time", pageName)
	}
	return views.View{
		PageName: pageName,
		Registry: RegistryPage.All(),
		Handlers: map[string]func(v views.View) http.Handler{
			http.MethodGet: func(v views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					v.RenderPage(w, r)
				})
			},
		},
	}
}
