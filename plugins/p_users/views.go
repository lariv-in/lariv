package p_users

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// --- Auth Handlers (user-specific, not generalizable) ---

func LoginHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		if views.HasErrors(fieldErrors) {
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

func SignupHandler(v views.View) http.Handler {
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

		if views.HasErrors(fieldErrors) {
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

		if views.HasErrors(fieldErrors) {
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

func LogoutHandler(v views.View) http.Handler {
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
func ChangePasswordHandler(v views.View) http.Handler {
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

		if views.HasErrors(fieldErrors) {
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

		var targetUser User

		err = db.Model(User{}).Last(&targetUser, "ID = ?", id).Error
		if err != nil {
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		targetUser.Password = []byte(newPassword)
		err = db.Save(&targetUser).Error
		if err != nil {
			fieldErrors["_form"] = fmt.Errorf("%v", err)
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		ctx := context.WithValue(r.Context(), "$id", fmt.Sprintf("%d", targetUser.ID))
		lago.NewRedirectView("users.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterKey("$id"),
		}).ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- View Registrations ---

func init() {
	// List view
	lago.RegistryView.Register("users.ListView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[User]("users")(
					lago.GetPageView("users.UserTable")))))

	// Detail view
	lago.RegistryView.Register("users.DetailView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[User]("user")(
					lago.GetPageView("users.UserDetail")))))

	// Create view
	lago.RegistryView.Register("users.CreateView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.CreateView[User](lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
					lago.GetPageView("users.UserCreateForm")))))

	// Update view
	lago.RegistryView.Register("users.UpdateView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[User]("user")(
					views.UpdateView[User](lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
						lago.GetPageView("users.UserUpdateForm"))))))

	// Delete view
	lago.RegistryView.Register("users.DeleteView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[User]("user")(
					views.DeleteView[User](lago.GetterRoutePath("users.ListRoute", nil))(
						lago.GetPageView("users.UserDeleteForm"))))))

	// Change password view (user-specific handler)
	lago.RegistryView.Register("users.ChangePasswordView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[User]("user")(
					lago.GetPageView("users.ChangePasswordForm").
						WithMethod(http.MethodPost, ChangePasswordHandler)))))

	// Selection views
	lago.RegistryView.Register("users.SelectView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[User]("users")(
					lago.GetPageView("users.UserSelectionTable")))))

	lago.RegistryView.Register("users.MultiSelectView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[User]("users")(
					lago.GetPageView("users.UserMultiSelectionTable")))))

	lago.RegistryView.Register("users.RoleSelectView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[Role]("roles")(
					lago.GetPageView("users.RoleSelectionTable")))))

	lago.RegistryView.Register("users.RoleMultiSelectView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[Role]("roles")(
					lago.GetPageView("users.RoleMultiSelectionTable")))))

	// Role CRUD views
	lago.RegistryView.Register("users.RoleListView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.ListView[Role]("roles")(
					lago.GetPageView("users.RoleTable")))))

	lago.RegistryView.Register("users.RoleDetailView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[Role]("role")(
					lago.GetPageView("users.RoleDetail")))))

	lago.RegistryView.Register("users.RoleCreateView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.CreateView[Role](lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
					lago.GetPageView("users.RoleCreateForm")))))

	lago.RegistryView.Register("users.RoleUpdateView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[Role]("role")(
					views.UpdateView[Role](lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
						lago.GetPageView("users.RoleUpdateForm"))))))

	lago.RegistryView.Register("users.RoleDeleteView",
		AuthenticationMiddleware(
			RoleAuthorizationMiddleware([]string{""})(
				views.DetailView[Role]("role")(
					views.DeleteView[Role](lago.GetterRoutePath("users.RoleListRoute", nil))(
						lago.GetPageView("users.RoleDeleteForm"))))))

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
