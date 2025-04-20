package transport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
	repository "github.com/nik-mLb/avito_task/internal/repository/pickup_point"
)

type PickupPointUsecase interface {
	CreatePickupPoint(ctx context.Context, city string) (*pickup.PickupPoint, error)
	GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]repository.PickupPointWithReceptions, error)
}

type PickupPointHandler struct {
	uc PickupPointUsecase
}

func NewPickupPointHandler(uc PickupPointUsecase) *PickupPointHandler {
	return &PickupPointHandler{uc: uc}
}

type createPickupPointRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
}

func (h *PickupPointHandler) CreatePickupPoint(w http.ResponseWriter, r *http.Request) {
	var req createPickupPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	PickupPoint, err := h.uc.CreatePickupPoint(r.Context(), req.City)
	if err != nil {
		if err.Error() == "city not allowed" {
			http.Error(w, "City not allowed", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create PickupPoint", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PickupPoint)
}

type getPickupPointsResponse struct {
    PickupPoint pickup.PickupPoint         `json:"PickupPoint"`
    Receptions []receptionResponse `json:"receptions"`
}

type receptionResponse struct {
    Reception reception.Reception `json:"reception"`
	Products  []product.Product   `json:"products"`
}

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
		log.Printf("Error getting pickup points: %v", err) // Добавьте это
		http.Error(w, "Failed to get PickupPoints", http.StatusInternalServerError)
		return
	}

    // Преобразуем в ответ
    response := make([]getPickupPointsResponse, 0, len(result))
    for _, item := range result {
        receptions := make([]receptionResponse, 0, len(item.Receptions))
        for _, rec := range item.Receptions {
            receptions = append(receptions, receptionResponse{
                Reception: rec.Reception,
                Products:  rec.Products,
            })
        }
        
        response = append(response, getPickupPointsResponse{
            PickupPoint:        item.PickupPoint,
            Receptions: receptions,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}