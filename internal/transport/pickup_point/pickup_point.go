package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
)

//go:generate mockgen -source=pickup_point.go -destination=../../usecase/mocks/pickup_point_usecase_mock.go -package=mocks PickupPointUsecase
type PickupPointUsecase interface {
	CreatePickupPoint(ctx context.Context, city string) (*pickup.PickupPoint, error)
	GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]dto.PickupPointListResponse, error)
}

type PickupPointHandler struct {
	uc PickupPointUsecase
}

func NewPickupPointHandler(uc PickupPointUsecase) *PickupPointHandler {
	return &PickupPointHandler{uc: uc}
}

func (h *PickupPointHandler) CreatePickupPoint(w http.ResponseWriter, r *http.Request) {
	const op = "PickupPointHandler.CreatePickupPoint"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)
	
	var req dto.PickupPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("invalid request body")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	PickupPoint, err := h.uc.CreatePickupPoint(r.Context(), req.City)
	if err != nil {
		logger.WithError(err).WithField("city", req.City).Warn("failed to create pickup point")
		switch err {
		case errs.ErrCityNotAllowed:
			response.SendError(r.Context(), w, http.StatusBadRequest, "City not allowed")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to create PickupPoint")
		}
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, PickupPoint)
}

func (h *PickupPointHandler) GetPickupPointsWithReceptions(w http.ResponseWriter, r *http.Request) {
	const op = "PickupPointHandler.GetPickupPointsWithReceptions"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	query := r.URL.Query()

	// Парсим даты
	var startDate, endDate *time.Time
	if startStr := query.Get("startDate"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		} else {
			logger.WithField("startDate", startStr).WithError(err).Warn("invalid startDate")
		}
	}
	if endStr := query.Get("endDate"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
		} else {
			logger.WithField("endDate", endStr).WithError(err).Warn("invalid endDate")
		}
	}

	// Парсим пагинацию
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 30 {
		limit = 10
	}

	// Получаем данные
	result, err := h.uc.GetPickupPointsWithReceptions(r.Context(), startDate, endDate, page, limit)
	if err != nil {
		logger.WithError(err).Error("failed to get pickup points with receptions")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed to get PickupPoints")
		return
	}

	// Преобразуем в ответ
	resp := make([]dto.PickupPointListResponse, 0, len(result))
	for _, item := range result {
		receptions := make([]dto.ReceptionWithProducts, 0, len(item.Receptions))
		for _, rec := range item.Receptions {
			receptions = append(receptions, dto.ReceptionWithProducts{
				Reception: rec.Reception,
				Products:  rec.Products,
			})
		}

		resp = append(resp, dto.PickupPointListResponse{
			PickupPoint: item.PickupPoint,
			Receptions:  receptions,
		})
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, resp)
}
