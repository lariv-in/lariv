package p_pwa

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	manifestViewKey      = "pwa.ManifestView"
	serviceWorkerViewKey = "pwa.ServiceWorkerView"
	offlineViewKey       = "pwa.OfflineView"
	staticPwaViewKey     = "pwa.StaticPwaView"
	pwaAssetPageName     = "pwa.AssetPlaceholder"
	assetLinksViewKey    = "pwa.assetLinksView"
)

type pwaAssetPage struct {
	components.Page
}

func (pwaAssetPage) Build(context.Context) Node {
	return Group{}
}

func (p pwaAssetPage) GetKey() string {
	return p.Page.Key
}

func (p pwaAssetPage) GetRoles() []string {
	return p.Page.Roles
}

func pwaAssetPageLookup(string) (components.PageInterface, bool) {
	return pwaAssetPage{Page: components.Page{Key: pwaAssetPageName}}, true
}

func pwaAssetView(method string, handler func(*views.View) http.Handler) *views.View {
	return &views.View{
		PageName:   pwaAssetPageName,
		PageLookup: pwaAssetPageLookup,
		Layers: []registry.Pair[string, views.Layer]{
			{Key: "pwa.asset", Value: views.MethodLayer{Method: method, Handler: handler}},
		},
	}
}

func manifestHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")

		manifest := map[string]any{
			"name":             Config.AppName,
			"description":      Config.AppDescription,
			"theme_color":      Config.AppThemeColor,
			"background_color": Config.AppBackgroundColor,
			"display":          Config.AppDisplay,
			"scope":            Config.AppScope,
			"orientation":      Config.AppOrientation,
			"start_url":        Config.AppStartURL,
			"dir":              Config.AppDir,
			"lang":             Config.AppLang,
			"icons":            Config.AppIcons,
			"shortcuts":        Config.AppShortcuts,
			"screenshots":      Config.AppScreenshots,

			"status_bar_color": Config.AppStatusBarColor,
			"icons_apple":      Config.AppIconsApple,
			"splash_screen":    Config.AppSplashScreen,
		}

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(manifest)
	})
}

func serviceWorkerHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")

		if Config.ServiceWorkerPath != "" {
			if _, err := os.Stat(Config.ServiceWorkerPath); err == nil {
				http.ServeFile(w, r, Config.ServiceWorkerPath)
				return
			}
			http.NotFound(w, r)
			return
		}

		w.Write([]byte(`/* lago p_pwa default service worker */
const CACHE_NAME = "lago-pwa-v1";
const OFFLINE_URL = "/offline";

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll([OFFLINE_URL]))
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener("fetch", (event) => {
  const req = event.request;
  if (req.method !== "GET") return;

  event.respondWith(
    fetch(req).catch(async () => {
      if (req.mode === "navigate") {
        const cache = await caches.open(CACHE_NAME);
        const cached = await cache.match(OFFLINE_URL);
        if (cached) return cached;
      }
      throw new Error("Network error");
    })
  );
});
`))
	})
}

func offlineHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Config.OfflineViewName != "" {
			lago.NewDynamicView(Config.OfflineViewName).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Offline</title>
  </head>
  <body>
    <h1>You're offline</h1>
    <p>Please check your connection and try again.</p>
  </body>
</html>`))
	})
}

func staticPwaHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Config.StaticDir == "" {
			slog.Warn("p_pwa: staticDir not configured; returning 404", "path", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		dir := Config.StaticDir
		if !filepath.IsAbs(dir) {
			exe, err := os.Executable()
			if err != nil {
				slog.Error("p_pwa: failed resolving executable path for staticDir", "err", err, "staticDir", dir, "path", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			dir = filepath.Join(filepath.Dir(exe), dir)
		}

		st, err := os.Stat(dir)
		if err != nil {
			slog.Error("p_pwa: staticDir does not exist or is not accessible", "err", err, "resolvedDir", dir, "path", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if !st.IsDir() {
			slog.Error("p_pwa: staticDir is not a directory", "resolvedDir", dir, "path", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		fs := http.FileServer(http.Dir(dir))
		http.StripPrefix("/static/pwa/", fs).ServeHTTP(w, r)
	})
}

func assetLinksHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		assetLinks := []map[string]any{}
		assetLinks = append(assetLinks, map[string]any{
			"relation": []string{
				"delegate_permission/common.handle_all_urls",
				"delegate_permission/common.get_login_creds",
			},
			"target": map[string]any{
				"namespace":    "android_app",
				"package_name": Config.AppPackageName,
				"sha256_cert_fingerprints": []string{
					Config.AppSHA256CertFingerprints,
				},
			},
		})

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(assetLinks)
	})
}

func init() {
	_ = components.RegistryShellHeadNodes.Register("pwa.manifestLink", Link(Rel("manifest"), Href("/app.webmanifest")))

	lago.RegistryView.Register(manifestViewKey, pwaAssetView(http.MethodGet, manifestHandler))
	lago.RegistryView.Register(serviceWorkerViewKey, pwaAssetView(http.MethodGet, serviceWorkerHandler))
	lago.RegistryView.Register(offlineViewKey, pwaAssetView(http.MethodGet, offlineHandler))
	lago.RegistryView.Register(staticPwaViewKey, pwaAssetView(http.MethodGet, staticPwaHandler))
	lago.RegistryView.Register(assetLinksViewKey, pwaAssetView(http.MethodGet, assetLinksHandler))
}
