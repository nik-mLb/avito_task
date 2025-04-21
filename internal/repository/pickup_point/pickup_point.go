package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
)

const (
	CreatePickupPointQuery = `
		INSERT INTO pickup_point (id, city) 
		VALUES ($1, $2)
		RETURNING id, city, registration_date`

	GetPickupPointsWithReceptionsQuery = `
        SELECT 
			pp.id, pp.city, pp.registration_date,
			r.id, r.reception_date, r.status,
			p.id, p.product_type, p.reception_date
		FROM pickup_point pp
		LEFT JOIN reception r ON pp.id = r.pickup_point_id
		INNER JOIN product p ON r.id = p.reception_id  -- <- Тут INNER вместо LEFT
		WHERE ($1::timestamp IS NULL OR r.reception_date >= $1)
		AND ($2::timestamp IS NULL OR r.reception_date <= $2)
		ORDER BY pp.registration_date DESC
		LIMIT $3 OFFSET $4`
)

type PickupPointRepository struct {
	db *sql.DB
}

func NewPickupPointRepository(db *sql.DB) *PickupPointRepository {
	return &PickupPointRepository{db: db}
}

func (r *PickupPointRepository) CreatePickupPoint(ctx context.Context, city string) (*pickup.PickupPoint, error) {
	const op = "PickupPointRepository.CreatePickupPoint"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("city", city)
	
	pickupPoint := &pickup.PickupPoint{
		ID:   uuid.New(),
		City: city,
	}

	err := r.db.QueryRowContext(ctx, CreatePickupPointQuery,
		pickupPoint.ID, pickupPoint.City).
		Scan(&pickupPoint.ID, &pickupPoint.City, &pickupPoint.RegistrationDate)

	if err != nil {
		logger.WithError(err).Error("failed to create pickup point")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pickupPoint, err
}

func (r *PickupPointRepository) GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]dto.PickupPointListResponse, error) {
	const op = "PickupPointRepository.GetPickupPointsWithReceptions"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("start_date", startDate).
		WithField("end_date", endDate).
		WithField("page", page).
		WithField("limit", limit)

	offset := (page - 1) * limit

	rows, err := r.db.QueryContext(ctx, GetPickupPointsWithReceptionsQuery, startDate, endDate, limit, offset)
	if err != nil {
		logger.WithError(err).Error("failed to query pickup points")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	results := make(map[uuid.UUID]*dto.PickupPointListResponse)

	for rows.Next() {
		var (
			pp         pickup.PickupPoint
			rec        reception.Reception
			prod       product.Product
			recID      uuid.NullUUID
			prodID     uuid.NullUUID
		)

		err := rows.Scan(
			&pp.ID, &pp.City, &pp.RegistrationDate,
			&recID, &rec.ReceptionDate, &rec.Status,
			&prodID, &prod.ProductType, &prod.ReceptionDate,
		)
		if err != nil {
			scanErr := fmt.Errorf("failed to scan row: %w", err)
			logger.WithError(scanErr).Error("scan error")
			continue
		}

		// Если это новый ПВЗ, добавляем его в результаты
		if _, exists := results[pp.ID]; !exists {
			results[pp.ID] = &dto.PickupPointListResponse{
				PickupPoint: pp,
				Receptions:  []dto.ReceptionWithProducts{},
			}
		}

		// Если есть приемка, добавляем ее
		if recID.Valid {
			rec.ID = recID.UUID
			rec.PickupPointID = pp.ID

			// Ищем приемку в списке
			var foundRec *dto.ReceptionWithProducts
			for i := range results[pp.ID].Receptions {
				if results[pp.ID].Receptions[i].Reception.ID == rec.ID {
					foundRec = &results[pp.ID].Receptions[i]
					break
				}
			}

			// Если приемка новая, добавляем ее
			if foundRec == nil {
				results[pp.ID].Receptions = append(results[pp.ID].Receptions, dto.ReceptionWithProducts{
					Reception: rec,
					Products:  []product.Product{},
				})
				foundRec = &results[pp.ID].Receptions[len(results[pp.ID].Receptions)-1]
			}

			// Если есть товар, добавляем его
			if prodID.Valid {
				prod.ID = prodID.UUID
				prod.ReceptionID = rec.ID
				foundRec.Products = append(foundRec.Products, prod)
			}
		}
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("rows iteration error")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Преобразуем map в slice
	output := make([]dto.PickupPointListResponse, 0, len(results))
	for _, v := range results {
		output = append(output, *v)
	}

	return output, nil
}