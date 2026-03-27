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

// --- Auth Handlers (user-specific, not generalizable) ---

func SelfUserQueryPatcher(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	user := r.Context().Value("$user").(User)
	return query.Where("id = ?", user.ID)
}

func changeUserPassword(db *gorm.DB, userID uint, newPassword string) error {
	var targetUser User
	if err := db.Model(User{}).Last(&targetUser, "ID = ?", userID).Error; err != nil {
		return err
	}
	targetUser.Password = []byte(newPassword)
	return db.Save(&targetUser).Error
}

func SelfDetailHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)

		query := db.Model(new(User))
		for _, queryPatcher := range v.QueryPatchers {
			query = queryPatcher.Value(v, r, query)
		}

		var instance User
		if err := query.First(&instance, user.ID).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "user", instance)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func SelfUpdateHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			v.RenderWithErrors(w, r, map[string]error{"_form": err}, values)
			return
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		user := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)
		query := db.Model(new(User)).Where("id = ?", user.ID)
		for _, queryPatcher := range v.QueryPatchers {
			query = queryPatcher.Value(v, r, query)
		}

		if err := query.Updates(values).Error; err != nil {
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		lago.NewRedirectView("users.SelfDetailRoute").ServeHTTP(w, r)
	})
}

func LoginHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		email, _ := values["Email"].(string)
		password, _ := values["Password"].(string)

		db := r.Context().Value("$db").(*gorm.DB)
		user, err := Authenticate(db, email, password)
		if err != nil {
			fieldErrors["Password"] = fmt.Errorf("Invalid email or password")
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		user.Login(w)
		lago.NewRedirectView("users.LoginSuccessRoute").ServeHTTP(w, r)
	})
}

func SignupHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		password1Str, _ := values["password1"].(string)
		password2Str, _ := values["password2"].(string)

		termsAndConditions, _ := values["terms_accepted"].(bool)
		if !termsAndConditions {
			fieldErrors["terms_accepted"] = fmt.Errorf("Terms and conditions need to be accepted")
		}

		if password1Str != password2Str {
			fieldErrors["password2"] = fmt.Errorf("Passwords do not match")
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		name, _ := values["Name"].(string)
		email, _ := values["Email"].(string)
		phone, _ := values["Phone"].(string)
		db := r.Context().Value("$db").(*gorm.DB)

		// Check for existing user by email and phone to surface friendly errors
		var existingByEmail User
		if err := db.Where("email = ?", email).Last(&existingByEmail).Error; err == nil {
			fieldErrors["Email"] = fmt.Errorf("An account with this email already exists")
		}

		var existingByPhone User
		if err := db.Where("phone = ?", phone).Last(&existingByPhone).Error; err == nil {
			fieldErrors["Phone"] = fmt.Errorf("An account with this phone number already exists")
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		var unassignedRole Role
		unassignedRole.Name = "Unassigned"
		if err := db.Where(unassignedRole).Attrs(unassignedRole).FirstOrCreate(&unassignedRole).Error; err == nil {
			fieldErrors["_form"] = fmt.Errorf("Unknown error with role %e", err)
		}

		user := User{
			Name:        name,
			Email:       email,
			Phone:       phone,
			IsSuperuser: false,
			Password:    []byte(password1Str),
			Role:        unassignedRole,
		}
		err = db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&user).Error
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		user.Login(w)
		lago.NewRedirectView("users.LoginSuccessRoute").ServeHTTP(w, r)
	})
}

func LogoutHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "auth-token",
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})
		lago.NewRedirectView("users.LoginRoute").ServeHTTP(w, r)
	})
}

// ChangePasswordHandler is user-specific so it stays here.
func ChangePasswordHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		newPassword, _ := values["new_password"].(string)
		confirmPassword, _ := values["confirm_password"].(string)

		if newPassword != confirmPassword {
			fieldErrors["confirm_password"] = fmt.Errorf("Passwords do not match")
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		user := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if !user.IsSuperuser {
			if uint(id) != user.ID {
				fieldErrors["_form"] = fmt.Errorf("Unauthorized")
				v.RenderWithErrors(w, r, fieldErrors, values)
				return

			}
		}

		if err := changeUserPassword(db, uint(id), newPassword); err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", id))
		lago.NewRedirectView("users.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterKey[string]("$id")),
		}).ServeHTTP(w, r.WithContext(ctx))
	})
}

func SelfChangePasswordHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		newPassword, _ := values["new_password"].(string)
		confirmPassword, _ := values["confirm_password"].(string)

		if newPassword != confirmPassword {
			fieldErrors["confirm_password"] = fmt.Errorf("Passwords do not match")
		}

		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		user := r.Context().Value("$user").(User)
		db := r.Context().Value("$db").(*gorm.DB)

		if err := changeUserPassword(db, user.ID, newPassword); err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		lago.NewRedirectView("users.SelfDetailRoute", nil).ServeHTTP(w, r)
	})
}

// --- View Registrations ---

func init() {
	// List view
	lago.RegistryView.Register("users.ListView",
		views.ListView[User]("users")(
			lago.GetPageView("users.UserTable")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Detail view
	lago.RegistryView.Register("users.DetailView",
		views.DetailView[User]("user", "id")(
			lago.GetPageView("users.UserDetail")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Create view
	lago.RegistryView.Register("users.CreateView",
		views.CreateView[User](lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("users.UserCreateForm")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Update view
	lago.RegistryView.Register("users.UpdateView",
		views.DetailView[User]("user", "id")(
			views.UpdateView[User]("id", lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("users.UserUpdateForm"))).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Self detail view (only current user)
	lago.RegistryView.Register("users.SelfDetailView",
		lago.GetPageView("users.SelfDetail").
			WithMethod(http.MethodGet, SelfDetailHandler).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithQueryPatcher("users.self_detail", SelfUserQueryPatcher))

	// Self update view (only current user)
	lago.RegistryView.Register("users.SelfUpdateView",
		lago.GetPageView("users.SelfUpdateForm").
			WithMethod(http.MethodGet, SelfDetailHandler).
			WithMethod(http.MethodPost, SelfUpdateHandler).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithQueryPatcher("users.self_update", SelfUserQueryPatcher))

	// Self change password (only current user)
	lago.RegistryView.Register("users.SelfChangePasswordView",
		lago.GetPageView("users.SelfChangePasswordForm").
			WithMethod(http.MethodGet, SelfDetailHandler).
			WithMethod(http.MethodPost, SelfChangePasswordHandler).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithQueryPatcher("users.self_change_password", SelfUserQueryPatcher))

	// Delete view
	lago.RegistryView.Register("users.DeleteView",
		views.DetailView[User]("user", "id")(
			views.DeleteView[User]("id", lago.GetterRoutePath("users.ListRoute", nil))(
				lago.GetPageView("users.UserDeleteForm"))).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Change password view (user-specific handler)
	lago.RegistryView.Register("users.ChangePasswordView",
		views.DetailView[User]("user", "id")(
			lago.GetPageView("users.ChangePasswordForm").
				WithMethod(http.MethodPost, ChangePasswordHandler)).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Selection views
	lago.RegistryView.Register("users.SelectView",
		views.ListView[User]("users")(
			lago.GetPageView("users.UserSelectionTable")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	lago.RegistryView.Register("users.RoleSelectView",
		views.ListView[Role]("roles")(
			lago.GetPageView("users.RoleSelectionTable")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Role CRUD views
	lago.RegistryView.Register("users.RoleListView",
		views.ListView[Role]("roles")(
			lago.GetPageView("users.RoleTable")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	lago.RegistryView.Register("users.RoleDetailView",
		views.DetailView[Role]("role", "id")(
			lago.GetPageView("users.RoleDetail")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	lago.RegistryView.Register("users.RoleCreateView",
		views.CreateView[Role](lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("users.RoleCreateForm")).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	lago.RegistryView.Register("users.RoleUpdateView",
		views.DetailView[Role]("role", "id")(
			views.UpdateView[Role]("id", lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("users.RoleUpdateForm"))).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	lago.RegistryView.Register("users.RoleDeleteView",
		views.DetailView[Role]("role", "id")(
			views.DeleteView[Role]("id", lago.GetterRoutePath("users.RoleListRoute", nil))(
				lago.GetPageView("users.RoleDeleteForm"))).
			WithMiddleware("users.auth", AuthenticationMiddleware).
			WithMiddleware("users.role", RoleAuthorizationMiddleware([]string{""})))

	// Auth views
	lago.RegistryView.Register("users.LogoutView",
		lago.GetPageView("users.UnauthenticatedPage").
			WithMethod(http.MethodPost, LogoutHandler).
			WithMethod(http.MethodGet, LogoutHandler))

	lago.RegistryView.Register("users.LoginView",
		lago.GetPageView("users.LoginPage").
			WithMethod(http.MethodPost, LoginHandler))

	lago.RegistryView.Register("users.SignupView",
		lago.GetPageView("users.SignupPage").
			WithMethod(http.MethodPost, SignupHandler))

	lago.RegistryView.Register("base.HomeView",
		lago.NewRedirectView("users.LoginRoute"))

	lago.RegistryView.Register("users.LoginSuccessView",
		lago.NewRedirectView("users.LoginRoute"))

	lago.RegistryView.Register("users.UnauthenticatedView",
		lago.GetPageView("users.UnauthenticatedPage"))
}
