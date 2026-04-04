package p_filesystem

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// createComponentVNode creates a VNode for a component-uploaded file, using a
// timestamped name to avoid unique-constraint collisions with other parentless nodes.
func createComponentVNode(db *gorm.DB, basePath string, file *multipart.FileHeader) (*VNode, error) {
	ext := filepath.Ext(file.Filename)
	base := strings.TrimSuffix(file.Filename, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", base, time.Now().UnixMilli(), ext)
	parent, err := EnsureDirectoryPath(db, basePath)
	if err != nil {
		slog.Error("failed to ensure directory path for component upload", "error", err, "basePath", basePath)
		return nil, err
	}
	return CreateVNode(db, uniqueName, false, file, parent)
}

func checkFileType(file *multipart.FileHeader, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ct := file.Header.Get("Content-Type")
	for _, ft := range allowed {
		ft = strings.TrimSpace(ft)
		if strings.EqualFold(ft, ext) || strings.EqualFold(ft, ct) {
			return nil
		}
	}
	return fmt.Errorf("file type %q is not allowed", ext)
}
