package p_nirmancampus_announcements

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type announcementsOrderReleaseAtQueryPatcherType struct{}

// announcementsOrderReleaseAtQueryPatcher defaults ordering to release_at ASC
// when the request did not specify sort=.
func (announcementsOrderReleaseAtQueryPatcherType) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Announcement]) gorm.ChainInterface[Announcement] {
	if r.URL.Query().Get("sort") != "" {
		return query
	}
	return query.Order("release_at ASC")
}

var announcementsOrderReleaseAtQueryPatcher views.QueryPatcher[Announcement] = announcementsOrderReleaseAtQueryPatcherType{}

// AnnouncementScopeByRole restricts announcement queries:
//   - superuser, admin: full queryset (all CRUD targets for those views)
//   - student: rows where release_at <= now and (expiry_at IS NULL or expiry_at > now)
//   - any other role: empty queryset
type announcementScopeByRole struct{}

func (announcementScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Announcement]) gorm.ChainInterface[Announcement] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("AnnouncementScopeByRole: missing $user in context – auth middleware not applied?")
		panic("AnnouncementScopeByRole: $user is nil in context")
	}
	if _, ok := rawUser.(p_users.User); !ok {
		slog.Error("AnnouncementScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("AnnouncementScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("AnnouncementScopeByRole: missing $role in context – auth middleware not applied?")
		panic("AnnouncementScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("AnnouncementScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("AnnouncementScopeByRole: $role has wrong type in context")
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		now := time.Now()
		return query.Where("release_at <= ?", now).Where("(expiry_at IS NULL OR expiry_at > ?)", now)
	default:
		return query.Where("1 = 0")
	}
}

var AnnouncementScopeByRole views.QueryPatcher[Announcement] = announcementScopeByRole{}
