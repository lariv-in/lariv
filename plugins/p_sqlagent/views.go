package sqlagent

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func queryPatcherScopeConversationByUser() views.QueryPatcher {
	return func(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		u, ok := r.Context().Value("$user").(p_users.User)
		if !ok {
			slog.Error("sqlagent: missing $user in query patcher")
			return query.Where("1 = 0")
		}
		return query.Where("created_by_id = ?", u.ID)
	}
}

func listRenderMiddleware() views.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyMessages, []ConversationMessage{})
			ctx = context.WithValue(ctx, ContextKeyActiveConversationID, uint(0))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func detailRenderMiddleware() views.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conv, ok := r.Context().Value("conversation").(Conversation)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			db := r.Context().Value("$db").(*gorm.DB)
			u, ok := r.Context().Value("$user").(p_users.User)
			if !ok {
				slog.Error("sqlagent: detail render missing $user")
				next.ServeHTTP(w, r)
				return
			}
			msgs, err := LoadMessagesForConversation(db, conv.ID)
			if err != nil {
				slog.Error("sqlagent: load messages", "error", err, "conversation_id", conv.ID)
				msgs = []ConversationMessage{}
			}
			var convs []Conversation
			if err := db.Where("created_by_id = ?", u.ID).Order("updated_at DESC").Limit(200).Find(&convs).Error; err != nil {
				slog.Error("sqlagent: load conversation list", "error", err)
				convs = nil
			}
			ol := components.ObjectList[Conversation]{
				Items:    convs,
				Total:    int64(len(convs)),
				Number:   1,
				NumPages: 1,
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyMessages, msgs)
			ctx = context.WithValue(ctx, ContextKeyActiveConversationID, conv.ID)
			ctx = context.WithValue(ctx, "conversations", ol)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func formPatcherCreatedByFromUser() views.FormPatcher {
	return func(v *views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
		u, ok := r.Context().Value("$user").(p_users.User)
		if !ok {
			formErrors["_form"] = errors.New("not authenticated")
			return formData, formErrors
		}
		formData["CreatedByID"] = u.ID
		return formData, formErrors
	}
}

func init() {
	auth := "users.auth"

	lago.RegistryView.Register("sqlagent.ListView",
		views.ListView[Conversation]("conversations")(
			lago.GetPageView("sqlagent.ConversationListPage"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithQueryPatcher("sqlagent.scope_user", queryPatcherScopeConversationByUser()).
			WithQueryPatcher("sqlagent.order", views.QueryPatcherOrderBy("updated_at DESC")).
			WithRenderMiddleware("sqlagent.list_ctx", listRenderMiddleware()))

	lago.RegistryView.Register("sqlagent.ConversationDetailView",
		views.DetailView[Conversation]("conversation", "conversation_id")(
			lago.GetPageView("sqlagent.ConversationDetailPage"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithQueryPatcher("sqlagent.scope_user", queryPatcherScopeConversationByUser()).
			WithRenderMiddleware("sqlagent.detail_ctx", detailRenderMiddleware()))

	lago.RegistryView.Register("sqlagent.ConversationCreateView",
		views.CreateView[Conversation](
			lago.RoutePath("sqlagent.ConversationDetailRoute", map[string]getters.Getter[any]{
				"conversation_id": getters.Any(getters.Key[uint]("$id")),
			}),
		)(
			lago.GetPageView("sqlagent.ConversationCreateForm"),
		).
			WithMiddleware(auth, p_users.AuthenticationMiddleware).
			WithFormPatcher("sqlagent.created_by", formPatcherCreatedByFromUser()))
}
