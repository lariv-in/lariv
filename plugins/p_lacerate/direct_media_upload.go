package p_lacerate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

// directMediaFSURLPrefix marks a DirectMediaSource.URL that points at a [p_filesystem.VNode]
// created by a form upload (see [directMediaPersistUploadedFile]).
const directMediaFSURLPrefix = "lacerate-fs:"

func parseDirectMediaFSURLNodeID(raw string) (uint, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, directMediaFSURLPrefix) {
		return 0, fmt.Errorf("not a direct media filesystem url")
	}
	idStr := strings.TrimSpace(strings.TrimPrefix(raw, directMediaFSURLPrefix))
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id64 == 0 {
		return 0, fmt.Errorf("invalid direct media filesystem node id")
	}
	return uint(id64), nil
}

func directMediaPersistUploadedFile(ctx context.Context, db *gorm.DB, fh *multipart.FileHeader) (string, error) {
	if fh == nil || strings.TrimSpace(fh.Filename) == "" {
		return "", fmt.Errorf("no upload file")
	}
	if db == nil {
		return "", fmt.Errorf("database is required to store upload")
	}
	if p_filesystem.Store == nil {
		return "", fmt.Errorf("filesystem store is not configured; cannot save upload")
	}

	limit := Config.DirectMedia.MaxDownloadBytes
	if limit <= 0 {
		limit = defaultDirectMediaMaxDownloadBytes
	}

	rc, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open upload: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, limit+1))
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > limit {
		return "", fmt.Errorf("upload exceeds maxDownloadBytes (%d)", limit)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("upload file is empty")
	}

	ct := strings.TrimSpace(fh.Header.Get("Content-Type"))
	ext := filepath.Ext(sanitizePreviewFileName(fh.Filename))
	if ext == "" {
		ext = extFromContentType(ct)
	}
	if ext == "" {
		ext = ".bin"
	}

	dir := strings.TrimSpace(Config.DirectMedia.UploadDirectory)
	if dir == "" {
		dir = defaultDirectMediaUploadDirectory
	}
	parent, err := p_filesystem.EnsureDirectoryPath(db, dir)
	if err != nil {
		slog.Error("lacerate: ensure direct media upload directory", "error", err, "path", dir)
		return "", fmt.Errorf("ensure upload directory: %w", err)
	}

	baseName := strings.TrimSuffix(sanitizePreviewFileName(fh.Filename), ext)
	if baseName == "" {
		baseName = "upload"
	}
	name := fmt.Sprintf("direct_media_%s_%d%s", baseName, time.Now().UnixNano(), ext)
	name = sanitizePreviewFileName(name)

	storedPath, err := p_filesystem.Store.SaveFromReader(bytes.NewReader(data), ext)
	if err != nil {
		slog.Error("lacerate: direct media upload save to store", "error", err)
		return "", fmt.Errorf("save upload: %w", err)
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

	if err := db.WithContext(ctx).Create(node).Error; err != nil {
		slog.Error("lacerate: direct media upload vnode create", "error", err)
		if delErr := p_filesystem.Store.Delete(storedPath); delErr != nil {
			slog.Error("lacerate: cleanup store after vnode error", "error", delErr, "path", storedPath)
		}
		return "", fmt.Errorf("create filesystem node: %w", err)
	}
	return fmt.Sprintf("%s%d", directMediaFSURLPrefix, node.ID), nil
}

func directMediaUploadRootID(db *gorm.DB) (uint, error) {
	dir := strings.TrimSpace(Config.DirectMedia.UploadDirectory)
	if dir == "" {
		dir = defaultDirectMediaUploadDirectory
	}
	root, err := p_filesystem.EnsureDirectoryPath(db, dir)
	if err != nil {
		return 0, err
	}
	if root == nil {
		return 0, fmt.Errorf("upload directory root missing")
	}
	return root.ID, nil
}

func directMediaVNodeDescendsFrom(db *gorm.DB, nodeID, ancestorID uint) (bool, error) {
	for range 1024 {
		var n p_filesystem.VNode
		if err := db.First(&n, nodeID).Error; err != nil {
			return false, err
		}
		if n.ID == ancestorID {
			return true, nil
		}
		if n.ParentID == nil {
			return false, nil
		}
		nodeID = *n.ParentID
	}
	return false, fmt.Errorf("filesystem node parent chain too deep")
}

func directMediaFetchUploadedVNode(ctx context.Context, db *gorm.DB, raw string) (directMediaAsset, error) {
	nodeID, err := parseDirectMediaFSURLNodeID(raw)
	if err != nil {
		return directMediaAsset{}, err
	}
	rootID, err := directMediaUploadRootID(db)
	if err != nil {
		return directMediaAsset{}, fmt.Errorf("upload directory: %w", err)
	}
	var node p_filesystem.VNode
	if err := db.WithContext(ctx).First(&node, nodeID).Error; err != nil {
		return directMediaAsset{}, fmt.Errorf("load uploaded file: %w", err)
	}
	if node.IsDirectory {
		return directMediaAsset{}, fmt.Errorf("upload url points at a directory")
	}
	ok, err := directMediaVNodeDescendsFrom(db, node.ID, rootID)
	if err != nil {
		return directMediaAsset{}, err
	}
	if !ok {
		return directMediaAsset{}, fmt.Errorf("upload url is outside the direct media upload directory")
	}

	data, err := vnodeFileBytes(&node)
	if err != nil {
		return directMediaAsset{}, fmt.Errorf("read uploaded file: %w", err)
	}
	if len(data) == 0 {
		return directMediaAsset{}, fmt.Errorf("uploaded file is empty")
	}

	name := node.Name
	mimeType := directMediaInferMIMEType(name, "", data)
	return directMediaAsset{
		SourceURL:   raw,
		DisplayName: sanitizePreviewFileName(name),
		MIMEType:    mimeType,
		SizeBytes:   int64(len(data)),
		Bytes:       data,
	}, nil
}
