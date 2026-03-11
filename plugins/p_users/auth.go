package p_users

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/lariv-in/lago"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

func HashPassword(password []byte, passwordSalt []byte) []byte {
	key, err := scrypt.Key(password, passwordSalt, 32768, 8, 1, 32)
	if err != nil {
		panic("According to the docs for scrypt, this should be impossible")
	}
	return key
}

func Authenticate(db *gorm.DB, email string, password string) (*User, error) {
	var salt [][]byte
	err := db.Model(User{}).Select("PasswordSalt").Last(&salt, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	if len(salt) == 0 {
		return nil, errors.New("Could not authenticate user")
	}
	passwordKey := HashPassword([]byte(password), []byte(salt[0]))

	var user User

	err = db.Last(&user, "Password = ?", passwordKey, "Email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type AuthConfig struct {
	SigningKey string `toml:"signingKey"`
	JwtIssuer  string `toml:"jwtIssuer"`
}

var Config = &AuthConfig{}

var signingKey [256]byte
var jwtIssuer [256]byte

func init() {
	rand.Read(signingKey[:])
	rand.Read(jwtIssuer[:])

	lago.RegistryConfig.Register("p_users", Config)
}

func (c *AuthConfig) PostConfig() {
	if c.SigningKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.SigningKey)
		if err == nil {
			copy(signingKey[:], decoded)
		}
	}

	if c.JwtIssuer != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.JwtIssuer)
		if err == nil {
			copy(jwtIssuer[:], decoded)
		}
	}
}

func (u *User) GetClaims(currentTime time.Time, expiryTime time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    "lariv",
		Subject:   fmt.Sprintf("%d-%s", u.ID, base64.StdEncoding.EncodeToString(u.PasswordSalt)),
		Audience:  jwt.ClaimStrings{"lariv-" + base64.StdEncoding.EncodeToString(jwtIssuer[:])},
		ExpiresAt: jwt.NewNumericDate(expiryTime),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		NotBefore: jwt.NewNumericDate(currentTime),
		ID:        uuid.New().String(),
	}
}

func (u *User) GetJwt(currentTime time.Time, expiryTime time.Time) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, u.GetClaims(currentTime, expiryTime)).SignedString(signingKey[:])
}

func (u *User) Login(w http.ResponseWriter) {
	currentTime := time.Now()
	nextDayTime := currentTime.Add(time.Hour * 24)
	jwt, err := u.GetJwt(currentTime, nextDayTime)
	if err != nil {
		http.Error(w, "Could not authenticate the user", http.StatusForbidden)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth-token",
		Value:    jwt,
		Expires:  nextDayTime,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
}
