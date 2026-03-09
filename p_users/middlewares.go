package p_users

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lariv-in/lago"
	"gorm.io/gorm"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("auth-token")
		unauthenticatedRoute, _ := lago.RegistryRoute.Get("users.UnauthenticatedRoute")
		if err != nil {
			slog.Warn("Auth Cookie not found", "err", err)
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}
		token, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (any, error) {
			return signingKey[:], nil
		}, jwt.WithAllAudiences("lariv-"+base64.StdEncoding.EncodeToString(jwtIssuer[:])),
			jwt.WithExpirationRequired(),
			jwt.WithIssuer("lariv"),
			jwt.WithNotBeforeRequired(),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
			jwt.WithLeeway(time.Hour*24),
		)
		if err != nil {
			slog.Warn("Error while parsing token", "err", err)
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}
		subject, err := token.Claims.GetSubject()
		if err != nil {
			slog.Warn("Error while getting subject", "err", err)
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}

		userId, err := strconv.ParseInt(strings.Split(subject, "-")[0], 10, 32)
		if err != nil {
			slog.Warn("Error while parsing user id", "err", err)
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}

		db := r.Context().Value("$db").(*gorm.DB)
		var user User
		err = db.Model(User{}).Last(&user, "ID = ?", userId).Error
		if err != nil {
			slog.Warn("Error while parsing user id", "err", err)
			http.Redirect(w, r, unauthenticatedRoute.Path, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "$user", user)))
	})
}
