package p_filesystem

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const vnodeTableName = "filesystem_nodes"

// ErrVNodeNoFileSize is returned by [VNode.GetFileSize] for directories and for file
// nodes with an empty storage path.
var ErrVNodeNoFileSize = errors.New("p_filesystem: vnode has no file size")

type VNode struct {
	gorm.Model

	Name        string `gorm:"notnull"`
	IsDirectory bool   `gorm:"notnull"`
	FilePath    string
	ParentID    *uint  `gorm:"index"`
	Parent      *VNode `gorm:"constraint:OnDelete:CASCADE"`

	// ResolvedPath is "/name/..." and ListChildrenCount is a label like "3 items" or "-", set by view
	// layers; not in the database.
	ResolvedPath      string `gorm:"-"`
	ListChildrenCount string `gorm:"-"`
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
	node, err := gorm.G[VNode](db).Where("id = ?", id).First(context.Background())
	if err != nil {
		return nil, err
	}
	return new(node), nil
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

		chain := gorm.G[VNode](db).Where("name = ?", name)
		if current == nil {
			chain = chain.Where("parent_id IS NULL")
		} else {
			chain = chain.Where("parent_id = ?", current.ID)
		}

		next, err := chain.First(context.Background())
		if err != nil {
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

// EnsureDirectoryPath walks rawPath and returns the VNode for the final segment,
// creating missing directory nodes as needed. For "" or "/", returns (nil, nil) for
// uploads at the filesystem root.
func EnsureDirectoryPath(db *gorm.DB, rawPath string) (*VNode, error) {
	cleaned := strings.TrimSpace(rawPath)
	if cleaned == "" || cleaned == "/" {
		return nil, nil
	}

	parts := strings.Split(strings.Trim(cleaned, "/"), "/")
	var current *VNode

	for _, part := range parts {
		name := sanitizeNodeName(part)
		if name == "" {
			return nil, fmt.Errorf("invalid path segment %q", part)
		}

		chain := gorm.G[VNode](db).Where("name = ?", name)
		if current == nil {
			chain = chain.Where("parent_id IS NULL")
		} else {
			chain = chain.Where("parent_id = ?", current.ID)
		}

		next, err := chain.First(context.Background())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created, cerr := CreateVNode(db, name, true, nil, current)
				if cerr != nil {
					return nil, cerr
				}
				current = created
				continue
			}
			return nil, err
		}
		if !next.IsDirectory {
			return nil, fmt.Errorf("%q is not a directory", name)
		}
		current = &next
	}

	return current, nil
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

	if err := gorm.G[VNode](db).Create(context.Background(), node); err != nil {
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
	children, err := gorm.G[VNode](db).Where("parent_id = ?", n.ID).Find(context.Background())
	if err != nil {
		return err
	}
	for i := range children {
		if err := children[i].DeleteTree(db); err != nil {
			return err
		}
	}
	_, err = gorm.G[VNode](db).Where("id = ?", n.ID).Delete(context.Background())
	return err
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

func (n *VNode) GetFileSize() (uint64, error) {
	if n.IsDirectory || n.FilePath == "" {
		return 0, ErrVNodeNoFileSize
	}
	if Store == nil {
		return 0, fmt.Errorf("p_filesystem: store not configured")
	}
	size, err := Store.StoredSize(n.FilePath)
	if err != nil {
		if !IsStoredFileMissing(err) {
			slog.Error("failed to stat stored file", "path", n.FilePath, "error", err)
		}
		return 0, err
	}
	if size < 0 {
		return 0, fmt.Errorf("p_filesystem: negative stored size %d", size)
	}
	return uint64(size), nil
}

// FileSizeDisplay returns a short label for UI: human-readable size, "-", "Missing", or "Error".
func (n *VNode) FileSizeDisplay() string {
	sz, err := n.GetFileSize()
	if err != nil {
		if errors.Is(err, ErrVNodeNoFileSize) {
			return "-"
		}
		if IsStoredFileMissing(err) {
			return "Missing"
		}
		return "Error"
	}
	return HumanReadableSize(sz)
}

func (n *VNode) GetChildrenCount(db *gorm.DB) string {
	if !n.IsDirectory {
		return "-"
	}

	count, err := gorm.G[VNode](db).Where("parent_id = ?", n.ID).Count(context.Background(), "*")
	if err != nil {
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

func HumanReadableSize(size uint64) string {
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
	lago.OnDBInit("p_filesystem.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[VNode](d)
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
