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
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type queryPatcherScopeConversationByUser struct{}

func (queryPatcherScopeConversationByUser) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Conversation]) gorm.ChainInterface[Conversation] {
	u, ok := r.Context().Value("$user").(p_users.User)
	if !ok {
		slog.Error("sqlagent: missing $user in query patcher")
		return query.Where("1 = 0")
	}
	return query.Where("created_by_id = ?", u.ID)
}

type listContextMiddleware struct{}

func (listContextMiddleware) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyMessages, []ConversationMessage{})
		ctx = context.WithValue(ctx, ContextKeyActiveConversationID, uint(0))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type detailContextMiddleware struct{}

func (detailContextMiddleware) Next(_ views.View, next http.Handler) http.Handler {
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
		convs, err := gorm.G[Conversation](db).Where("created_by_id = ?", u.ID).Order("updated_at DESC").Limit(200).Find(r.Context())
		if err != nil {
			slog.Error("sqlagent: load conversation list", "error", err)
			convs = nil
		}
		ol := components.ObjectList[Conversation]{
			Items:    convs,
			Total:    uint64(len(convs)),
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

type formPatcherCreatedByFromUser struct{}

func (formPatcherCreatedByFromUser) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	u, ok := r.Context().Value("$user").(p_users.User)
	if !ok {
		formErrors["_form"] = errors.New("not authenticated")
		return formData, formErrors
	}
	formData["CreatedByID"] = u.ID
	return formData, formErrors
}

func init() {
	auth := "users.auth"

	lago.RegistryView.Register("sqlagent.ListView",
		lago.GetPageView("sqlagent.ConversationListPage").
			WithMiddleware(auth, p_users.AuthenticationMiddleware{}).
			WithMiddleware("sqlagent.list", views.MiddlewareList[Conversation]{
				Key: getters.Static("conversations"),
				QueryPatchers: views.QueryPatchers[Conversation]{
					registry.Pair[string, views.QueryPatcher[Conversation]]{Key: "sqlagent.scope_user", Value: queryPatcherScopeConversationByUser{}},
					registry.Pair[string, views.QueryPatcher[Conversation]]{Key: "sqlagent.order", Value: views.QueryPatcherOrderBy[Conversation]{Order: "updated_at DESC"}},
				},
			}).
			WithMiddleware("sqlagent.list_ctx", listContextMiddleware{}))

	lago.RegistryView.Register("sqlagent.ConversationDetailView",
		lago.GetPageView("sqlagent.ConversationDetailPage").
			WithMiddleware(auth, p_users.AuthenticationMiddleware{}).
			WithMiddleware("sqlagent.detail", views.MiddlewareDetail[Conversation]{
				Key:          getters.Static("conversation"),
				PathParamKey: getters.Static("conversation_id"),
				QueryPatchers: views.QueryPatchers[Conversation]{
					registry.Pair[string, views.QueryPatcher[Conversation]]{Key: "sqlagent.scope_user", Value: queryPatcherScopeConversationByUser{}},
				},
			}).
			WithMiddleware("sqlagent.detail_ctx", detailContextMiddleware{}))

	lago.RegistryView.Register("sqlagent.ConversationCreateView",
		lago.GetPageView("sqlagent.ConversationCreateForm").
			WithMiddleware(auth, p_users.AuthenticationMiddleware{}).
			WithMiddleware("sqlagent.create", views.MiddlewareCreate[Conversation]{
				SuccessURL: lago.RoutePath("sqlagent.ConversationDetailRoute", map[string]getters.Getter[any]{
					"conversation_id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "sqlagent.created_by", Value: formPatcherCreatedByFromUser{}},
				},
			}))
}
