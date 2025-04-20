package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/product"
)

//go:generate mockgen -source=product.go -destination=../../repository/mocks/product_repository_mock.go -package=mocks ProductRepository
type ProductRepository interface {
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType string) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type ProductUsecase struct {
	repo ProductRepository
}

func NewProductUsecase(repo ProductRepository) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (uc *ProductUsecase) AddProduct(ctx context.Context, pvzID, productType string) (*models.Product, error) {
	// Валидация UUID
	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		return nil, fmt.Errorf("invalid pvzId: %w", err)
	}

	// Валидация типа товара
	switch models.ProductType(productType) {
	case models.Electronics, models.Clothing, models.Shoes:
		// valid type
	default:
		return nil, fmt.Errorf("invalid product type: %s", productType)
	}

	return uc.repo.AddProduct(ctx, uuidPvzID, productType)
}

func (uc *ProductUsecase) DeleteLastProduct(ctx context.Context, pvzID string) error {
    // Валидация UUID
    uuidPvzID, err := uuid.Parse(pvzID)
    if err != nil {
        return fmt.Errorf("invalid pvzId: %w", err)
    }

    return uc.repo.DeleteLastProduct(ctx, uuidPvzID)
}