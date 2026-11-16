package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"clothingStore/internal/middleware"
	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

var phoneRe = regexp.MustCompile(`^\+\d{11}$`)
var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

type UserHandler struct{ repo *repository.UserRepo }

func NewUserHandler(r *repository.UserRepo) *UserHandler { return &UserHandler{repo: r} }

type createUserReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	RoleID   string `json:"role_id"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !emailRe.MatchString(req.Email) {
		writeError(w, http.StatusBadRequest, "Укажите корректный email (например, user@example.com)")
		return
	}
	if req.Phone != "" && !phoneRe.MatchString(req.Phone) {
		writeError(w, http.StatusBadRequest, "Телефон формата +XXXXXXXXXXX")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	u := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		RoleID:       req.RoleID,
	}
	id, err := h.repo.Create(r.Context(), u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	u.ID = id
	writeJSON(w, http.StatusCreated, u)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := h.repo.List(r.Context(),
		parseInt(q.Get("limit"), 20), parseInt(q.Get("offset"), 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *UserHandler) ListByRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный role id")
		return
	}
	q := r.URL.Query()
	items, err := h.repo.ListByRole(r.Context(), roleID,
		parseInt(q.Get("limit"), 20), parseInt(q.Get("offset"), 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if !emailRe.MatchString(u.Email) {
		writeError(w, http.StatusBadRequest, "Укажите корректный email (например, user@example.com)")
		return
	}
	if u.Phone != "" && !phoneRe.MatchString(u.Phone) {
		writeError(w, http.StatusBadRequest, "Телефон формата +XXXXXXXXXX")
		return
	}
	u.ID = id
	if err := h.repo.Update(r.Context(), &u); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateRoleReq struct {
	RoleID string `json:"role_id"`
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var req updateRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	validRoles := map[string]bool{
		models.RoleGuest:   true,
		models.RoleClient:  true,
		models.RoleManager: true,
		models.RoleAdmin:   true,
	}
	if !validRoles[req.RoleID] {
		writeError(w, http.StatusBadRequest, "Неизвестный role_id")
		return
	}
	if err := h.repo.UpdateRole(r.Context(), id, req.RoleID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "role_id": req.RoleID})
}
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	u, err := h.repo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type topUpReq struct {
	Amount float64 `json:"amount"`
}

func (h *UserHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	roleID, _ := middleware.RoleIDFromContext(r.Context())

	targetID := userID
	if urlID := chi.URLParam(r, "id"); urlID != "" {
		if roleID != models.RoleAdmin {
			writeError(w, http.StatusForbidden, "Пополнение чужого баланса только для администратора")
			return
		}
		targetID = urlID
	}

	var req topUpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if req.Amount <= 0 || req.Amount > 1_000_000 {
		writeError(w, http.StatusBadRequest, "Сумма от 0 до 1.000.000")
		return
	}

	bal, err := h.repo.AddBalance(r.Context(), targetID, req.Amount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": targetID,
		"balance": bal,
	})
}
