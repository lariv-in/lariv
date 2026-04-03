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
	lago.Redirect(w, r, url)
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

type authenticatedUserDetailMiddleware struct{}

func (authenticatedUserDetailMiddleware) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)
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
		db := r.Context().Value("$db").(*gorm.DB)
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
		db := r.Context().Value("$db").(*gorm.DB)

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

		authUser := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)
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

		user := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)
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
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.list", views.MiddlewareList[User]{
				Key: getters.Static("users"),
			}))

	lago.RegistryView.Register("users.DetailView",
		lago.GetPageView("users.UserDetail").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.detail", views.MiddlewareDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("users.CreateView",
		lago.GetPageView("users.UserCreateForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.create", views.MiddlewareCreate[User]{
				SuccessURL: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("users.UpdateView",
		lago.GetPageView("users.UserUpdateForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.detail", views.MiddlewareDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("users.update", views.MiddlewareUpdate[User]{
				Key: getters.Static("user"),
				SuccessURL: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			}))

	lago.RegistryView.Register("users.SelfDetailView",
		lago.GetPageView("users.SelfDetail").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.self_detail", authenticatedUserDetailMiddleware{}))

	lago.RegistryView.Register("users.SelfUpdateView",
		lago.GetPageView("users.SelfUpdateForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.self_detail", authenticatedUserDetailMiddleware{}).
			WithMiddleware("users.self_update", views.MiddlewareUpdate[User]{
				Key:        getters.Static("user"),
				SuccessURL: lago.RoutePath("users.SelfDetailRoute", nil),
			}))

	lago.RegistryView.Register("users.SelfChangePasswordView",
		lago.GetPageView("users.SelfChangePasswordForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.self_detail", authenticatedUserDetailMiddleware{}).
			WithMiddleware("users.self_change_password", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: selfChangePasswordHandler,
			}))

	lago.RegistryView.Register("users.DeleteView",
		lago.GetPageView("users.UserDeleteForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.detail", views.MiddlewareDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("users.delete", views.MiddlewareDelete[User]{
				Key:        getters.Static("user"),
				SuccessURL: lago.RoutePath("users.ListRoute", nil),
			}))

	lago.RegistryView.Register("users.ChangePasswordView",
		lago.GetPageView("users.ChangePasswordForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.detail", views.MiddlewareDetail[User]{
				Key:          getters.Static("user"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("users.change_password", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: changePasswordHandler,
			}))

	lago.RegistryView.Register("users.SelectView",
		lago.GetPageView("users.UserSelectionTable").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.select", views.MiddlewareList[User]{
				Key: getters.Static("users"),
			}))

	lago.RegistryView.Register("users.RoleSelectView",
		lago.GetPageView("users.RoleSelectionTable").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_select", views.MiddlewareList[Role]{
				Key: getters.Static("roles"),
			}))

	lago.RegistryView.Register("users.RoleListView",
		lago.GetPageView("users.RoleTable").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_list", views.MiddlewareList[Role]{
				Key: getters.Static("roles"),
			}))

	lago.RegistryView.Register("users.RoleDetailView",
		lago.GetPageView("users.RoleDetail").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_detail", views.MiddlewareDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("users.RoleCreateView",
		lago.GetPageView("users.RoleCreateForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_create", views.MiddlewareCreate[Role]{
				SuccessURL: lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("users.RoleUpdateView",
		lago.GetPageView("users.RoleUpdateForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_detail", views.MiddlewareDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("users.role_update", views.MiddlewareUpdate[Role]{
				Key: getters.Static("role"),
				SuccessURL: lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("role.ID")),
				}),
			}))

	lago.RegistryView.Register("users.RoleDeleteView",
		lago.GetPageView("users.RoleDeleteForm").
			WithMiddleware("users.auth", AuthenticationMiddleware{}).
			WithMiddleware("users.role", RoleAuthorizationMiddleware{Roles: []string{""}}).
			WithMiddleware("users.role_detail", views.MiddlewareDetail[Role]{
				Key:          getters.Static("role"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("users.role_delete", views.MiddlewareDelete[Role]{
				Key:        getters.Static("role"),
				SuccessURL: lago.RoutePath("users.RoleListRoute", nil),
			}))

	lago.RegistryView.Register("users.LogoutView",
		lago.GetPageView("users.UnauthenticatedPage").
			WithMiddleware("users.logout_post", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: logoutHandler,
			}).
			WithMiddleware("users.logout_get", views.MethodMiddleware{
				Method:  http.MethodGet,
				Handler: logoutHandler,
			}))

	lago.RegistryView.Register("users.LoginView",
		lago.GetPageView("users.LoginPage").
			WithMiddleware("users.login", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: loginHandler,
			}))

	lago.RegistryView.Register("users.SignupView",
		lago.GetPageView("users.SignupPage").
			WithMiddleware("users.signup", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: signupHandler,
			}))

	lago.RegistryView.Register("base.HomeView", lago.RedirectView(lago.RoutePath("users.LoginRoute", nil)))
	lago.RegistryView.Register("users.LoginSuccessView", lago.RedirectView(lago.RoutePath("users.LoginRoute", nil)))
	lago.RegistryView.Register("users.UnauthenticatedView", lago.GetPageView("users.UnauthenticatedPage"))
}
