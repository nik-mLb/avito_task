package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
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
	// Получаем активную приемку
	var receptionID uuid.UUID
	err := r.db.QueryRowContext(ctx, GetActiveReceptionQuery, pvzID).Scan(&receptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNoActiveReception
		}
		return nil, err
	}

	// Создаем товар
	product := &models.Product{}
	err = r.db.QueryRowContext(ctx, CreateProductQuery, uuid.New(), receptionID, productType).
		Scan(&product.ID, &product.ReceptionID, &product.ProductType, &product.ReceptionDate)

	return product, err
}

func (r *ProductRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
    // Начинаем транзакцию
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Получаем ID последнего товара
    var productID uuid.UUID
    err = tx.QueryRowContext(ctx, GetLastProductQuery, pvzID).Scan(&productID)
    if err != nil {
        if err == sql.ErrNoRows {
            return errs.ErrNoProductsToDelete
        }
        return err
    }

    // 2. Удаляем товар
    _, err = tx.ExecContext(ctx, DeleteProductQuery, productID)
    if err != nil {
        return err
    }

    // Фиксируем транзакцию
    return tx.Commit()
}