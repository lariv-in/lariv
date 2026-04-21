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
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
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

	db, dberr := getters.DBFromContext(r.Context())
	if dberr != nil {
		slog.Warn("resolveAuth: db from context", "err", dberr)
		return nil
	}
	user, err := gorm.G[User](db).Where("id = ?", uint(userID)).Last(r.Context())
	if err != nil {
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

// AuthenticationLayer requires a valid auth token. If the user is not
// authenticated the request is redirected to the unauthenticated route.
type AuthenticationLayer struct{}

func (AuthenticationLayer) Next(_ views.View, next http.Handler) http.Handler {
	return RequireAuth(next)
}

// RequireAuth wraps a handler so it only runs for authenticated requests (same
// cookie rules as [AuthenticationLayer]). Unauthenticated requests are redirected
// to the unauthenticated route.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := resolveAuth(r)
		if ctx == nil {
			unauthenticatedRoute, _ := lago.RegistryRoute.Get("users.UnauthenticatedRoute")
			views.HtmxRedirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthLayer enriches the request context with $user, $role, and
// $tz when a valid auth token is present. If the user is not authenticated the
// request continues without those context values.
type OptionalAuthLayer struct{}

func (OptionalAuthLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ctx := resolveAuth(r); ctx != nil {
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

type RoleAuthorizationLayer struct {
	Roles []string
}

func (m RoleAuthorizationLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context(), "RoleAuthorizationLayer")

		var roleName string
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("RoleAuthorizationLayer: db from context", "error", dberr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		db.Model(&Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)

		authorized := slices.Contains(m.Roles, roleName)
		if user.IsSuperuser {
			authorized = true
		}

		if !authorized {
			slog.Error("RoleAuthorizationLayer: user is not authorized", "role", roleName, "roles", m.Roles)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
