package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
)

//go:generate mockgen -source=reception.go -destination=../../repository/mocks/reception_repository_mock.go -package=mocks ReceptionRepository
type ReceptionRepository interface {
	CreateReception(ctx context.Context, receptionID uuid.UUID, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type ReceptionUsecase struct {
	repo ReceptionRepository
}

func NewReceptionUsecase(repo ReceptionRepository) *ReceptionUsecase {
	return &ReceptionUsecase{repo: repo}
}

func (uc *ReceptionUsecase) CreateReception(ctx context.Context, pvzID string) (*models.Reception, error) {
	const op = "ReceptionUsecase.CreateReception"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pvz_id", pvzID)

	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		logger.WithError(err).Warn("invalid pvzID")
		return nil, err
	}

	receptionID := uuid.New()

	reception, err := uc.repo.CreateReception(ctx, receptionID, uuidPvzID)
	if err != nil {
		logger.WithError(err).Error("failed to create reception")
		return nil, err
	}

	return reception, nil
}

func (uc *ReceptionUsecase) CloseReception(ctx context.Context, pvzID string) (*models.Reception, error) {
    const op = "ReceptionUsecase.CloseReception"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pvz_id", pvzID)

	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		logger.WithError(err).Warn("invalid pvzID")
		return nil, fmt.Errorf("invalid pvzId: %w", err)
	}

	reception, err := uc.repo.CloseReception(ctx, uuidPvzID)
	if err != nil {
		logger.WithError(err).Error("failed to close reception")
		return nil, err
	}

	return reception, nil
}