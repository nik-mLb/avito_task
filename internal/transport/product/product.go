package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	repository "github.com/nik-mLb/avito_task/internal/repository/product"
)

type ProductUsecase interface {
	AddProduct(ctx context.Context, pvzID, productType string) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID string) error
}

type ProductHandler struct {
	uc ProductUsecase
}

func NewProductHandler(uc ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

type addProductRequest struct {
	Type  string `json:"type"`
	PvzID string `json:"pvzId"`
}

func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var req addProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.uc.AddProduct(r.Context(), req.PvzID, req.Type)
	if err != nil {
		switch err {
		case repository.ErrNoActiveReception:
			http.Error(w, "No active reception found", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to add product", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
    // Получаем pvzId из URL
    vars := mux.Vars(r)
    pvzID := vars["pvzId"]

    err := h.uc.DeleteLastProduct(r.Context(), pvzID)
    if err != nil {
        switch err {
        case repository.ErrNoActiveReception:
            http.Error(w, "No active reception found", http.StatusBadRequest)
        case repository.ErrNoProductsToDelete:
            http.Error(w, "No products to delete", http.StatusBadRequest)
        default:
            http.Error(w, "Failed to delete product", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}