package usecase

import (
	"context"
	"errors"
	"time"

	models "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	repository "github.com/nik-mLb/avito_task/internal/repository/pickup_point"
)

var (
	allowedCities = map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}
	ErrCityNotAllowed = errors.New("city not allowed")
)

type PickupPointRepository interface {
	CreatePickupPoint(ctx context.Context, city string) (*models.PickupPoint, error)
	GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]repository.PickupPointWithReceptions, error)
}

type PickupPointUsecase struct {
	repo PickupPointRepository
}

func NewPickupPointUsecase(repo PickupPointRepository) *PickupPointUsecase {
	return &PickupPointUsecase{repo: repo}
}

func (uc *PickupPointUsecase) CreatePickupPoint(ctx context.Context, city string) (*models.PickupPoint, error) {
	if !allowedCities[city] {
		return nil, ErrCityNotAllowed
	}

	return uc.repo.CreatePickupPoint(ctx, city)
}

func (uc *PickupPointUsecase) GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]repository.PickupPointWithReceptions, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 30 {
        limit = 10
    }

    return uc.repo.GetPickupPointsWithReceptions(ctx, startDate, endDate, page, limit)
}