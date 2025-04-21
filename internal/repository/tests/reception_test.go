package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/reception"
	repository "github.com/nik-mLb/avito_task/internal/repository/reception"
)

func TestCreateReception(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewReceptionRepository(db)

	receptionID := uuid.MustParse("9a080ac9-7577-4e9c-97ab-2a0de0e55fad")
	pvzID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	now := time.Date(2025, 4, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		mock        func()
		expected    *models.Reception
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery(repository.CheckActiveReceptionQuery).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				mock.ExpectQuery(repository.CreateReceptionQuery).
					WithArgs(receptionID, pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "reception_date", "pickup_point_id", "status"}).
						AddRow(receptionID, now, pvzID, "in_progress"))
			},
			expected: &models.Reception{
				ID:             receptionID,
				ReceptionDate:  now,
				PickupPointID:  pvzID,
				Status:         "in_progress",
			},
			expectedErr: nil,
		},
		{
			name: "Active Reception Exists",
			mock: func() {
				mock.ExpectQuery(repository.CheckActiveReceptionQuery).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected:    nil,
			expectedErr: errs.ErrActiveReceptionExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.CreateReception(context.Background(), receptionID, pvzID)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCloseReception(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewReceptionRepository(db)

	receptionID := uuid.MustParse("4e94cf16-5b74-4d7b-88d2-3334501329b5")
	pvzID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	now := time.Date(2025, 4, 20, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		mock        func()
		expected    *models.Reception
		expectedErr error
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery(repository.CloseReceptionQuery).
					WithArgs(pvzID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "reception_date", "pickup_point_id", "status"}).
						AddRow(receptionID, now, pvzID, "close"))
			},
			expected: &models.Reception{
				ID:             receptionID,
				ReceptionDate:  now,
				PickupPointID:  pvzID,
				Status:         "close",
			},
			expectedErr: nil,
		},
		{
			name: "No Active Reception",
			mock: func() {
				mock.ExpectQuery(repository.CloseReceptionQuery).
					WithArgs(pvzID).
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: errs.ErrNoActiveReceptionToClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.CloseReception(context.Background(), pvzID)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
