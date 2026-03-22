package p_nirmancampus_website

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_announcements"
	"github.com/lariv-in/lago/p_filesystem"
	"gorm.io/gorm"
)

const (
	homeAnnouncementLimit  = 10
	popupImagesDirectory   = "/popup_images"
	popupImageNotFoundHint = "path not found:"
)

type homePageData struct {
	Announcements []homeAnnouncement
	PopupImageURL string
	HasPopup      bool
}

type homeAnnouncement struct {
	Text string
}

func buildHomePageData(ctx context.Context) homePageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building homepage", "error", err)
		return homePageData{}
	}

	data := homePageData{
		Announcements: loadHomeAnnouncements(db, time.Now()),
	}

	popupImageURL, err := loadRandomPopupImageURL(ctx, db)
	if err != nil {
		slog.Error("nirmancampus_website: failed loading popup image", "error", err)
		return data
	}

	data.PopupImageURL = popupImageURL
	data.HasPopup = popupImageURL != ""
	return data
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
	for _, announcement := range announcements {
		text := homeAnnouncementText(announcement)
		if text == "" {
			continue
		}
		items = append(items, homeAnnouncement{Text: text})
	}
	return items
}

func homeAnnouncementText(announcement p_announcements.Announcement) string {
	title := strings.TrimSpace(announcement.Title)
	description := strings.TrimSpace(announcement.Description)

	switch {
	case title != "" && description != "":
		return title + ": " + description
	case description != "":
		return description
	default:
		return title
	}
}

func loadRandomPopupImageURL(ctx context.Context, db *gorm.DB) (string, error) {
	popupDirectory, err := popupImagesDirectoryNode(db)
	if err != nil || popupDirectory == nil {
		return "", err
	}

	var candidates []p_filesystem.VNode
	if err := p_filesystem.ListChildrenForParent(db, &popupDirectory.ID).
		Where("is_directory = ?", false).
		Find(&candidates).Error; err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", nil
	}

	return popupImageURL(ctx, candidates[rand.Intn(len(candidates))].ID)
}

func popupImageURL(ctx context.Context, id uint) (string, error) {
	return lago.GetterRoutePath("nirmancampus_website.PopupImageRoute", map[string]getters.Getter[any]{
		"id": getters.GetterAny(getters.GetterStatic(id)),
	})(ctx)
}

func popupImagesDirectoryNode(db *gorm.DB) (*p_filesystem.VNode, error) {
	node, _, err := p_filesystem.GetVNodeByPath(db, popupImagesDirectory)
	if err != nil {
		if strings.Contains(err.Error(), popupImageNotFoundHint) {
			return nil, nil
		}
		return nil, err
	}
	if node == nil || !node.IsDirectory {
		return nil, nil
	}
	return node, nil
}

func loadPublicPopupImageNodeByID(db *gorm.DB, id uint) (*p_filesystem.VNode, error) {
	node, err := p_filesystem.GetVNodeByID(db, id)
	if err != nil {
		return nil, err
	}
	if node.IsDirectory {
		return nil, gorm.ErrRecordNotFound
	}

	popupDirectory, err := popupImagesDirectoryNode(db)
	if err != nil {
		return nil, err
	}
	if popupDirectory == nil {
		return nil, gorm.ErrRecordNotFound
	}

	allowed, err := node.IsDescendantOf(db, popupDirectory.ID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, gorm.ErrRecordNotFound
	}

	return node, nil
}
