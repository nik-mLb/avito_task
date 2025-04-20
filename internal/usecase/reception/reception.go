package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
)

type ReceptionRepository interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type ReceptionUsecase struct {
	repo ReceptionRepository
}

func NewReceptionUsecase(repo ReceptionRepository) *ReceptionUsecase {
	return &ReceptionUsecase{repo: repo}
}

func (uc *ReceptionUsecase) CreateReception(ctx context.Context, pvzID string) (*models.Reception, error) {
	// Валидация UUID
	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		return nil, err
	}

	return uc.repo.CreateReception(ctx, uuidPvzID)
}

func (uc *ReceptionUsecase) CloseReception(ctx context.Context, pvzID string) (*models.Reception, error) {
    // Валидация UUID
    uuidPvzID, err := uuid.Parse(pvzID)
    if err != nil {
        return nil, fmt.Errorf("invalid pvzId: %w", err)
    }

    return uc.repo.CloseReception(ctx, uuidPvzID)
}