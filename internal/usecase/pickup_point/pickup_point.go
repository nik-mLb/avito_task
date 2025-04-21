package usecase

import (
	"context"
	"time"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
)

var (
	allowedCities = map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}
)

//go:generate mockgen -source=pickup_point.go -destination=../../repository/mocks/pickup_point_repository_mock.go -package=mocks PickupPointRepository
type PickupPointRepository interface {
	CreatePickupPoint(ctx context.Context, city string) (*models.PickupPoint, error)
	GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]dto.PickupPointListResponse, error)
}

type PickupPointUsecase struct {
	repo PickupPointRepository
}

func NewPickupPointUsecase(repo PickupPointRepository) *PickupPointUsecase {
	return &PickupPointUsecase{repo: repo}
}

func (uc *PickupPointUsecase) CreatePickupPoint(ctx context.Context, city string) (*models.PickupPoint, error) {
	const op = "PickupPointUsecase.CreatePickupPoint"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("city", city)

	if !allowedCities[city] {
		logger.Warn("attempt to create pickup point in disallowed city")
		return nil, errs.ErrCityNotAllowed
	}

	pvz, err := uc.repo.CreatePickupPoint(ctx, city)
	if err != nil {
		logger.WithError(err).Error("failed to create pickup point")
		return nil, err
	}

	return pvz, nil
}

func (uc *PickupPointUsecase) GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]dto.PickupPointListResponse, error) {
    const op = "PickupPointUsecase.GetPickupPointsWithReceptions"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithFields(map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
		"page":      page,
		"limit":     limit,
	})

	if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 30 {
        limit = 10
    }

    list, err := uc.repo.GetPickupPointsWithReceptions(ctx, startDate, endDate, page, limit)
	if err != nil {
		logger.WithError(err).Error("failed to get pickup points with receptions")
		return nil, err
	}

	return list, nil
}