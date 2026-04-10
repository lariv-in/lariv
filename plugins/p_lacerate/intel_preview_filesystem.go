package p_lacerate

import (
	"context"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

func normalizePreviewFetchURL(raw string) (string, error) {
	s := strings.TrimSpace(html.UnescapeString(raw))
	if s == "" {
		return "", fmt.Errorf("empty url")
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("invalid or non-http(s) url")
	}
	return u.String(), nil
}

func extFromContentType(ct string) string {
	ct = strings.TrimSpace(strings.ToLower(strings.Split(ct, ";")[0]))
	switch ct {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

// persistRedditPreviewImage downloads a thumbnail into [p_filesystem] storage and returns the new [p_filesystem.VNode] id.
// On any failure it logs and returns nil (ingest still succeeds without a preview file).
func persistRedditPreviewImage(ctx context.Context, db *gorm.DB, post RedditPostData, imageURL string) *uint {
	imageURL, err := normalizePreviewFetchURL(imageURL)
	if err != nil || imageURL == "" {
		if err != nil {
			slog.Warn("lacerate: intel preview url normalize", "error", err)
		}
		return nil
	}
	if p_filesystem.Store == nil {
		slog.Warn("lacerate: filesystem store not configured; skipping intel preview download")
		return nil
	}

	dir := Config.IntelPreview.Directory
	parent, err := p_filesystem.EnsureDirectoryPath(db, dir)
	if err != nil {
		slog.Error("lacerate: ensure intel preview directory", "error", err, "path", dir)
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		slog.Error("lacerate: intel preview request", "error", err, "url", imageURL)
		return nil
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.reddit.com/")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("lacerate: intel preview download", "error", err, "url", imageURL)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("lacerate: intel preview download status", "status", resp.Status, "url", imageURL)
		return nil
	}

	ext := extFromContentType(resp.Header.Get("Content-Type"))
	baseID := strings.TrimSpace(post.ID)
	if baseID == "" {
		baseID = "post"
	}
	// Uniq suffix: listings can repeat the same post; [filesystem_nodes] enforces (parent_id, name, is_directory).
	name := fmt.Sprintf("reddit_%s_%d%s", baseID, time.Now().UnixNano(), ext)
	name = sanitizePreviewFileName(name)

	storedPath, err := p_filesystem.Store.SaveFromReader(resp.Body, ext)
	if err != nil {
		slog.Error("lacerate: intel preview save to store", "error", err, "url", imageURL)
		return nil
	}

	node := &p_filesystem.VNode{
		Name:        name,
		IsDirectory: false,
		FilePath:    storedPath,
	}
	if parent != nil {
		pid := parent.ID
		node.ParentID = &pid
	}

	if err := db.Create(node).Error; err != nil {
		slog.Error("lacerate: intel preview vnode create", "error", err, "url", imageURL)
		if delErr := p_filesystem.Store.Delete(storedPath); delErr != nil {
			slog.Error("lacerate: cleanup store after vnode error", "error", delErr, "path", storedPath)
		}
		return nil
	}
	return &node.ID
}

// persistIntelPreviewImage downloads imageURL into filesystem storage for ingest previews.
// fileNamePrefix is used in the vnode file name (sanitized); referer is optional.
func persistIntelPreviewImage(ctx context.Context, db *gorm.DB, fileNamePrefix, imageURL, referer string) *uint {
	imageURL, err := normalizePreviewFetchURL(imageURL)
	if err != nil || imageURL == "" {
		if err != nil {
			slog.Warn("lacerate: intel preview url normalize", "error", err)
		}
		return nil
	}
	if p_filesystem.Store == nil {
		slog.Warn("lacerate: filesystem store not configured; skipping intel preview download")
		return nil
	}

	dir := Config.IntelPreview.Directory
	parent, err := p_filesystem.EnsureDirectoryPath(db, dir)
	if err != nil {
		slog.Error("lacerate: ensure intel preview directory", "error", err, "path", dir)
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		slog.Error("lacerate: intel preview request", "error", err, "url", imageURL)
		return nil
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if strings.TrimSpace(referer) != "" {
		req.Header.Set("Referer", referer)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("lacerate: intel preview download", "error", err, "url", imageURL)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("lacerate: intel preview download status", "status", resp.Status, "url", imageURL)
		return nil
	}

	ext := extFromContentType(resp.Header.Get("Content-Type"))
	baseID := strings.TrimSpace(fileNamePrefix)
	if baseID == "" {
		baseID = "tweet"
	}
	name := fmt.Sprintf("twitter_%s_%d%s", baseID, time.Now().UnixNano(), ext)
	name = sanitizePreviewFileName(name)

	storedPath, err := p_filesystem.Store.SaveFromReader(resp.Body, ext)
	if err != nil {
		slog.Error("lacerate: intel preview save to store", "error", err, "url", imageURL)
		return nil
	}

	node := &p_filesystem.VNode{
		Name:        name,
		IsDirectory: false,
		FilePath:    storedPath,
	}
	if parent != nil {
		pid := parent.ID
		node.ParentID = &pid
	}

	if err := db.Create(node).Error; err != nil {
		slog.Error("lacerate: intel preview vnode create", "error", err, "url", imageURL)
		if delErr := p_filesystem.Store.Delete(storedPath); delErr != nil {
			slog.Error("lacerate: cleanup store after vnode error", "error", delErr, "path", storedPath)
		}
		return nil
	}
	return &node.ID
}

func sanitizePreviewFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "preview.bin"
	}
	return name
}

func vnodeFileBytes(n *p_filesystem.VNode) ([]byte, error) {
	if n == nil || n.IsDirectory {
		return nil, nil
	}
	dl, err := n.OpenDownload()
	if err != nil {
		return nil, err
	}
	defer dl.Reader.Close()
	return io.ReadAll(dl.Reader)
}
