package p_users

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func redirectToRoute(w http.ResponseWriter, r *http.Request, routeKey string, args ...map[string]getters.Getter[any]) bool {
	var routeArgs map[string]getters.Getter[any]
	if len(args) > 0 {
		routeArgs = args[0]
	}
	url, err := getters.IfOr(lago.RoutePath(routeKey, routeArgs), r.Context(), "")
	if err != nil || url == "" {
		http.NotFound(w, r)
		return false
	}
	views.HtmxRedirect(w, r, url, http.StatusSeeOther)
	return true
}

func changeUserPassword(db *gorm.DB, userID uint, newPassword string) error {
	targetUser, err := gorm.G[User](db).Where("id = ?", userID).Last(context.Background())
	if err != nil {
		return err
	}
	targetUser.Password = []byte(newPassword)
	return db.Save(&targetUser).Error
}

type authenticatedUserDetailLayer struct{}

func (authenticatedUserDetailLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser := UserFromContext(r.Context(), "authenticatedUserDetailLayer")
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		user, err := gorm.G[User](db).Where("id = ?", authUser.ID).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loginHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		email, _ := values["Email"].(string)
		password, _ := values["Password"].(string)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		user, err := Authenticate(db, email, password)
		if err != nil {
			fieldErrors["Password"] = fmt.Errorf("invalid email or password")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user.Login(w, r)
		_ = redirectToRoute(w, r, "p_users.LoginSuccessRoute")
	})
}

func signupHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		password1, _ := values["password1"].(string)
		password2, _ := values["password2"].(string)
		termsAccepted, _ := values["terms_accepted"].(bool)
		if !termsAccepted {
			fieldErrors["terms_accepted"] = fmt.Errorf("terms and conditions need to be accepted")
		}
		if password1 != password2 {
			fieldErrors["password2"] = fmt.Errorf("passwords do not match")
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		name, _ := values["Name"].(string)
		email, _ := values["Email"].(string)
		phone, _ := values["Phone"].(string)
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if _, err := gorm.G[User](db).Where("email = ?", email).Last(r.Context()); err == nil {
			fieldErrors["Email"] = fmt.Errorf("an account with this email already exists")
		}
		if _, err := gorm.G[User](db).Where("phone = ?", phone).Last(r.Context()); err == nil {
			fieldErrors["Phone"] = fmt.Errorf("an account with this phone number already exists")
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		unassignedRole := Role{Name: "unassigned"}
		if err := db.Where(unassignedRole).Attrs(unassignedRole).FirstOrCreate(&unassignedRole).Error; err != nil {
			fieldErrors["_form"] = fmt.Errorf("failed to load unassigned role: %v", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user := User{
			Name:        name,
			Email:       email,
			Phone:       phone,
			IsSuperuser: false,
			Password:    []byte(password1),
			Role:        unassignedRole,
		}
		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&user).Error; err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user.Login(w, r)
		_ = redirectToRoute(w, r, "p_users.LoginSuccessRoute")
	})
}

func logoutHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "auth-token",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})
		_ = redirectToRoute(w, r, "p_users.LoginRoute")
	})
}

func changePasswordHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		newPassword, _ := values["new_password"].(string)
		confirmPassword, _ := values["confirm_password"].(string)
		if newPassword != confirmPassword {
			fieldErrors["confirm_password"] = fmt.Errorf("passwords do not match")
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		authUser := UserFromContext(r.Context(), "changePasswordHandler")
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		id64, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		targetID := uint(id64)

		if !authUser.IsSuperuser && targetID != authUser.ID {
			fieldErrors["_form"] = fmt.Errorf("unauthorized")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if err := changeUserPassword(db, targetID, newPassword); err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		_ = redirectToRoute(w, r, "p_users.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(targetID)),
		})
	})
}

func selfChangePasswordHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		newPassword, _ := values["new_password"].(string)
		confirmPassword, _ := values["confirm_password"].(string)
		if newPassword != confirmPassword {
			fieldErrors["confirm_password"] = fmt.Errorf("passwords do not match")
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user := UserFromContext(r.Context(), "selfChangePasswordHandler")
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if err := changeUserPassword(db, user.ID, newPassword); err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		_ = redirectToRoute(w, r, "p_users.SelfDetailRoute")
	})
}

func pluginViews() lago.PluginFeatures[*views.View] {
	return lago.PluginFeatures[*views.View]{
		Entries: []registry.Pair[string, *views.View]{
			{Key: "p_users.ListView", Value: lago.GetPageView("p_users.UserTable").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.list", views.LayerList[User]{
					Key: getters.Static("users"),
				})},
			{Key: "p_users.DetailView", Value: lago.GetPageView("p_users.UserDetail").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.detail", views.LayerDetail[User]{
					Key:          getters.Static("user"),
					PathParamKey: getters.Static("id"),
				})},
			{Key: "p_users.CreateView", Value: lago.GetPageView("p_users.UserCreateForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.create", views.LayerCreate[User]{
					SuccessURL: lago.RoutePath("p_users.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$id")),
					}),
				})},
			{Key: "p_users.UpdateView", Value: lago.GetPageView("p_users.UserUpdateForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.detail", views.LayerDetail[User]{
					Key:          getters.Static("user"),
					PathParamKey: getters.Static("id"),
				}).
				WithLayer("p_users.update", views.LayerUpdate[User]{
					Key: getters.Static("user"),
					SuccessURL: lago.RoutePath("p_users.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("user.ID")),
					}),
				})},
			{Key: "p_users.SelfDetailView", Value: lago.GetPageView("p_users.SelfDetail").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.self_detail", authenticatedUserDetailLayer{})},
			{Key: "p_users.SelfUpdateView", Value: lago.GetPageView("p_users.SelfUpdateForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.self_detail", authenticatedUserDetailLayer{}).
				WithLayer("p_users.self_update", views.LayerUpdate[User]{
					Key:        getters.Static("user"),
					SuccessURL: lago.RoutePath("p_users.SelfDetailRoute", nil),
				})},
			{Key: "p_users.SelfChangePasswordView", Value: lago.GetPageView("p_users.SelfChangePasswordForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.self_detail", authenticatedUserDetailLayer{}).
				WithLayer("p_users.self_change_password", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: selfChangePasswordHandler,
				})},
			{Key: "p_users.DeleteView", Value: lago.GetPageView("p_users.UserDeleteForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.detail", views.LayerDetail[User]{
					Key:          getters.Static("user"),
					PathParamKey: getters.Static("id"),
				}).
				WithLayer("p_users.delete", views.LayerDelete[User]{
					Key:        getters.Static("user"),
					SuccessURL: lago.RoutePath("p_users.ListRoute", nil),
				})},
			{Key: "p_users.ChangePasswordView", Value: lago.GetPageView("p_users.ChangePasswordForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.detail", views.LayerDetail[User]{
					Key:          getters.Static("user"),
					PathParamKey: getters.Static("id"),
				}).
				WithLayer("p_users.change_password", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: changePasswordHandler,
				})},
			{Key: "p_users.SelectView", Value: lago.GetPageView("p_users.UserSelectionTable").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.select", views.LayerList[User]{
					Key: getters.Static("users"),
				})},
			{Key: "p_users.RoleSelectView", Value: lago.GetPageView("p_users.RoleSelectionTable").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_select", views.LayerList[Role]{
					Key: getters.Static("roles"),
				})},
			{Key: "p_users.RoleListView", Value: lago.GetPageView("p_users.RoleTable").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_list", views.LayerList[Role]{
					Key: getters.Static("roles"),
				})},
			{Key: "p_users.RoleDetailView", Value: lago.GetPageView("p_users.RoleDetail").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_detail", views.LayerDetail[Role]{
					Key:          getters.Static("role"),
					PathParamKey: getters.Static("id"),
				})},
			{Key: "p_users.RoleCreateView", Value: lago.GetPageView("p_users.RoleCreateForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_create", views.LayerCreate[Role]{
					SuccessURL: lago.RoutePath("p_users.RoleDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$id")),
					}),
				})},
			{Key: "p_users.RoleUpdateView", Value: lago.GetPageView("p_users.RoleUpdateForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_detail", views.LayerDetail[Role]{
					Key:          getters.Static("role"),
					PathParamKey: getters.Static("id"),
				}).
				WithLayer("p_users.role_update", views.LayerUpdate[Role]{
					Key: getters.Static("role"),
					SuccessURL: lago.RoutePath("p_users.RoleDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("role.ID")),
					}),
				})},
			{Key: "p_users.RoleDeleteView", Value: lago.GetPageView("p_users.RoleDeleteForm").
				WithLayer("p_users.auth", AuthenticationLayer{}).
				WithLayer("p_users.role", RoleAuthorizationLayer{Roles: []string{""}}).
				WithLayer("p_users.role_detail", views.LayerDetail[Role]{
					Key:          getters.Static("role"),
					PathParamKey: getters.Static("id"),
				}).
				WithLayer("p_users.role_delete", views.LayerDelete[Role]{
					Key:        getters.Static("role"),
					SuccessURL: lago.RoutePath("p_users.RoleListRoute", nil),
				})},
			{Key: "p_users.LogoutView", Value: lago.GetPageView("p_users.UnauthenticatedPage").
				WithLayer("p_users.logout_post", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: logoutHandler,
				}).
				WithLayer("p_users.logout_get", views.MethodLayer{
					Method:  http.MethodGet,
					Handler: logoutHandler,
				})},
			{Key: "p_users.LoginView", Value: lago.GetPageView("p_users.LoginPage").
				WithLayer("p_users.login", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: loginHandler,
				})},
			{Key: "p_users.SignupView", Value: lago.GetPageView("p_users.SignupPage").
				WithLayer("p_users.signup", views.MethodLayer{
					Method:  http.MethodPost,
					Handler: signupHandler,
				})},
			// Post-login/signup landing; dashboards often patch this to a concrete home (e.g. apps list).
			// Default avoids sending an authenticated session back to the login form.
			{Key: "p_users.LoginSuccessView", Value: lago.RedirectView(lago.RoutePath("core.HomeRoute", nil))},
			{Key: "p_users.UnauthenticatedView", Value: lago.GetPageView("p_users.UnauthenticatedPage")},
		},
	}
}
