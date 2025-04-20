package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
)

const (
	CreateReceptionQuery = `
		INSERT INTO reception (id, pickup_point_id, status) 
		VALUES ($1, $2, 'in_progress')
		RETURNING id, reception_date, pickup_point_id, status`

	CheckActiveReceptionQuery = `
		SELECT EXISTS (
			SELECT 1 FROM reception 
			WHERE pickup_point_id = $1 AND status = 'in_progress'
		)`

	CloseReceptionQuery = `
        UPDATE reception 
        SET status = 'close' 
        WHERE id = (
            SELECT id FROM reception 
            WHERE pickup_point_id = $1 AND status = 'in_progress'
            LIMIT 1
        )
        RETURNING id, reception_date, pickup_point_id, status`
)

var (
	ErrActiveReceptionExists = errors.New("active reception already exists")
	ErrNoActiveReceptionToClose = errors.New("no active reception to close")
)

type ReceptionRepository struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

func (r *ReceptionRepository) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	// Проверяем, есть ли активная приемка
	var exists bool
	err := r.db.QueryRowContext(ctx, CheckActiveReceptionQuery, pvzID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrActiveReceptionExists
	}

	// Создаем новую приемку
	reception := &models.Reception{}
	err = r.db.QueryRowContext(ctx, CreateReceptionQuery, uuid.New(), pvzID).
		Scan(&reception.ID, &reception.ReceptionDate, &reception.PickupPointID, &reception.Status)

	return reception, err
}

func (r *ReceptionRepository) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
    reception := &models.Reception{}
    
    err := r.db.QueryRowContext(ctx, CloseReceptionQuery, pvzID).
        Scan(&reception.ID, &reception.ReceptionDate, &reception.PickupPointID, &reception.Status)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrNoActiveReceptionToClose
        }
        return nil, err
    }
    
    return reception, nil
}