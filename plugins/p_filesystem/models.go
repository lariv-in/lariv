package p_filesystem

import (
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const vnodeTableName = "filesystem_nodes"

type VNode struct {
	gorm.Model

	Name        string `gorm:"notnull"`
	IsDirectory bool   `gorm:"notnull"`
	FilePath    string
	ParentID    *uint  `gorm:"index"`
	Parent      *VNode `gorm:"constraint:OnDelete:CASCADE"`
}

func (VNode) TableName() string {
	return vnodeTableName
}

func sanitizeNodeName(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func GetVNodeByID(db *gorm.DB, id uint) (*VNode, error) {
	var node VNode
	if err := db.First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func GetVNodeByPath(db *gorm.DB, rawPath string) (*VNode, string, error) {
	cleaned := strings.TrimSpace(rawPath)
	if cleaned == "" || cleaned == "/" {
		return nil, "/", nil
	}

	parts := strings.Split(strings.Trim(cleaned, "/"), "/")
	var current *VNode

	for i, part := range parts {
		name := sanitizeNodeName(part)
		if name == "" {
			return nil, "", fmt.Errorf("invalid path segment %q", part)
		}

		query := db.Where("name = ?", name)
		if current == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", current.ID)
		}

		var next VNode
		if err := query.First(&next).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				traversed := "/"
				if i > 0 {
					traversed += strings.Join(parts[:i], "/")
				}
				return nil, "", fmt.Errorf("path not found: %q does not exist in %q", name, traversed)
			}
			return nil, "", err
		}
		current = &next
	}

	return current, "/" + strings.Join(parts, "/"), nil
}

func ListChildrenForParent(db *gorm.DB, parentID *uint) *gorm.DB {
	query := db.Model(&VNode{}).Order("is_directory DESC").Order("name ASC")
	if parentID == nil {
		return query.Where("parent_id IS NULL")
	}
	return query.Where("parent_id = ?", *parentID)
}

func CreateVNode(db *gorm.DB, name string, isDirectory bool, file *multipart.FileHeader, parent *VNode) (*VNode, error) {
	if parent != nil && !parent.IsDirectory {
		return nil, fmt.Errorf("%q is not a directory", parent.Name)
	}

	if !isDirectory && file == nil {
		return nil, fmt.Errorf("file upload is required")
	}

	if file != nil && strings.TrimSpace(name) == "" {
		name = file.Filename
	}
	name = sanitizeNodeName(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	storedPath, err := Store.Save(file)
	if err != nil {
		return nil, err
	}

	node := &VNode{
		Name:        name,
		IsDirectory: isDirectory,
		FilePath:    storedPath,
	}
	if parent != nil {
		node.ParentID = &parent.ID
	}

	if err := db.Create(node).Error; err != nil {
		if deleteErr := Store.Delete(storedPath); deleteErr != nil {
			slog.Error("failed cleaning up stored file after create error", "path", storedPath, "error", deleteErr)
		}
		return nil, err
	}
	return node, nil
}

func (n *VNode) Update(db *gorm.DB, name string, file *multipart.FileHeader) error {
	name = sanitizeNodeName(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}

	n.Name = name
	oldPath := n.FilePath
	if file != nil {
		if n.IsDirectory {
			return fmt.Errorf("cannot upload a file for a directory")
		}
		storedPath, err := Store.Save(file)
		if err != nil {
			return err
		}
		n.FilePath = storedPath
		if err := db.Save(n).Error; err != nil {
			if deleteErr := Store.Delete(storedPath); deleteErr != nil {
				slog.Error("failed cleaning up stored file after update error", "path", storedPath, "error", deleteErr)
			}
			return err
		}
		if oldPath != "" && oldPath != storedPath {
			if err := Store.Delete(oldPath); err != nil {
				slog.Error("failed deleting replaced stored file", "path", oldPath, "error", err)
			}
		}
		return nil
	}

	return db.Save(n).Error
}

func (n *VNode) MoveToNode(db *gorm.DB, destination *VNode) error {
	if destination != nil {
		if !destination.IsDirectory {
			return fmt.Errorf("destination must be a directory")
		}
		if destination.ID == n.ID {
			return fmt.Errorf("cannot move an item into itself")
		}
		isDescendant, err := destination.IsDescendantOf(db, n.ID)
		if err != nil {
			return err
		}
		if isDescendant {
			return fmt.Errorf("cannot move an item into its descendants")
		}
		n.ParentID = &destination.ID
	} else {
		n.ParentID = nil
	}

	return db.Save(n).Error
}

func (n *VNode) DeleteTree(db *gorm.DB) error {
	var children []VNode
	if err := db.Where("parent_id = ?", n.ID).Find(&children).Error; err != nil {
		return err
	}
	for i := range children {
		if err := children[i].DeleteTree(db); err != nil {
			return err
		}
	}
	return db.Delete(n).Error
}

func (n *VNode) IsDescendantOf(db *gorm.DB, ancestorID uint) (bool, error) {
	currentParentID := n.ParentID
	for currentParentID != nil {
		if *currentParentID == ancestorID {
			return true, nil
		}
		parent, err := GetVNodeByID(db, *currentParentID)
		if err != nil {
			return false, err
		}
		currentParentID = parent.ParentID
	}
	return false, nil
}

func (n *VNode) GetPath(db *gorm.DB) string {
	segments := []string{n.Name}
	currentParentID := n.ParentID

	for currentParentID != nil {
		parent, err := GetVNodeByID(db, *currentParentID)
		if err != nil {
			slog.Error("failed to resolve vnode path", "id", n.ID, "parent_id", *currentParentID, "error", err)
			break
		}
		segments = append([]string{parent.Name}, segments...)
		currentParentID = parent.ParentID
	}

	return "/" + strings.Join(segments, "/")
}

func (n *VNode) GetItemType() string {
	if n.IsDirectory {
		return "Directory"
	}
	return "File"
}

func (n *VNode) GetFileSize() string {
	if n.IsDirectory || n.FilePath == "" {
		return "-"
	}
	info, err := os.Stat(n.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "Missing"
		}
		slog.Error("failed to stat stored file", "path", n.FilePath, "error", err)
		return "Error"
	}
	return humanReadableSize(info.Size())
}

func (n *VNode) GetChildrenCount(db *gorm.DB) string {
	if !n.IsDirectory {
		return "-"
	}

	var count int64
	if err := db.Model(&VNode{}).Where("parent_id = ?", n.ID).Count(&count).Error; err != nil {
		slog.Error("failed to count vnode children", "id", n.ID, "error", err)
		return "Error"
	}
	return fmt.Sprintf("%d items", count)
}

func (n *VNode) OpenDownload() (*FileDownload, error) {
	if n.IsDirectory {
		return nil, fmt.Errorf("cannot download a directory")
	}
	if n.FilePath == "" {
		return nil, fmt.Errorf("file not found")
	}
	return Store.Open(n.FilePath, n.Name)
}

func (n *VNode) AfterDelete(*gorm.DB) error {
	if err := Store.Delete(n.FilePath); err != nil {
		slog.Error("failed deleting stored file after vnode delete", "path", n.FilePath, "error", err)
	}
	return nil
}

func humanReadableSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	for _, unit := range units {
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
		value /= 1024
	}
	return "-"
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&VNode{}); err != nil {
			panic(err)
		}
		// Replace the earlier index with a partial unique index so soft-deleted
		// rows do not block re-creating files/folders with the same name.
		if err := d.Exec(
			"DROP INDEX IF EXISTS filesystem_nodes_parent_name_dir_uidx",
		).Error; err != nil {
			panic(err)
		}
		if err := d.Exec(
			"CREATE UNIQUE INDEX IF NOT EXISTS filesystem_nodes_parent_name_dir_uidx ON filesystem_nodes (COALESCE(parent_id, 0), name, is_directory) WHERE deleted_at IS NULL",
		).Error; err != nil {
			panic(err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_filesystem", lago.AdminPanel[VNode]{
		SearchField: "Name",
		ListFields:  []string{"Name", "IsDirectory", "ParentID", "UpdatedAt"},
	})
}
