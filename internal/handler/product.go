package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type ProductHandler struct{ repo *repository.ProductRepo }

func NewProductHandler(r *repository.ProductRepo) *ProductHandler {
	return &ProductHandler{repo: r}
}

func (h *ProductHandler) validate(w http.ResponseWriter, p *models.Product) bool {
	if p.Gender != "" && !slices.Contains(models.ValidGenders, p.Gender) {
		writeError(w, http.StatusBadRequest, "Некорректный гендер (M/W/U)")
		return false
	}
	if p.Season != "" && !slices.Contains(models.ValidSeasons, p.Season) {
		writeError(w, http.StatusBadRequest,
			"invalid season; allowed: "+fmt.Sprint(models.ValidSeasons))
		return false
	}
	if p.Price < 0 {
		writeError(w, http.StatusBadRequest, "Цена должна быть > 0")
		return false
	}
	if p.Quantity < 0 {
		writeError(w, http.StatusBadRequest, "Количчество должно быть > 0")
		return false
	}
	return true
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if !h.validate(w, &p) {
		return
	}
	result, merged, err := h.repo.Create(r.Context(), &p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	code := http.StatusCreated
	if merged {
		code = http.StatusOK
	}
	writeJSON(w, code, map[string]any{
		"product":   result,
		"restocked": merged,
	})
}

func (h *ProductHandler) Import(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json: ожидается массив товаров")
		return
	}
	if len(products) == 0 {
		writeError(w, http.StatusBadRequest, "пустой список")
		return
	}

	imported, restocked := 0, 0
	out := make([]models.Product, 0, len(products))
	for i := range products {
		p := &products[i]
		if !h.validate(w, p) {
			return
		}
		result, merged, err := h.repo.Create(r.Context(), p)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, *result)
		if merged {
			restocked++
		} else {
			imported++
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"imported":  imported,
		"restocked": restocked,
		"products":  out,
	})
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.ProductFilter{
		Type:        q.Get("type"),
		Gender:      q.Get("gender"),
		Size:        q.Get("size"),
		Color:       q.Get("color"),
		Brand:       q.Get("brand"),
		Seasons:     q["season"],
		Search:      q.Get("q"),
		SortByPrice: q.Get("sort_price"),
		Limit:       parseInt(q.Get("limit"), 20),
		Offset:      parseInt(q.Get("offset"), 0),
	}
	items, err := h.repo.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Не найдено")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный id")
		return
	}
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный json")
		return
	}
	if !h.validate(w, &p) {
		return
	}
	p.ID = id
	if err := h.repo.Update(r.Context(), &p); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
