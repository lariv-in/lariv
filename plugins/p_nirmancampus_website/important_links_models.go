package p_nirmancampus_website

import (
	"fmt"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

// importantLinkItemBasePath is the URL prefix for ImportantLinkItemRoute (file/download handler).
const importantLinkItemBasePath = "/important-links/item/"

// ImportantLinkPublicURL returns the public href for a row (trimmed external link, or item path for downloads).
func ImportantLinkPublicURL(l ImportantLink) string {
	if l.IsLink {
		return strings.TrimSpace(l.Link)
	}
	return fmt.Sprintf("%s%d/", importantLinkItemBasePath, l.ID)
}

type ImportantLink struct {
	gorm.Model

	Title string `gorm:"notnull"`
	Order int    `gorm:"notnull;default:0"`

	// If IsLink is true, open Link as a normal URL.
	// If IsLink is false, download the attached File.
	IsLink bool
	Link   string

	FileID *uint
	File   *p_filesystem.VNode `gorm:"constraint:OnDelete:SET NULL;foreignKey:FileID;references:ID"`
}

func init() {
	lago.OnDBInit("p_nirmancampus_website.important_links_models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ImportantLink](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_website.important_links", lago.AdminPanel[ImportantLink]{
		SearchField: "Title",
		ListFields:  []string{"Title", "Order", "IsLink", "Link", "UpdatedAt"},
		Preload:     []string{"File"},
	})
}
