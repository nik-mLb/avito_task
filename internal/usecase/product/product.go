package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
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
	const op = "ProductUsecase.AddProduct"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithFields(map[string]interface{}{
		"pvz_id":      pvzID,
		"productType": productType,
	})
	
	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		logger.WithError(err).Warn("invalid pvzID")
		return nil, fmt.Errorf("invalid pvzId: %w", err)
	}

	switch models.ProductType(productType) {
	case models.Electronics, models.Clothing, models.Shoes:
		// valid type
	default:
		logger.Warn("invalid product type")
		return nil, fmt.Errorf("invalid product type: %s", productType)
	}

	product, err := uc.repo.AddProduct(ctx, uuidPvzID, productType)
	if err != nil {
		logger.WithError(err).Error("failed to add product")
		return nil, err
	}

	return product, nil
}

func (uc *ProductUsecase) DeleteLastProduct(ctx context.Context, pvzID string) error {
	const op = "ProductUsecase.DeleteLastProduct"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pvz_id", pvzID)

	uuidPvzID, err := uuid.Parse(pvzID)
	if err != nil {
		logger.WithError(err).Warn("invalid pvzID")
		return fmt.Errorf("invalid pvzId: %w", err)
	}

	err = uc.repo.DeleteLastProduct(ctx, uuidPvzID)
	if err != nil {
		logger.WithError(err).Error("failed to delete last product")
		return err
	}

	return nil
}