package handler

import (
	"encoding/json"
	"net/http"

	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type RoleHandler struct{ repo *repository.RoleRepo }

func NewRoleHandler(r *repository.RoleRepo) *RoleHandler { return &RoleHandler{repo: r} }

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	id, err := h.repo.Create(r.Context(), &role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	role.ID = id
	writeJSON(w, http.StatusCreated, role)
}

func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var role models.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	role.ID = id
	if err := h.repo.Update(r.Context(), &role); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, role)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
