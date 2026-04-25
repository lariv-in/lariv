package p_users

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/lariv-in/lago/lago"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

func HashPassword(password, passwordSalt []byte) []byte {
	key, err := scrypt.Key(password, passwordSalt, 32768, 8, 1, 32)
	if err != nil {
		panic("According to the docs for scrypt, this should be impossible")
	}
	return key
}

func Authenticate(db *gorm.DB, email, password string) (*User, error) {
	user, err := gorm.G[User](db).Where("email = ?", email).Last(context.Background())
	if err != nil {
		return nil, err
	}

	passwordKey := HashPassword([]byte(password), user.PasswordSalt)
	if !bytes.Equal(passwordKey, user.PasswordHash) {
		return nil, errors.New("Could not authenticate user")
	}

	return new(user), nil
}

type AuthConfig struct {
	SigningKey string `toml:"signingKey"`
	JwtIssuer  string `toml:"jwtIssuer"`
}

var Config = &AuthConfig{}

var (
	signingKey []byte
	jwtIssuer  []byte
)

func init() {
	// Default to randomized keys so that when not configured,
	// sessions are invalidated on every server restart.
	signingKey = make([]byte, 64)
	jwtIssuer = make([]byte, 64)
	_, _ = rand.Read(signingKey)
	_, _ = rand.Read(jwtIssuer)

	lago.RegistryConfig.Register("p_users", Config)
}

func (c *AuthConfig) PostConfig() {
	if c.SigningKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.SigningKey)
		if err != nil {
			log.Panicf("Signing Key specified in config is invalid %s\n", c.SigningKey)
		}
		signingKey = decoded
	}

	if c.JwtIssuer != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.JwtIssuer)
		if err != nil {
			log.Panicf("JwtIssuer specified in config is invalid %s\n", c.SigningKey)
		}
		jwtIssuer = decoded
	}
}

func (u *User) GetClaims(currentTime, expiryTime time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    "lariv",
		Subject:   fmt.Sprintf("%d-%s", u.ID, base64.StdEncoding.EncodeToString(u.PasswordSalt)),
		Audience:  jwt.ClaimStrings{"lariv-" + base64.StdEncoding.EncodeToString(jwtIssuer)},
		ExpiresAt: jwt.NewNumericDate(expiryTime),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		NotBefore: jwt.NewNumericDate(currentTime),
		ID:        uuid.New().String(),
	}
}

func (u *User) GetJwt(currentTime, expiryTime time.Time) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, u.GetClaims(currentTime, expiryTime)).SignedString(signingKey[:])
}

func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now()
	nextDayTime := currentTime.Add(time.Hour * 24)
	jwt, err := u.GetJwt(currentTime, nextDayTime)
	if err != nil {
		http.Error(w, "Could not authenticate the user", http.StatusForbidden)
		return
	}
	secure := r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
	http.SetCookie(w, &http.Cookie{
		Name:     "auth-token",
		Value:    jwt,
		Expires:  nextDayTime,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}
