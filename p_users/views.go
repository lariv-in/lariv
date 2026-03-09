package p_users

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/views"
	"github.com/nyaruka/phonenumbers"
	"gorm.io/gorm"
)

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

		email, _ := values["email"].(*mail.Address)
		password, _ := values["password"].(string)

		db := r.Context().Value("$db").(*gorm.DB)
		user, err := Authenticate(db, email.Address, password)
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
		emailObj, _ := values["email"].(*mail.Address)
		email := emailObj.Address
		phone, _ := values["phone"].(*phonenumbers.PhoneNumber)

		db := r.Context().Value("$db").(*gorm.DB)
		// Setting the default to true, best if data is not changed in case of failure of assumptions
		userAlreadyExists := true
		db.Model(User{}).Select("Email = ?", email).Find(&userAlreadyExists)
		user := User{
			Name:        name,
			Email:       email,
			Phone:       phonenumbers.Format(phone, phonenumbers.E164),
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

func renderAllUsers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := r.Context().Value("$db").(*gorm.DB)
		var users []User
		err := db.Preload("Role").Find(&users).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "users", users)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func init() {
	lago.RegistryView.Register("users.AllUsersView",
		AuthMiddleware(renderAllUsers(lago.GetPageView("users.AllUsersPage"))))

	lago.RegistryView.Register("users.LogoutView",
		lago.GetPageView("users.UnauthenticatedPage").
			WithMethod(http.MethodPost, LogoutHandler).
			WithMethod(http.MethodGet, LogoutHandler))

	lago.RegistryView.Register("users.LoginView",
		lago.GetPageView("users.LoginPage").
			WithMethod(
				http.MethodPost,
				LoginHandler,
			),
	)

	lago.RegistryView.Register("users.SignupView",
		lago.GetPageView("users.SignupPage").
			WithMethod(
				http.MethodPost,
				SignupHandler,
			),
	)

	lago.RegistryView.Register("base.HomeView",
		lago.NewRedirectView("users.LoginRoute"))

	lago.RegistryView.Register("users.LoginSuccessView",
		lago.NewRedirectView("users.LoginRoute"))

	lago.RegistryView.Register("users.UnauthenticatedView",
		lago.GetPageView("users.UnauthenticatedPage"))
}
