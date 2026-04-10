package p_lacerate

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	envRodFallback = "LAGO_lacerate_ROD_FALLBACK"
	envChromeBin   = "LAGO_lacerate_CHROME_BIN"

	rodNavigateTimeout = 50 * time.Second
)

func rodFallbackEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(envRodFallback)))
	return v == "1" || v == "true" || v == "yes"
}

var (
	rodBrowserMu sync.Mutex
	rodBrowser   *rod.Browser
)

// getRodBrowser lazily launches one headless Chromium for the process (shared across link fetches).
func getRodBrowser(ctx context.Context) (*rod.Browser, error) {
	rodBrowserMu.Lock()
	defer rodBrowserMu.Unlock()
	if rodBrowser != nil {
		return rodBrowser, nil
	}

	launchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	l := launcher.New().Context(launchCtx).Leakless(true).Headless(true)
	if bin := strings.TrimSpace(os.Getenv(envChromeBin)); bin != "" {
		l = l.Bin(bin)
	}
	ws, err := l.Launch()
	if err != nil {
		slog.Error("lacerate: rod launch", "error", err)
		return nil, err
	}
	b := rod.New().Context(ctx).ControlURL(ws)
	if err := b.Connect(); err != nil {
		slog.Error("lacerate: rod connect", "error", err)
		return nil, err
	}
	rodBrowser = b
	slog.Info("lacerate: rod browser started for link extraction fallback")
	return rodBrowser, nil
}

// fetchHTMLViaRod loads pageURL in headless Chrome and returns the rendered document HTML.
func fetchHTMLViaRod(ctx context.Context, pageURL string) (string, error) {
	b, err := getRodBrowser(ctx)
	if err != nil {
		slog.Error("lacerate: rod get browser", "error", err)
		return "", err
	}
	rodCtx, cancel := context.WithTimeout(ctx, rodNavigateTimeout)
	defer cancel()

	page, err := b.Page(proto.TargetCreateTarget{})
	if err != nil {
		slog.Error("lacerate: rod new page", "error", err)
		return "", err
	}
	defer func() {
		if err := page.Close(); err != nil {
			slog.Warn("lacerate: rod page close", "error", err)
		}
	}()

	p := page.Context(rodCtx)
	if err := p.Navigate(pageURL); err != nil {
		slog.Error("lacerate: rod navigate", "error", err, "url", pageURL)
		return "", err
	}
	if err := p.WaitStable(500 * time.Millisecond); err != nil {
		slog.Error("lacerate: rod wait stable", "error", err, "url", pageURL)
		return "", err
	}
	html, err := p.HTML()
	if err != nil {
		slog.Error("lacerate: rod html", "error", err, "url", pageURL)
		return "", err
	}
	return html, nil
}
