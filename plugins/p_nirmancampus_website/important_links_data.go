package p_nirmancampus_website

import (
	"context"
	"log/slog"
	"strings"

	"gorm.io/gorm"
)

type importantLinkHomeItem struct {
	Title string
	URL   string
}

func buildImportantLinksHomeItems(ctx context.Context, db *gorm.DB) []importantLinkHomeItem {
	links, err := gorm.G[ImportantLink](db).Order(`"order" ASC`).Find(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: failed loading important links", "error", err)
		return nil
	}

	items := make([]importantLinkHomeItem, 0, len(links))
	for _, l := range links {
		title := strings.TrimSpace(l.Title)
		if title == "" {
			continue
		}
		url := strings.TrimSpace(ImportantLinkPublicURL(l))
		if url == "" {
			continue
		}
		items = append(items, importantLinkHomeItem{Title: title, URL: url})
	}
	return items
}
