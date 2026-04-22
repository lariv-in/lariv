package p_seer_websites

import (
	"github.com/lariv-in/lago/syncmap"
)

var websiteSourceCrawlBusy *syncmap.SyncMap[uint, struct{}] = &syncmap.SyncMap[uint, struct{}]{}

// WebsiteSourceCrawlIsRunning reports whether a crawl is in progress for this source
// ([FetchWebsiteSource] holds the slot until return).
func WebsiteSourceCrawlIsRunning(sourceID uint) bool {
	if sourceID == 0 {
		return false
	}
	_, ok := websiteSourceCrawlBusy.Load(sourceID)
	return ok
}
