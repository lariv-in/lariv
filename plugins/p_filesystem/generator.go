package p_filesystem

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const (
	maxPhotosToDownload = 15
	picsumURL           = "https://picsum.photos/400/300.jpg"
	httpTimeout         = 15 * time.Second
)

// downloadPhoto fetches a random JPEG from picsum.photos and saves it via
// the configured Filestore, returning the stored path and a display name.
func downloadPhoto(index int) (storedPath, fileName string, err error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(picsumURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to download photo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("picsum returned status %d", resp.StatusCode)
	}

	fileName = fmt.Sprintf("photo_%04d.jpg", index)
	storedPath, err = Store.SaveFromReader(resp.Body, ".jpg")
	if err != nil {
		return "", "", fmt.Errorf("failed saving downloaded photo: %w", err)
	}

	return storedPath, fileName, nil
}

// GeneratePhotoFile downloads a new photo or picks an existing file VNode
// (once we have enough photos). This mirrors the Django generate_photo_file().
func GeneratePhotoFile(db *gorm.DB) (*VNode, error) {
	fileCount, err := gorm.G[VNode](db).Where("is_directory = ?", false).Count(context.Background(), "*")
	if err != nil {
		return nil, err
	}

	if fileCount < int64(maxPhotosToDownload) {
		storedPath, fileName, err := downloadPhoto(int(fileCount) + 1)
		if err != nil {
			slog.Warn("photo download failed, skipping", "error", err)
			return nil, nil
		}

		node := &VNode{
			Name:        fileName,
			IsDirectory: false,
			FilePath:    storedPath,
		}
		if err := gorm.G[VNode](db).Create(context.Background(), node); err != nil {
			// Clean up the written file on DB failure.
			if deleteErr := Store.Delete(storedPath); deleteErr != nil {
				slog.Error("failed cleaning up stored file after create error", "path", storedPath, "error", deleteErr)
			}
			return nil, err
		}
		return node, nil
	}

	// Pick a random existing file.
	files, err := gorm.G[VNode](db).Where("is_directory = ?", false).Find(context.Background())
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	picked := files[rand.Intn(len(files))]
	return new(picked), nil
}

func init() {
	lago.RegistryGenerator.Register("filesystem.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			// 1. Create a root directory "Generated Photos"
			dir := &VNode{
				Name:        "Generated Photos",
				IsDirectory: true,
			}
			if err := gorm.G[VNode](db).Create(context.Background(), dir); err != nil {
				return fmt.Errorf("failed to create photos directory: %w", err)
			}
			fmt.Println("Created directory: Generated Photos")

			// 2. Download photos and add them as children of the directory
			const photosInDir = 8
			created := 0
			for i := range photosInDir {
				storedPath, fileName, err := downloadPhoto(1000 + i)
				if err != nil {
					slog.Warn("photo download failed, skipping", "index", i, "error", err)
					continue
				}

				node := &VNode{
					Name:        fileName,
					IsDirectory: false,
					FilePath:    storedPath,
					ParentID:    &dir.ID,
				}
				if err := gorm.G[VNode](db).Create(context.Background(), node); err != nil {
					slog.Error("failed creating photo VNode", "name", fileName, "error", err)
					if deleteErr := Store.Delete(storedPath); deleteErr != nil {
						slog.Error("failed cleaning up stored file", "path", storedPath, "error", deleteErr)
					}
					continue
				}
				created++
			}
			fmt.Printf("Created %d photos inside 'Generated Photos' directory\n", created)

			// 3. Also create a few loose (root-level) photo files
			const loosePhotos = 5
			looseCreated := 0
			for i := range loosePhotos {
				storedPath, fileName, err := downloadPhoto(2000 + i)
				if err != nil {
					slog.Warn("loose photo download failed, skipping", "index", i, "error", err)
					continue
				}

				node := &VNode{
					Name:        fileName,
					IsDirectory: false,
					FilePath:    storedPath,
				}
				if err := gorm.G[VNode](db).Create(context.Background(), node); err != nil {
					slog.Error("failed creating loose photo VNode", "name", fileName, "error", err)
					if deleteErr := Store.Delete(storedPath); deleteErr != nil {
						slog.Error("failed cleaning up stored file", "path", storedPath, "error", deleteErr)
					}
					continue
				}
				looseCreated++
			}
			fmt.Printf("Created %d loose root-level photos\n", looseCreated)

			return nil
		},
		Remove: func(db *gorm.DB) error {
			// VNode.AfterDelete hook handles file cleanup on disk automatically.
			return db.Unscoped().Where("1=1").Delete(&VNode{}).Error
		},
	})
}
