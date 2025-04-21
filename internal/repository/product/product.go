package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
)

const (
	CreateProductQuery = `
		INSERT INTO product (id, reception_id, product_type, reception_date) 
		VALUES ($1, $2, $3, now())
		RETURNING id, reception_id, product_type, reception_date`

	GetActiveReceptionQuery = `
		SELECT id FROM reception 
		WHERE pickup_point_id = $1 AND status = 'in_progress'
		ORDER BY created_at DESC
		LIMIT 1`

	GetLastProductQuery = `
        SELECT id FROM product 
        WHERE reception_id = (
            SELECT id FROM reception 
            WHERE pickup_point_id = $1 AND status = 'in_progress'
            LIMIT 1
        )
        ORDER BY created_at DESC
        LIMIT 1`

    DeleteProductQuery = `
        DELETE FROM product 
        WHERE id = $1
        RETURNING id`
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) AddProduct(ctx context.Context, pvzID uuid.UUID, productType string) (*models.Product, error) {
	const op = "ProductRepository.AddProduct"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pvz_id", pvzID).WithField("product_type", productType)
    
	var receptionID uuid.UUID
	err := r.db.QueryRowContext(ctx, GetActiveReceptionQuery, pvzID).Scan(&receptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("no active reception found")
			return nil, errs.ErrNoActiveReception
		}
		logger.WithError(err).Error("query active reception")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	product := &models.Product{}
	err = r.db.QueryRowContext(ctx, CreateProductQuery, uuid.New(), receptionID, productType).
		Scan(&product.ID, &product.ReceptionID, &product.ProductType, &product.ReceptionDate)

    if err != nil {
        logger.WithError(err).Error("create product")
        return nil, fmt.Errorf("%s: %w", op, err)
    }

	return product, nil
}

func (r *ProductRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
    const op = "ProductRepository.DeleteLastProduct"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pvz_id", pvzID)

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
		logger.WithError(err).Error("begin transaction")
		return fmt.Errorf("%s: %w", op, err)
	}
    defer tx.Rollback()

    var productID uuid.UUID
    err = tx.QueryRowContext(ctx, GetLastProductQuery, pvzID).Scan(&productID)
    if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("no products to delete")
			return errs.ErrNoProductsToDelete
		}
		logger.WithError(err).Error("query last product")
		return fmt.Errorf("%s: %w", op, err)
	}

    _, err = tx.ExecContext(ctx, DeleteProductQuery, productID)
    if err != nil {
		logger.WithError(err).Error("delete product")
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		logger.WithError(err).Error("commit transaction")
		return fmt.Errorf("%s: %w", op, err)
	}

    return nil
}