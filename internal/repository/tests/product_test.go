package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/product"
	repository "github.com/nik-mLb/avito_task/internal/repository/product"
	"github.com/stretchr/testify/assert"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
)

func TestAddProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repository.NewProductRepository(db)

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		productType string
		mock        func()
		expected    *models.Product
		expectedErr error
	}{
		{
			name:        "Success - Electronics",
			pvzID:       uuid.New(),
			productType: "электроника",
			mock: func() {
				receptionID := uuid.New()
				productID := uuid.New()
				now := time.Now()

				// Mock GetActiveReceptionQuery
				rows := sqlmock.NewRows([]string{"id"}).AddRow(receptionID)
				mock.ExpectQuery(`SELECT id FROM reception WHERE pickup_point_id = \$1 AND status = 'in_progress'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)

				// Mock CreateProductQuery
				rows = sqlmock.NewRows([]string{"id", "reception_id", "product_type", "reception_date"}).
					AddRow(productID, receptionID, "электроника", now)
				mock.ExpectQuery(`INSERT INTO product`).
					WithArgs(sqlmock.AnyArg(), receptionID, "электроника").
					WillReturnRows(rows)
			},
			expected: &models.Product{
				ProductType: models.ProductType("электроника"),
			},
			expectedErr: nil,
		},
		{
			name:        "Success - Clothing",
			pvzID:       uuid.New(),
			productType: "одежда",
			mock: func() {
				receptionID := uuid.New()
				productID := uuid.New()
				now := time.Now()

				// Mock GetActiveReceptionQuery
				rows := sqlmock.NewRows([]string{"id"}).AddRow(receptionID)
				mock.ExpectQuery(`SELECT id FROM reception WHERE pickup_point_id = \$1 AND status = 'in_progress'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)

				// Mock CreateProductQuery
				rows = sqlmock.NewRows([]string{"id", "reception_id", "product_type", "reception_date"}).
					AddRow(productID, receptionID, "одежда", now)
				mock.ExpectQuery(`INSERT INTO product`).
					WithArgs(sqlmock.AnyArg(), receptionID, "одежда").
					WillReturnRows(rows)
			},
			expected: &models.Product{
				ProductType: models.ProductType("одежда"),
			},
			expectedErr: nil,
		},
		{
			name:        "No Active Reception",
			pvzID:       uuid.New(),
			productType: "электроника",
			mock: func() {
				// Mock GetActiveReceptionQuery returning no rows
				mock.ExpectQuery(`SELECT id FROM reception WHERE pickup_point_id = \$1 AND status = 'in_progress'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: errs.ErrNoActiveReception,
		},
		{
			name:        "Database Error",
			pvzID:       uuid.New(),
			productType: "электроника",
			mock: func() {
				receptionID := uuid.New()

				// Mock GetActiveReceptionQuery
				rows := sqlmock.NewRows([]string{"id"}).AddRow(receptionID)
				mock.ExpectQuery(`SELECT id FROM reception WHERE pickup_point_id = \$1 AND status = 'in_progress'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)

				// Mock CreateProductQuery with error
				mock.ExpectQuery(`INSERT INTO product`).
					WithArgs(sqlmock.AnyArg(), receptionID, "электроника").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.AddProduct(context.Background(), tt.pvzID, tt.productType)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expected.ProductType, got.ProductType)
			assert.NotEmpty(t, got.ID)
			assert.NotEmpty(t, got.ReceptionID)
			assert.NotEmpty(t, got.ReceptionDate)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteLastProduct(t *testing.T) {
    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := repository.NewProductRepository(db)

    tests := []struct {
        name        string
        pvzID       uuid.UUID
        mock        func()
        expectedErr error
    }{
        {
            name:  "Success",
            pvzID: uuid.New(),
            mock: func() {
                productID := uuid.New()

                // Mock transaction begin
                mock.ExpectBegin()

                // Mock GetLastProductQuery - должно точно соответствовать запросу из репозитория
                mock.ExpectQuery(`
                    SELECT id FROM product 
                    WHERE reception_id = (
                        SELECT id FROM reception 
                        WHERE pickup_point_id = $1 AND status = 'in_progress'
                        LIMIT 1
                    )
                    ORDER BY reception_date DESC
                    LIMIT 1`).
                    WithArgs(sqlmock.AnyArg()).
                    WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

                // Mock DeleteProductQuery
                mock.ExpectExec(`
                    DELETE FROM product 
                    WHERE id = $1
                    RETURNING id`).
                    WithArgs(productID).
                    WillReturnResult(sqlmock.NewResult(1, 1))

                // Mock transaction commit
                mock.ExpectCommit()
            },
            expectedErr: nil,
        },
        {
            name:  "No Products to Delete",
            pvzID: uuid.New(),
            mock: func() {
                // Mock transaction begin
                mock.ExpectBegin()

                // Mock GetLastProductQuery returning no rows
                mock.ExpectQuery(`
                    SELECT id FROM product 
                    WHERE reception_id = (
                        SELECT id FROM reception 
                        WHERE pickup_point_id = $1 AND status = 'in_progress'
                        LIMIT 1
                    )
                    ORDER BY reception_date DESC
                    LIMIT 1`).
                    WithArgs(sqlmock.AnyArg()).
                    WillReturnError(sql.ErrNoRows)

                // Mock transaction rollback
                mock.ExpectRollback()
            },
            expectedErr: errs.ErrNoProductsToDelete,
        },
        {
            name:  "Transaction Error",
            pvzID: uuid.New(),
            mock: func() {
                // Mock transaction begin with error
                mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
            },
            expectedErr: sql.ErrConnDone,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.mock()

            err := repo.DeleteLastProduct(context.Background(), tt.pvzID)
            if tt.expectedErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.expectedErr)
                return
            }

            assert.NoError(t, err)
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}