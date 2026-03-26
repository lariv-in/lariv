package p_nirmancampus_website

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_announcements"
	"gorm.io/gorm"
)

const homeAnnouncementLimit = 10
const homeImportantLinksLimit = 6

type homePageData struct {
	Announcements   []homeAnnouncement
	ImportantLinks []homeImportantLink
}

type homeAnnouncement struct {
	Title       string
	Description template.HTML
	Date        string
	URL         string
}

type homeImportantLink struct {
	Title string
	URL   string
}

func buildHomePageData(ctx context.Context) homePageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building home page", "error", err)
		return homePageData{}
	}

	return homePageData{
		Announcements:   loadHomeAnnouncements(db, time.Now()),
		ImportantLinks: loadHomeImportantLinks(db),
	}
}

func homePageDB(ctx context.Context) (*gorm.DB, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("missing $db in context")
	}
	return db, nil
}

func loadHomeAnnouncements(db *gorm.DB, now time.Time) []homeAnnouncement {
	var announcements []p_nirmancampus_announcements.Announcement
	if err := db.Model(&p_nirmancampus_announcements.Announcement{}).
		Where("release_at <= ?", now).
		Where("expiry_at IS NULL OR expiry_at > ?", now).
		Order("release_at DESC").
		Limit(homeAnnouncementLimit).
		Find(&announcements).Error; err != nil {
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

func loadHomeImportantLinks(db *gorm.DB) []homeImportantLink {
	var links []ImportantLink
	if err := db.Order(`"order" ASC`).Limit(homeImportantLinksLimit).Find(&links).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading important links", "error", err)
		return nil
	}

	items := make([]homeImportantLink, 0, len(links))
	for _, l := range links {
		title := strings.TrimSpace(l.Title)
		if title == "" {
			continue
		}

		url := ""
		if l.IsLink {
			url = strings.TrimSpace(l.Link)
		} else {
			url = fmt.Sprintf("/important-links/item/%d/", l.ID)
		}

		if strings.TrimSpace(url) == "" {
			continue
		}

		items = append(items, homeImportantLink{
			Title: title,
			URL:   url,
		})
	}
	return items
}
