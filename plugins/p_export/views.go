package p_export

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const exportCatalogContextKey = "export.catalog"

type catalogLayer struct{}

func (catalogLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			slog.Error("export: missing $db in catalog layer")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		catalog, err := BuildExportCatalog(db)
		if err != nil {
			slog.Error("export: build catalog", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), exportCatalogContextKey, catalog)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type methodGateLayer struct {
	Method string
}

func (m methodGateLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m.Method {
			http.Error(w, fmt.Sprintf("method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	lago.RegistryView.Register("export.PageView",
		lago.GetPageView("export.Page").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("export.catalog", catalogLayer{}))

	lago.RegistryView.Register("export.DownloadView",
		lago.GetPageView("export.Page").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("export.post_only", methodGateLayer{Method: http.MethodPost}).
			WithLayer("export.download", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: downloadHandler,
			}))
}
