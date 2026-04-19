package p_seer_deepsearch

import (
	"net/http"

	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type deepSearchActiveOnlyPatcher struct{}

func (deepSearchActiveOnlyPatcher) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[DeepSearch]) gorm.ChainInterface[DeepSearch] {
	return q.Where("deleted_at IS NULL")
}
