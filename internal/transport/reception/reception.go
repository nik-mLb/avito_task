package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
)

//go:generate mockgen -source=reception.go -destination=../../usecase/mocks/reception_usecase_mock.go -package=mocks ReceptionUsecase
type ReceptionUsecase interface {
	CreateReception(ctx context.Context, pvzID string) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID string) (*models.Reception, error)
}

type ReceptionHandler struct {
	uc ReceptionUsecase
}

func NewReceptionHandler(uc ReceptionUsecase) *ReceptionHandler {
	return &ReceptionHandler{uc: uc}
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	const op = "ReceptionHandler.CreateReception"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	
	var req dto.ReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("invalid request body")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	reception, err := h.uc.CreateReception(r.Context(), req.PickupPointID)
	if err != nil {
		logger.WithError(err).Warn("failed to create reception")
		switch err {
		case errs.ErrActiveReceptionExists:
			response.SendError(r.Context(), w, http.StatusBadRequest, "Active reception already exists")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to create reception")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, reception)
}

func (h *ReceptionHandler) CloseReception(w http.ResponseWriter, r *http.Request) {
	const op = "ReceptionHandler.CloseReception"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

    vars := mux.Vars(r)
    pvzID := vars["pvzId"]

    reception, err := h.uc.CloseReception(r.Context(), pvzID)
    if err != nil {
		logger.WithError(err).Warn("failed to close reception")
		switch err {
		case errs.ErrNoActiveReceptionToClose:
			response.SendError(r.Context(), w, http.StatusBadRequest, "No active reception to close")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to close reception")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, reception)
}