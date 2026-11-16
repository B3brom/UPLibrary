package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type AuthHandler struct {
	users     *repository.UserRepo
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthHandler(users *repository.UserRepo, secret string, ttl time.Duration) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: secret, jwtTTL: ttl}
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string   `json:"token"`
	User  authUser `json:"user"`
}

type authUser struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	RoleID  string  `json:"role_id"`
	Balance float64 `json:"balance"`
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}
	h.writeAuthResponse(w, u)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Name == "" || !emailRe.MatchString(req.Email) {
		writeError(w, http.StatusBadRequest, "Укажите имя и корректный email (например, user@example.com)")
		return
	}
	if len([]rune(req.Password)) < 8 {
		writeError(w, http.StatusBadRequest, "Пароль должен содержать минимум 8 символов")
		return
	}
	if req.Phone != "" && !phoneRe.MatchString(req.Phone) {
		writeError(w, http.StatusBadRequest, "Телефон должен быть в формате +XXXXXXXXXXX")
		return
	}

	_, err := h.users.GetByEmail(r.Context(), req.Email)
	if err == nil {
		writeError(w, http.StatusConflict, "Этот email уже зарегистрирован")
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "Не удалось проверить email")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось захешировать пароль")
		return
	}

	u := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		RoleID:       models.RoleClient,
	}
	if _, err := h.users.Create(r.Context(), u); err != nil {
		low := strings.ToLower(err.Error())
		switch {
		case strings.Contains(low, "unique constraint failed"):
			writeError(w, http.StatusConflict, "Email или телефон уже зарегистрированы")
		case strings.Contains(low, "check constraint failed"):
			writeError(w, http.StatusBadRequest, "Укажите корректный email (например, user@example.com)")
		default:
			writeError(w, http.StatusInternalServerError, "Не удалось создать аккаунт")
		}
		return
	}

	created, err := h.users.GetByEmail(r.Context(), u.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Аккаунт создан, но не удалось загрузить данные")
		return
	}
	h.writeAuthResponse(w, created)
}

func (h *AuthHandler) writeAuthResponse(w http.ResponseWriter, u *models.User) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":     u.ID,
		"role_id": u.RoleID,
		"iat":     now.Unix(),
		"exp":     now.Add(h.jwtTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось сформировать токен")
		return
	}
	writeJSON(w, http.StatusOK, loginResp{
		Token: signed,
		User: authUser{
			ID:      u.ID,
			Name:    u.Name,
			Email:   u.Email,
			RoleID:  u.RoleID,
			Balance: u.Balance,
		},
	})
}
