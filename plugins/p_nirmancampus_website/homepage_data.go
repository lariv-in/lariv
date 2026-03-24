package p_nirmancampus_website

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/p_announcements"
	"gorm.io/gorm"
)

const homeAnnouncementLimit = 10

type homePageData struct {
	Announcements []homeAnnouncement
}

type homeAnnouncement struct {
	Title       string
	Description template.HTML
	Date        string
}

func buildHomePageData(ctx context.Context) homePageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building homepage", "error", err)
		return homePageData{}
	}

	return homePageData{
		Announcements: loadHomeAnnouncements(db, time.Now()),
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
	var announcements []p_announcements.Announcement
	if err := db.Model(&p_announcements.Announcement{}).
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
		})
	}
	return items
}
