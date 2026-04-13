package p_nirmancampus_website

import (
	"fmt"
	"strings"
)

// ImportantLinkItemURLPrefix is the path prefix for ImportantLinkItemRoute (public file/link handler).
const ImportantLinkItemURLPrefix = "/important-links/item/"

// ImportantLinkPublicURL returns the public href for a row (trimmed external link, or item path for downloads).
func ImportantLinkPublicURL(l ImportantLink) string {
	if l.IsLink {
		return strings.TrimSpace(l.Link)
	}
	return fmt.Sprintf("%s%d/", ImportantLinkItemURLPrefix, l.ID)
}
