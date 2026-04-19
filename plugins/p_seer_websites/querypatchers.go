package p_seer_websites

import (
	"net/http"

	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type websiteActiveOnlyPatcher struct{}

func (websiteActiveOnlyPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[Website]) gorm.ChainInterface[Website] {
	return q.Where("deleted_at IS NULL")
}
