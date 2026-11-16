package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

type PhotoHandler struct {
	repo      *repository.PhotoRepo
	uploadDir string
}

func NewPhotoHandler(r *repository.PhotoRepo, dir string) *PhotoHandler {
	return &PhotoHandler{repo: r, uploadDir: dir}
}

func (h *PhotoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректная форма")
		return
	}
	productID := r.FormValue("product_id")
	if _, err := uuid.Parse(productID); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный UUID")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Файл не выбран")
		return
	}
	defer file.Close()

	name := fmt.Sprintf("p%s_%d%s", productID[:8], time.Now().UnixNano(), filepath.Ext(header.Filename))
	dst, err := os.Create(filepath.Join(h.uploadDir, name))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p := &models.Photo{Filename: name, ProductID: productID}
	id, err := h.repo.Create(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	p.ID = id
	writeJSON(w, http.StatusCreated, p)
}

func (h *PhotoHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	http.ServeFile(w, r, filepath.Join(h.uploadDir, p.Filename))
}

func (h *PhotoHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if _, err := uuid.Parse(productID); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный UUID")
		return
	}
	items, err := h.repo.ListByProduct(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *PhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	_ = os.Remove(filepath.Join(h.uploadDir, p.Filename))
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
