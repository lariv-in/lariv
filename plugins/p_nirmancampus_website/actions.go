package p_nirmancampus_website

import (
	"context"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_announcements"
	"gorm.io/gorm"
)

type homePageData struct {
	Announcements  []homeAnnouncement
	ImportantLinks []importantLinkHomeItem
}

type homeAnnouncement struct {
	Title       string
	Description template.HTML
	Date        string
	URL         string
}

// ImportantLinkItemURLPrefix is the path prefix for ImportantLinkItemRoute (public file/link handler).
const ImportantLinkItemURLPrefix = "/important-links/item/"

// ImportantLinkPublicURL returns the public href for a row (trimmed external link, or item path for downloads).
func ImportantLinkPublicURL(l ImportantLink) string {
	if l.IsLink {
		return strings.TrimSpace(l.Link)
	}
	return fmt.Sprintf("%s%d/", ImportantLinkItemURLPrefix, l.ID)
}

func buildHomePageData(ctx context.Context) homePageData {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building home page", "error", err)
		return homePageData{}
	}

	return homePageData{
		Announcements:  loadHomeAnnouncements(ctx, db, time.Now()),
		ImportantLinks: buildImportantLinksHomeItems(ctx, db),
	}
}

func loadHomeAnnouncements(ctx context.Context, db *gorm.DB, now time.Time) []homeAnnouncement {
	announcements, err := gorm.G[p_nirmancampus_announcements.Announcement](db).
		Where("release_at <= ?", now).
		Where("expiry_at IS NULL OR expiry_at > ?", now).
		Order("release_at DESC").
		Find(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: failed loading announcements", "error", err)
		return nil
	}

	items := make([]homeAnnouncement, 0, len(announcements))
	for _, a := range announcements {
		title := strings.TrimSpace(a.Title)
		if title == "" {
			continue
		}
		desc := strings.TrimSpace(a.Description)
		items = append(items, homeAnnouncement{
			Title:       title,
			Description: template.HTML(components.RenderMarkdown(desc)),
			Date:        a.ReleaseAt.Format("Jan 2, 2006"),
			URL:         strings.TrimSpace(a.URL),
		})
	}
	return items
}
