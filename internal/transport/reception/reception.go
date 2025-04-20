package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
	repository "github.com/nik-mLb/avito_task/internal/repository/reception"
)

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

type createReceptionRequest struct {
	PvzID string `json:"pvzId"`
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req createReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reception, err := h.uc.CreateReception(r.Context(), req.PvzID)
	if err != nil {
		switch err {
		case repository.ErrActiveReceptionExists:
			http.Error(w, "Active reception already exists", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create reception", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reception)
}

func (h *ReceptionHandler) CloseReception(w http.ResponseWriter, r *http.Request) {
    // Получаем pvzId из URL
    vars := mux.Vars(r)
    pvzID := vars["pvzId"]

    reception, err := h.uc.CloseReception(r.Context(), pvzID)
    if err != nil {
        switch err {
        case repository.ErrNoActiveReceptionToClose:
            http.Error(w, "No active reception to close", http.StatusBadRequest)
        default:
            http.Error(w, "Failed to close reception", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(reception)
}