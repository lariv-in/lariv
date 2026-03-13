package p_users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// --- Auth Handlers (user-specific, not generalizable) ---

func LoginHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
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
		// Setting the default to true, best if data is not changed in case of failure of assumptions
		userAlreadyExists := true
		db.Model(User{}).Select("Email = ?", email).Find(&userAlreadyExists)
		user := User{
			Name:        name,
			Email:       email,
			Phone:       phone,
			IsSuperuser: false,
			Password:    []byte(password1Str),
			Role: Role{
				Name: "Unassigned",
			},
		}
		err = db.Create(&user).Error

		if err != nil {
			ctx := context.WithValue(r.Context(), views.GlobalContextError, map[string]any{"_form": fmt.Errorf("%v", err)})
			v.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
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

		user := r.Context().Value("user").(User)
		db := r.Context().Value("$db").(*gorm.DB)

		user.Password = []byte(newPassword)
		err = db.Save(&user).Error
		if err != nil {
			ctx := context.WithValue(r.Context(), views.GlobalContextError, map[string]any{"_form": fmt.Errorf("%v", err)})
			v.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%s%d/", AppUrl, user.ID), http.StatusSeeOther)
	})
}

// --- View Registrations ---

func init() {
	// List view
	lago.RegistryView.Register("users.ListView",
		AuthMiddleware(
			views.ListView[User]("users")(
				lago.GetPageView("users.UserTable"))))

	// Detail view
	lago.RegistryView.Register("users.DetailView",
		AuthMiddleware(
			views.DetailView[User]("user")(
				lago.GetPageView("users.UserDetail"))))

	// Create view
	lago.RegistryView.Register("users.CreateView",
		AuthMiddleware(
			views.CreateView[User](AppUrl+"%v/")(
				lago.GetPageView("users.UserCreateForm"))))

	// Update view
	lago.RegistryView.Register("users.UpdateView",
		AuthMiddleware(
			views.DetailView[User]("user")(
				views.UpdateView[User](AppUrl+"%v/")(
					lago.GetPageView("users.UserUpdateForm")))))

	// Delete view
	lago.RegistryView.Register("users.DeleteView",
		AuthMiddleware(
			views.DetailView[User]("user")(
				views.DeleteView[User](AppUrl)(
					lago.GetPageView("users.UserDeleteForm")))))

	// Change password view (user-specific handler)
	lago.RegistryView.Register("users.ChangePasswordView",
		AuthMiddleware(
			views.DetailView[User]("user")(
				lago.GetPageView("users.ChangePasswordForm").
					WithMethod(http.MethodPost, ChangePasswordHandler))))

	// Selection views
	lago.RegistryView.Register("users.SelectView",
		AuthMiddleware(
			views.ListView[User]("users")(
				lago.GetPageView("users.UserSelectionTable"))))

	lago.RegistryView.Register("users.MultiSelectView",
		AuthMiddleware(
			views.ListView[User]("users")(
				lago.GetPageView("users.UserMultiSelectionTable"))))

	lago.RegistryView.Register("users.RoleSelectView",
		AuthMiddleware(
			views.ListView[Role]("roles")(
				lago.GetPageView("users.RoleSelectionTable"))))

	lago.RegistryView.Register("users.RoleMultiSelectView",
		AuthMiddleware(
			views.ListView[Role]("roles")(
				lago.GetPageView("users.RoleMultiSelectionTable"))))

	// Role CRUD views
	lago.RegistryView.Register("users.RoleListView",
		AuthMiddleware(
			views.ListView[Role]("roles")(
				lago.GetPageView("users.RoleTable"))))

	lago.RegistryView.Register("users.RoleDetailView",
		AuthMiddleware(
			views.DetailView[Role]("role")(
				lago.GetPageView("users.RoleDetail"))))

	lago.RegistryView.Register("users.RoleCreateView",
		AuthMiddleware(
			views.CreateView[Role](RoleUrl+"%v/")(
				lago.GetPageView("users.RoleCreateForm"))))

	lago.RegistryView.Register("users.RoleUpdateView",
		AuthMiddleware(
			views.DetailView[Role]("role")(
				views.UpdateView[Role](RoleUrl+"%v/")(
					lago.GetPageView("users.RoleUpdateForm")))))

	lago.RegistryView.Register("users.RoleDeleteView",
		AuthMiddleware(
			views.DetailView[Role]("role")(
				views.DeleteView[Role](RoleUrl)(
					lago.GetPageView("users.RoleDeleteForm")))))

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
