package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

// EnvChromeBin optional explicit Chromium/Chrome binary (same idea as p_lacerate).
const EnvChromeBin = "LAGO_seer_websites_CHROME_BIN"

const rodNavigateTimeout = 120 * time.Second

var (
	rodBrowserMu sync.Mutex
	rodBrowser   *rod.Browser
)

var errSSRFAfterRedirect = errors.New("p_seer_websites: redirect landed on non-public host")

func getRodBrowser() (*rod.Browser, error) {
	rodBrowserMu.Lock()
	defer rodBrowserMu.Unlock()
	if rodBrowser != nil {
		return rodBrowser, nil
	}

	// Launcher and [rod.Browser] outlive a single HTTP request. Do not cancel a child context when this
	// function returns — that tears down Chrome/CDP. Do not use [http.Request.Context] on the singleton
	// [rod.Browser] or it becomes unusable after the first POST returns.
	l := launcher.New().Context(context.Background()).Leakless(true).Headless(true)
	if bin := strings.TrimSpace(os.Getenv(EnvChromeBin)); bin != "" {
		l = l.Bin(bin)
	}
	if dir := strings.TrimSpace(WebsiteRod.UserDataDir); dir != "" {
		l = l.UserDataDir(dir)
		if prof := strings.TrimSpace(WebsiteRod.ProfileDir); prof != "" {
			l = l.ProfileDir(prof)
		}
	}
	ws, err := l.Launch()
	if err != nil {
		return nil, err
	}
	b := rod.New().Context(context.Background()).ControlURL(ws)
	if err := b.Connect(); err != nil {
		return nil, err
	}
	rodBrowser = b
	slog.Info("p_seer_websites: rod browser started")
	return rodBrowser, nil
}

// fetchRenderedHTML loads pageURL in headless Chromium and returns document HTML and final URL after navigation.
func fetchRenderedHTML(ctx context.Context, pageURL *url.URL) (html string, final *url.URL, err error) {
	if pageURL == nil {
		return "", nil, fmt.Errorf("page url is nil")
	}
	b, err := getRodBrowser()
	if err != nil {
		return "", nil, err
	}
	rodCtx, cancel := context.WithTimeout(context.Background(), rodNavigateTimeout)
	defer cancel()

	page, err := b.Page(proto.TargetCreateTarget{})
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if cerr := page.Close(); cerr != nil {
			slog.Warn("p_seer_websites: rod page close", "error", cerr)
		}
	}()

	p := page.Context(rodCtx)
	if err := p.Emulate(devices.IPhoneX); err != nil {
		return "", nil, fmt.Errorf("emulate device: %w", err)
	}
	pageURLStr := pageURL.String()
	if err := p.Navigate(pageURLStr); err != nil {
		return "", nil, err
	}
	if err := p.WaitStable(500 * time.Millisecond); err != nil {
		slog.Warn("p_seer_websites: rod wait stable", "error", err, "url", pageURLStr)
	}
	info, err := p.Info()
	if err != nil {
		return "", nil, err
	}
	finalU, err := url.Parse(info.URL)
	if err != nil {
		return "", nil, err
	}
	if urlFailsSSRF(ctx, finalU) {
		return "", nil, errSSRFAfterRedirect
	}
	h, err := p.HTML()
	if err != nil {
		return "", nil, err
	}
	if len(h) > maxScrapedHTMLBytes {
		h = h[:maxScrapedHTMLBytes]
	}
	return h, finalU, nil
}

// ScrapeToMarkdown validates URL, fetches rendered HTML with rod, returns markdown and canonical URL.
func ScrapeToMarkdown(ctx context.Context, rawURL string) (markdown string, canonical *url.URL, err error) {
	canon, err := fetchableWebsiteURL(ctx, rawURL)
	if err != nil {
		return "", nil, err
	}
	htmlStr, finalU, err := fetchRenderedHTML(ctx, canon)
	if err != nil {
		return "", nil, fmt.Errorf("fetch page: %w", err)
	}
	md := markdownFromRenderedHTML(htmlStr, finalU)
	if strings.TrimSpace(md) == "" {
		return "", nil, fmt.Errorf("no extractable text from page (readability empty)")
	}
	out := cloneURL(canon)
	if finalU != nil {
		if norm, e := normalizeWebsiteURL(finalU.String()); e == nil {
			out = cloneURL(norm)
		}
	}
	return md, out, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	return new(*u)
}

// websiteScrapeFormPatcher fills [Website.Markdown] and normalizes [Website.URL] before [views.LayerCreate] persists.
type websiteScrapeFormPatcher struct{}

func (websiteScrapeFormPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if len(formErrors) > 0 {
		return formData, formErrors
	}
	raw, _ := formData["URL"].(string)
	md, canon, err := ScrapeToMarkdown(r.Context(), raw)
	if err != nil {
		formErrors["_form"] = err
		return formData, formErrors
	}
	var pp lago.PageURL
	pp.SetFromURL(canon)
	formData["URL"] = pp
	formData["Markdown"] = md
	return formData, formErrors
}

// websiteTitleHint returns a short title for intel rows from a page URL.
func websiteTitleHint(u *url.URL) string {
	if u == nil || u.Host == "" {
		return "Website"
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}
