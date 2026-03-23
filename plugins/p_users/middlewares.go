package p_users

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// resolveAuth attempts to authenticate the request from the auth-token cookie.
// On success it returns a context enriched with $user, $role, and $tz.
// On failure it returns nil and the reason is logged.
func resolveAuth(r *http.Request) context.Context {
	authCookie, err := r.Cookie("auth-token")
	if err != nil {
		return nil
	}

	token, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (any, error) {
		return signingKey, nil
	}, jwt.WithAllAudiences("lariv-"+base64.StdEncoding.EncodeToString(jwtIssuer)),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer("lariv"),
		jwt.WithNotBeforeRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithLeeway(time.Hour*24),
	)
	if err != nil {
		slog.Warn("Error while parsing token", "err", err)
		return nil
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		slog.Warn("Error while getting subject", "err", err)
		return nil
	}

	userID, err := strconv.ParseInt(strings.Split(subject, "-")[0], 10, 32)
	if err != nil {
		slog.Warn("Error while parsing user id", "err", err)
		return nil
	}

	db, ok := r.Context().Value("$db").(*gorm.DB)
	if !ok || db == nil {
		slog.Warn("resolveAuth: missing $db in context")
		return nil
	}
	var user User
	if err := db.Model(User{}).Last(&user, "ID = ?", userID).Error; err != nil {
		slog.Warn("Error while loading user", "err", err)
		return nil
	}
	var roleName string
	if user.IsSuperuser {
		roleName = "superuser"
	} else {
		db.Model(&Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	}

	ctx := context.WithValue(r.Context(), "$user", user)
	ctx = context.WithValue(ctx, "$role", roleName)
	timezone, err := time.LoadLocation(user.Timezone)
	if err != nil {
		slog.Warn("Invalid timezone for user", "error", err)
	} else {
		ctx = context.WithValue(ctx, "$tz", timezone)
	}
	return ctx
}

// AuthenticationMiddleware requires a valid auth token. If the user is not
// authenticated the request is redirected to the unauthenticated route.
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := resolveAuth(r)
		if ctx == nil {
			unauthenticatedRoute, _ := lago.RegistryRoute.Get("users.UnauthenticatedRoute")
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware enriches the request context with $user, $role, and
// $tz when a valid auth token is present. If the user is not authenticated the
// request continues without those context values.
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ctx := resolveAuth(r); ctx != nil {
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func RoleAuthorizationMiddleware(roles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userObj := r.Context().Value("$user")
			user, ok := userObj.(User)
			if !ok {
				slog.Error("RoleAuthorizationMiddleware: missing $user in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var roleName string
			db, ok := r.Context().Value("$db").(*gorm.DB)
			if !ok {
				slog.Error("RoleAuthorizationMiddleware: missing $db in context")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			db.Model(&Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)

			authorized := slices.Contains(roles, roleName)
			if user.IsSuperuser {
				authorized = true
			}

			if !authorized {
				slog.Error("RoleAuthorizationMiddleware: user is not authorized", "role", roleName, "roles", roles)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
