package handler

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"clothingStore/internal/middleware"
	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type OrderHandler struct{ repo *repository.OrderRepo }

func NewOrderHandler(r *repository.OrderRepo) *OrderHandler { return &OrderHandler{repo: r} }

type createOrderReq struct {
	Items []orderItemInput `json:"items"`
}

type orderItemInput struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "Список товаров")
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	order := &models.Order{UserID: userID, Status: models.StatusAccepted}
	items := make([]models.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, models.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity})
	}
	id, err := h.repo.CreateWithItems(r.Context(), order, items)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	order.ID = id
	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}
	items, err := h.repo.ListByUser(r.Context(), userID,
		parseInt(r.URL.Query().Get("limit"), 20),
		parseInt(r.URL.Query().Get("offset"), 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.OrderFilter{
		Status: q.Get("status"),
		Limit:  parseInt(q.Get("limit"), 20),
		Offset: parseInt(q.Get("offset"), 0),
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = t
		}
	}
	items, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	roleID, _ := middleware.RoleIDFromContext(r.Context())

	if roleID == models.RoleClient {
		o, err := h.repo.GetByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Не найденр")
			return
		}
		if o.UserID != userID {
			writeError(w, http.StatusForbidden, "Не авторизован")
			return
		}
	}

	d, err := h.repo.GetDetail(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Не найдено")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var o models.Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if !slices.Contains(models.AllowedToChange, o.Status) {
		writeError(w, http.StatusBadRequest,
			"этот статус нельзя установить вручную; для отмены используйте кнопку")
		return
	}
	if err := h.repo.UpdateStatus(r.Context(), id, o.Status); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	o.ID = id
	writeJSON(w, http.StatusOK, o)
}

type cancelReq struct {
	Reason string `json:"reason"`
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный идентификатор")
		return
	}

	var req cancelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "Укажите причину отмены")
		return
	}
	if len([]rune(req.Reason)) > 255 {
		writeError(w, http.StatusBadRequest, "Причина отмены слишком длинная (макс. 255 символов)")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}
	roleID, _ := middleware.RoleIDFromContext(r.Context())
	isClient := roleID == models.RoleClient
	if err := h.repo.Cancel(r.Context(), id, userID, req.Reason, isClient); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": models.StatusCancelled})
}
