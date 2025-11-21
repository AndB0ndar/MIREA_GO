package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Validator interface {
	SignAccessToken(userID int64, email, role string) (string, error)
	SignRefreshToken(userID int64) (string, error)
	Parse(tokenStr string) (jwt.MapClaims, error)
}

type HS256 struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewHS256(secret []byte, accessTTL, refreshTTL time.Duration) *HS256 {
	return &HS256{
		secret:          secret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (h *HS256) SignAccessToken(userID int64, email, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"iat":   now.Unix(),
		"exp":   now.Add(h.accessTokenTTL).Unix(),
		"iss":   "app-auth",
		"aud":   "app-clients",
		"type":  "access",
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.secret)
}

func (h *HS256) SignRefreshToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"iat":  now.Unix(),
		"exp":  now.Add(h.refreshTokenTTL).Unix(),
		"iss":  "app-auth",
		"aud":  "app-clients",
		"type": "refresh",
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.secret)
}

func (h *HS256) Parse(tokenStr string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return h.secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithAudience("app-clients"),
		jwt.WithIssuer("app-auth"),
	)

	if err != nil || !t.Valid {
		return nil, err
	}

	return t.Claims.(jwt.MapClaims), nil
}
