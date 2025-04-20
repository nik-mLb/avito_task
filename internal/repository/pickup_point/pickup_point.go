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

type PickupPointWithReceptions struct {
	PickupPoint pickup.PickupPoint
	Receptions  []ReceptionWithProducts
}

type ReceptionWithProducts struct {
	Reception reception.Reception
	Products  []product.Product
}

type PickupPointRepository struct {
	db *sql.DB
}

func NewPickupPointRepository(db *sql.DB) *PickupPointRepository {
	return &PickupPointRepository{db: db}
}

func (r *PickupPointRepository) CreatePickupPoint(ctx context.Context, city string) (*pickup.PickupPoint, error) {
	pickupPoint := &pickup.PickupPoint{
		ID:   uuid.New(),
		City: city,
	}

	err := r.db.QueryRowContext(ctx, CreatePickupPointQuery,
		pickupPoint.ID, pickupPoint.City).
		Scan(&pickupPoint.ID, &pickupPoint.City, &pickupPoint.RegistrationDate)

	return pickupPoint, err
}

func (r *PickupPointRepository) GetPickupPointsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]PickupPointWithReceptions, error) {
	offset := (page - 1) * limit

	rows, err := r.db.QueryContext(ctx, GetPickupPointsWithReceptionsQuery, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query PickupPoints: %w", err)
	}
	defer rows.Close()

	results := make(map[uuid.UUID]*PickupPointWithReceptions)

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
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Если это новый ПВЗ, добавляем его в результаты
		if _, exists := results[pp.ID]; !exists {
			results[pp.ID] = &PickupPointWithReceptions{
				PickupPoint: pp,
				Receptions:  []ReceptionWithProducts{},
			}
		}

		// Если есть приемка, добавляем ее
		if recID.Valid {
			rec.ID = recID.UUID
			rec.PickupPointID = pp.ID

			// Ищем приемку в списке
			var foundRec *ReceptionWithProducts
			for i := range results[pp.ID].Receptions {
				if results[pp.ID].Receptions[i].Reception.ID == rec.ID {
					foundRec = &results[pp.ID].Receptions[i]
					break
				}
			}

			// Если приемка новая, добавляем ее
			if foundRec == nil {
				results[pp.ID].Receptions = append(results[pp.ID].Receptions, ReceptionWithProducts{
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
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Преобразуем map в slice
	output := make([]PickupPointWithReceptions, 0, len(results))
	for _, v := range results {
		output = append(output, *v)
	}

	return output, nil
}