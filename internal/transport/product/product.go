package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
)

//go:generate mockgen -source=product.go -destination=../../usecase/mocks/product_usecase_mock.go -package=mocks ProductUsecase
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

func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	const op = "ProductHandler.AddProduct"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("invalid request body")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	product, err := h.uc.AddProduct(r.Context(), req.PickupPointID, req.Type)
	if err != nil {
		logger.WithError(err).Warn("failed to add product")
		switch err {
		case errs.ErrNoActiveReception:
			response.SendError(r.Context(), w, http.StatusBadRequest, "No active reception found")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to add product")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, product)
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
    const op = "ProductHandler.DeleteLastProduct"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

    vars := mux.Vars(r)
    pvzID := vars["pvzId"]

    err := h.uc.DeleteLastProduct(r.Context(), pvzID)
    if err != nil {
		logger.WithError(err).Warn("failed to delete last product")
		switch err {
		case errs.ErrNoProductsToDelete:
			response.SendError(r.Context(), w, http.StatusBadRequest, "No active reception found or no products to delete")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to delete product")
		}
		return
	}

    w.WriteHeader(http.StatusOK)
}