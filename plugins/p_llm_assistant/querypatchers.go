package p_llm_assistant

import (
	"net/http"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type assistantSessionUserScope struct{}

func (assistantSessionUserScope) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[LlmAssistantSession]) gorm.ChainInterface[LlmAssistantSession] {
	ctx := r.Context()
	u, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		return query.Where("1 = 0")
	}
	if u.IsSuperuser {
		return query
	}
	return query.Where("user_id = ?", u.ID)
}
