package handler

import (
	"encoding/json"
	"net/http"

	"clothingStore/internal/middleware"
	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type CartHandler struct{ repo *repository.CartRepo }

func NewCartHandler(r *repository.CartRepo) *CartHandler { return &CartHandler{repo: r} }

func (h *CartHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c models.Cart
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if c.Count <= 0 {
		writeError(w, http.StatusBadRequest, "count must be > 0")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	c.UserID = userID
	id, err := h.repo.Create(r.Context(), &c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	c.ID = id
	writeJSON(w, http.StatusCreated, c)
}

func (h *CartHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "user_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный user_id")
		return
	}
	authUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	if authUserID != userID {
		writeError(w, http.StatusForbidden, "Доступ запрещен")
		return
	}
	items, err := h.repo.GetByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *CartHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var c models.Cart
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if c.Count <= 0 {
		writeError(w, http.StatusBadRequest, "Количество должно быть > 0")
		return
	}
	c.ID = id
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	updated, err := h.repo.UpdateForUser(r.Context(), &c, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !updated {
		writeError(w, http.StatusNotFound, "Позиция корзины не найдена")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CartHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	deleted, err := h.repo.DeleteForUser(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "Позиция корзины не найдена")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
