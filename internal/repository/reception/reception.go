package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"

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

type ReceptionRepository struct {
	db *sql.DB
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

func (r *ReceptionRepository) CreateReception(ctx context.Context, receptionID uuid.UUID, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "ReceptionRepository.CreateReception"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pickup_point_id", pvzID)
	
	var exists bool
	err := r.db.QueryRowContext(ctx, CheckActiveReceptionQuery, pvzID).Scan(&exists)
	if err != nil {
		logger.WithError(err).Error("check active reception")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		logger.Warn("active reception already exists")
		return nil, errs.ErrActiveReceptionExists
	}

	reception := &models.Reception{}
	err = r.db.QueryRowContext(ctx, CreateReceptionQuery, receptionID, pvzID).
		Scan(&reception.ID, &reception.ReceptionDate, &reception.PickupPointID, &reception.Status)

	if err != nil {
		logger.WithError(err).Error("create reception")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return reception, nil
}

func (r *ReceptionRepository) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "ReceptionRepository.CloseReception"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("pickup_point_id", pvzID)

    reception := &models.Reception{}
    
    err := r.db.QueryRowContext(ctx, CloseReceptionQuery, pvzID).
        Scan(&reception.ID, &reception.ReceptionDate, &reception.PickupPointID, &reception.Status)
    
    if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("no active reception to close")
			return nil, errs.ErrNoActiveReceptionToClose
		}
		logger.WithError(err).Error("close reception")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
    
    return reception, nil
}