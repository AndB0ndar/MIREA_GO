package core

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"app/internal/repo"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type userRepo interface {
	CheckPassword(email, pass string) (repo.UserRecord, error)
	ByID(id int64) (repo.UserRecord, error)
}

type blacklist interface {
	Add(token string, expiresAt time.Time)
	IsRevoked(token string) bool
}

type jwtSigner interface {
	SignAccessToken(userID int64, email, role string) (string, error)
	SignRefreshToken(userID int64) (string, error)
	Parse(tokenStr string) (jwt.MapClaims, error)
}

type Service struct {
	repo      userRepo
	blacklist blacklist
	jwt       jwtSigner
}

func NewService(r userRepo, b blacklist, j jwtSigner) *Service {
	return &Service{repo: r, blacklist: b, jwt: j}
}

type ctxKey string

const ClaimsKey ctxKey = "claims"

func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var in LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Email == "" || in.Password == "" {
		httpError(w, 400, "invalid_credentials", "Email and password required")
		return
	}

	u, err := s.repo.CheckPassword(in.Email, in.Password)
	if err != nil {
		httpError(w, 401, "unauthorized", "Invalid credentials")
		return
	}

	accessToken, err := s.jwt.SignAccessToken(u.ID, u.Email, u.Role)
	if err != nil {
		httpError(w, 500, "token_error", "Failed to generate access token")
		return
	}

	refreshToken, err := s.jwt.SignRefreshToken(u.ID)
	if err != nil {
		httpError(w, 500, "token_error", "Failed to generate refresh token")
		return
	}

	jsonOK(w, TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (s *Service) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var in RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		httpError(w, 400, "invalid_request", "Refresh token required")
		return
	}

	// Проверяем, не отозван ли токен
	if s.blacklist.IsRevoked(in.RefreshToken) {
		httpError(w, 401, "token_revoked", "Refresh token has been revoked")
		return
	}

	claims, err := s.jwt.Parse(in.RefreshToken)
	if err != nil {
		httpError(w, 401, "invalid_token", "Invalid refresh token")
		return
	}

	// Проверяем тип токена
	if claims["type"] != "refresh" {
		httpError(w, 401, "invalid_token_type", "Not a refresh token")
		return
	}

	// Извлекаем userID
	sub, ok := claims["sub"].(float64)
	if !ok {
		httpError(w, 401, "invalid_token", "Invalid user ID in token")
		return
	}

	userID := int64(sub)

	// Получаем пользователя
	u, err := s.repo.ByID(userID)
	if err != nil {
		httpError(w, 401, "user_not_found", "User not found")
		return
	}

	// Добавляем старый refresh токен в blacklist
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt := time.Unix(int64(exp), 0)
		s.blacklist.Add(in.RefreshToken, expiresAt)
	}

	// Генерируем новую пару токенов
	accessToken, err := s.jwt.SignAccessToken(u.ID, u.Email, u.Role)
	if err != nil {
		httpError(w, 500, "token_error", "Failed to generate access token")
		return
	}

	refreshToken, err := s.jwt.SignRefreshToken(u.ID)
	if err != nil {
		httpError(w, 500, "token_error", "Failed to generate refresh token")
		return
	}

	jsonOK(w, TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(ClaimsKey).(map[string]any)

	jsonOK(w, map[string]any{
		"id":    claims["sub"],
		"email": claims["email"],
		"role":  claims["role"],
	})
}

func (s *Service) AdminStats(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"users":   3,
		"version": "1.0",
		"stats":   "Administrative statistics here",
	})
}

func (s *Service) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID пользователя из URL параметра
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httpError(w, 400, "invalid_user_id", "Invalid user ID format")
		return
	}

	// Получаем пользователя из репозитория
	user, err := s.repo.ByID(userID)
	if err != nil {
		httpError(w, 404, "user_not_found", "User not found")
		return
	}

	// Возвращаем информацию о пользователе (без пароля)
	jsonOK(w, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// В реальной системе здесь можно добавить refresh токен в blacklist
	// Для упрощения просто возвращаем успех
	jsonOK(w, map[string]string{
		"message": "Logged out successfully",
	})
}

// Вспомогательные функции
func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, errorMsg, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorMsg,
		Details: details,
	})
}
