package p_users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

const usersTable = "users"
const rolesTable = "roles"

// --- Auth Handlers (user-specific, not generalizable) ---

func LoginHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()
		forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
		if len(forms) == 0 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		form := forms[0]
		values, fieldErrors, err := form.ParseForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hasErrors := false
		ctx := r.Context()
		for name, fieldErr := range fieldErrors {
			if fieldErr != nil {
				hasErrors = true
				ctx = context.WithValue(ctx, "$error."+name, fieldErr)
			}
		}

		if hasErrors {
			for name, value := range values {
				ctx = context.WithValue(ctx, "$in."+name, value)
			}
			page.Build(ctx).Render(w)
			return
		}

		email, _ := values["email"].(string)
		password, _ := values["password"].(string)

		db := r.Context().Value("$db").(*gorm.DB)
		user, err := Authenticate(db, email, password)
		if err != nil {
			ctx = context.WithValue(ctx, "$error.password", fmt.Errorf("Invalid email or password"))
			for name, value := range values {
				ctx = context.WithValue(ctx, "$in."+name, value)
			}
			page.Build(ctx).Render(w)
			return
		}
		user.Login(w)
		lago.NewRedirectView("users.LoginSuccessRoute").ServeHTTP(w, r)
	})
}

func SignupHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := v.GetPage()
		forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
		if len(forms) == 0 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		form := forms[0]
		values, fieldErrors, err := form.ParseForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		hasErrors := false
		for name, fieldErr := range fieldErrors {
			if fieldErr != nil {
				hasErrors = true
				ctx = context.WithValue(ctx, "$error."+name, fieldErr)
			}
		}

		password1Str, _ := values["password1"].(string)
		password2Str, _ := values["password2"].(string)

		if password1Str != password2Str {
			hasErrors = true
			ctx = context.WithValue(ctx, "$error.password2", fmt.Errorf("Passwords do not match"))
		}

		if hasErrors {
			for name, value := range values {
				ctx = context.WithValue(ctx, "$in."+name, value)
			}
			page.Build(ctx).Render(w)
			return
		}

		name, _ := values["name"].(string)
		email, _ := values["email"].(string)
		phone, _ := values["phone"].(string)

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
			http.Error(w, err.Error(), http.StatusBadRequest)
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
		page, _ := v.GetPage()
		forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
		if len(forms) == 0 {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		form := forms[0]
		values, fieldErrors, err := form.ParseForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		hasErrors := false
		for name, fieldErr := range fieldErrors {
			if fieldErr != nil {
				hasErrors = true
				ctx = context.WithValue(ctx, "$error."+name, fieldErr)
			}
		}

		newPassword, _ := values["new_password"].(string)
		confirmPassword, _ := values["confirm_password"].(string)

		if newPassword != confirmPassword {
			hasErrors = true
			ctx = context.WithValue(ctx, "$error.confirm_password", fmt.Errorf("Passwords do not match"))
		}

		if hasErrors {
			page.Build(ctx).Render(w)
			return
		}

		user := r.Context().Value("user").(User)
		db := r.Context().Value("$db").(*gorm.DB)

		user.Password = []byte(newPassword)
		err = db.Save(&user).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			views.ListView(usersTable, "users")(
				lago.GetPageView("users.UserTable"))))

	// Detail view
	lago.RegistryView.Register("users.DetailView",
		AuthMiddleware(
			views.DetailView(usersTable, "user")(
				lago.GetPageView("users.UserDetail"))))

	// Create view
	lago.RegistryView.Register("users.CreateView",
		AuthMiddleware(
			views.CreateView(usersTable, AppUrl+"%v/")(
				lago.GetPageView("users.UserCreateForm"))))

	// Update view
	lago.RegistryView.Register("users.UpdateView",
		AuthMiddleware(
			views.DetailView(usersTable, "user")(
				views.UpdateView(usersTable, AppUrl+"%v/")(
					lago.GetPageView("users.UserUpdateForm")))))

	// Delete view
	lago.RegistryView.Register("users.DeleteView",
		AuthMiddleware(
			views.DetailView(usersTable, "user")(
				views.DeleteView(usersTable, AppUrl)(
					lago.GetPageView("users.UserDeleteForm")))))

	// Change password view (user-specific handler)
	lago.RegistryView.Register("users.ChangePasswordView",
		AuthMiddleware(
			views.DetailView(usersTable, "user")(
				lago.GetPageView("users.ChangePasswordForm").
					WithMethod(http.MethodPost, ChangePasswordHandler))))

	// Selection views
	lago.RegistryView.Register("users.SelectView",
		AuthMiddleware(
			views.ListView(usersTable, "users")(
				lago.GetPageView("users.UserSelectionTable"))))

	lago.RegistryView.Register("users.MultiSelectView",
		AuthMiddleware(
			views.ListView(usersTable, "users")(
				lago.GetPageView("users.UserMultiSelectionTable"))))

	lago.RegistryView.Register("users.RoleSelectView",
		AuthMiddleware(
			views.ListView(rolesTable, "roles")(
				lago.GetPageView("users.RoleSelectionTable"))))

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
