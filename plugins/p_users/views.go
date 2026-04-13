package p_users

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
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
	views.HtmxRedirect(w, r, url, http.StatusMovedPermanently)
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
			fmt.Println(1)
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			fmt.Println(2)
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
			fmt.Println(3)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		user.Login(w, r)
		_ = redirectToRoute(w, r, "users.LoginSuccessRoute")
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
		_ = redirectToRoute(w, r, "users.LoginSuccessRoute")
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
		_ = redirectToRoute(w, r, "users.LoginRoute")
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

		_ = redirectToRoute(w, r, "users.DetailRoute", map[string]getters.Getter[any]{
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

		_ = redirectToRoute(w, r, "users.SelfDetailRoute")
	})
}

func init() {
	lago.RegistryView.Register("users.ListView",
		lago.GetPageView("users.UserTable").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.list", views.LayerList[User]{
				Key: getters.Static("users"),
			}))

	lago.RegistryView.Register("users.DetailView",
		lago.GetPageView("users.UserDetail").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.detail", views.LayerDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("users.CreateView",
		lago.GetPageView("users.UserCreateForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.create", views.LayerCreate[User]{
				SuccessURL: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("users.UpdateView",
		lago.GetPageView("users.UserUpdateForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.detail", views.LayerDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("users.update", views.LayerUpdate[User]{
				Key: getters.Static("user"),
				SuccessURL: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			}))

	lago.RegistryView.Register("users.SelfDetailView",
		lago.GetPageView("users.SelfDetail").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.self_detail", authenticatedUserDetailLayer{}))

	lago.RegistryView.Register("users.SelfUpdateView",
		lago.GetPageView("users.SelfUpdateForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.self_detail", authenticatedUserDetailLayer{}).
			WithLayer("users.self_update", views.LayerUpdate[User]{
				Key:        getters.Static("user"),
				SuccessURL: lago.RoutePath("users.SelfDetailRoute", nil),
			}))

	lago.RegistryView.Register("users.SelfChangePasswordView",
		lago.GetPageView("users.SelfChangePasswordForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.self_detail", authenticatedUserDetailLayer{}).
			WithLayer("users.self_change_password", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: selfChangePasswordHandler,
			}))

	lago.RegistryView.Register("users.DeleteView",
		lago.GetPageView("users.UserDeleteForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.detail", views.LayerDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("users.delete", views.LayerDelete[User]{
				Key:        getters.Static("user"),
				SuccessURL: lago.RoutePath("users.ListRoute", nil),
			}))

	lago.RegistryView.Register("users.ChangePasswordView",
		lago.GetPageView("users.ChangePasswordForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.detail", views.LayerDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("users.change_password", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: changePasswordHandler,
			}))

	lago.RegistryView.Register("users.SelectView",
		lago.GetPageView("users.UserSelectionTable").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.select", views.LayerList[User]{
				Key: getters.Static("users"),
			}))

	lago.RegistryView.Register("users.RoleSelectView",
		lago.GetPageView("users.RoleSelectionTable").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_select", views.LayerList[Role]{
				Key: getters.Static("roles"),
			}))

	lago.RegistryView.Register("users.RoleListView",
		lago.GetPageView("users.RoleTable").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_list", views.LayerList[Role]{
				Key: getters.Static("roles"),
			}))

	lago.RegistryView.Register("users.RoleDetailView",
		lago.GetPageView("users.RoleDetail").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_detail", views.LayerDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("users.RoleCreateView",
		lago.GetPageView("users.RoleCreateForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_create", views.LayerCreate[Role]{
				SuccessURL: lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("users.RoleUpdateView",
		lago.GetPageView("users.RoleUpdateForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_detail", views.LayerDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("users.role_update", views.LayerUpdate[Role]{
				Key: getters.Static("role"),
				SuccessURL: lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("role.ID")),
				}),
			}))

	lago.RegistryView.Register("users.RoleDeleteView",
		lago.GetPageView("users.RoleDeleteForm").
			WithLayer("users.auth", AuthenticationLayer{}).
			WithLayer("users.role", RoleAuthorizationLayer{Roles: []string{""}}).
			WithLayer("users.role_detail", views.LayerDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("users.role_delete", views.LayerDelete[Role]{
				Key:        getters.Static("role"),
				SuccessURL: lago.RoutePath("users.RoleListRoute", nil),
			}))

	lago.RegistryView.Register("users.LogoutView",
		lago.GetPageView("users.UnauthenticatedPage").
			WithLayer("users.logout_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: logoutHandler,
			}).
			WithLayer("users.logout_get", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: logoutHandler,
			}))

	lago.RegistryView.Register("users.LoginView",
		lago.GetPageView("users.LoginPage").
			WithLayer("users.login", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: loginHandler,
			}))

	lago.RegistryView.Register("users.SignupView",
		lago.GetPageView("users.SignupPage").
			WithLayer("users.signup", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: signupHandler,
			}))

	lago.RegistryView.Register("base.HomeView", lago.RedirectView(lago.RoutePath("users.LoginRoute", nil)))
	lago.RegistryView.Register("users.LoginSuccessView", lago.RedirectView(lago.RoutePath("users.LoginRoute", nil)))
	lago.RegistryView.Register("users.UnauthenticatedView", lago.GetPageView("users.UnauthenticatedPage"))
}
