package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
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

// CreatePickupPoint godoc
//	@Summary		Создать новый пункт выдачи
//	@Description	Создает новый пункт выдачи заказов (ПВЗ)
//	@Tags			PVZ
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			Authorization	header		string					true	"Токен авторизации (Bearer)"
//	@Param			request			body		dto.PickupPointRequest	true	"Данные для создания ПВЗ"
//	@Success		201				{object}	pickup.PickupPoint
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректный запрос или город не разрешен"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		403				{object}	dto.ErrorResponse	"Нет прав доступа"
//	@Failure		500				{object}	dto.ErrorResponse	"Ошибка сервера"
//	@Router			/pvz [post]
func (h *PickupPointHandler) CreatePickupPoint(w http.ResponseWriter, r *http.Request) {
	var req dto.PickupPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	PickupPoint, err := h.uc.CreatePickupPoint(r.Context(), req.City)
	if err != nil {
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

// GetPickupPointsWithReceptions godoc
//	@Summary		Получить список ПВЗ с приемками
//	@Description	Возвращает список пунктов выдачи с информацией о приемках товаров за указанный период
//	@Tags			PVZ
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			Authorization	header		string	true	"Токен авторизации (Bearer)"
//	@Param			startDate		query		string	false	"Начальная дата периода (формат RFC3339, например: 2023-01-01T00:00:00Z)"
//	@Param			endDate			query		string	false	"Конечная дата периода (формат RFC3339, например: 2023-12-31T23:59:59Z)"
//	@Param			page			query		int		false	"Номер страницы (по умолчанию 1)"
//	@Param			limit			query		int		false	"Количество элементов на странице (по умолчанию 10, максимум 30)"
//	@Success		200				{array}		dto.PickupPointListResponse
//	@Failure		400				{object}	dto.ErrorResponse	"Некорректные параметры запроса"
//	@Failure		401				{object}	dto.ErrorResponse	"Не авторизован"
//	@Failure		403				{object}	dto.ErrorResponse	"Нет прав доступа"
//	@Failure		500				{object}	dto.ErrorResponse	"Ошибка сервера"
//	@Router			/pvz [get]
func (h *PickupPointHandler) GetPickupPointsWithReceptions(w http.ResponseWriter, r *http.Request) {
	// Парсим параметры запроса
	query := r.URL.Query()

	// Парсим даты
	var startDate, endDate *time.Time
	if startStr := query.Get("startDate"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := query.Get("endDate"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
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
