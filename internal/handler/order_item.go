package handler

import (
	"encoding/json"
	"net/http"

	"clothingStore/internal/middleware"
	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type OrderItemHandler struct {
	repo      *repository.OrderItemRepo
	orderRepo *repository.OrderRepo
}

func NewOrderItemHandler(r *repository.OrderItemRepo, or *repository.OrderRepo) *OrderItemHandler {
	return &OrderItemHandler{repo: r, orderRepo: or}
}

func (h *OrderItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var it models.OrderItem
	if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	id, err := h.repo.Create(r.Context(), &it)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	it.ID = id
	writeJSON(w, http.StatusCreated, it)
}

func (h *OrderItemHandler) ListByOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseID(r, "order_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный order_id")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	roleID, _ := middleware.RoleIDFromContext(r.Context())
	if roleID == models.RoleClient {
		order, err := h.orderRepo.GetByID(r.Context(), orderID)
		if err != nil || order.UserID != userID {
			writeError(w, http.StatusForbidden, "Доступ запрещен")
			return
		}
	}

	items, err := h.repo.ListByOrder(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}
