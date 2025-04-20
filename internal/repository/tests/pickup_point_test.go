package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	pickup_point "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
	repository "github.com/nik-mLb/avito_task/internal/repository/pickup_point"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/stretchr/testify/assert"
)

func TestCreatePickupPoint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repository.NewPickupPointRepository(db)

	tests := []struct {
		name        string
		city        string
		mock        func()
		expected    *pickup_point.PickupPoint
		expectedErr bool
	}{
		{
			name: "Success",
			city: "Москва",
			mock: func() {
				expectedID := uuid.New()
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "city", "registration_date"}).
					AddRow(expectedID, "Москва", now)
				
				mock.ExpectQuery(`INSERT INTO pickup_point`).
					WithArgs(sqlmock.AnyArg(), "Москва").
					WillReturnRows(rows)
			},
			expected: &pickup_point.PickupPoint{
				City: "Москва",
			},
			expectedErr: false,
		},
		{
			name: "Database Error",
			city: "Москва",
			mock: func() {
				mock.ExpectQuery(`INSERT INTO pickup_point`).
					WithArgs(sqlmock.AnyArg(), "Москва").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.CreatePickupPoint(context.Background(), tt.city)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.expected.City, got.City)
				assert.NotEqual(t, uuid.Nil, got.ID)
				assert.NotZero(t, got.RegistrationDate)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetPickupPointsWithReceptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repository.NewPickupPointRepository(db)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0) // месяц назад
	endDate := now

	tests := []struct {
		name        string
		startDate   *time.Time
		endDate     *time.Time
		page        int
		limit       int
		mock        func()
		expected    []dto.PickupPointListResponse
		expectedErr bool
	}{
		{
			name:      "Success - Single Pickup Point with Reception and Product",
			startDate: &startDate,
			endDate:   &endDate,
			page:      1,
			limit:     10,
			mock: func() {
				ppID := uuid.New()
				recID := uuid.New()
				prodID := uuid.New()

				rows := sqlmock.NewRows([]string{
					"id", "city", "registration_date",
					"id", "reception_date", "status",
					"id", "product_type", "reception_date",
				}).AddRow(
					ppID, "Москва", now,
					recID, now, "in_progress",
					prodID, "электроника", now,
				)

				mock.ExpectQuery(`SELECT`).
					WithArgs(startDate, endDate, 10, 0).
					WillReturnRows(rows)
			},
			expected: []dto.PickupPointListResponse{
				{
					PickupPoint: pickup_point.PickupPoint{
						City: "Москва",
					},
					Receptions: []dto.ReceptionWithProducts{
						{
							Reception: reception.Reception{
								Status: "in_progress",
							},
							Products: []product.Product{
								{
									ProductType: "электроника",
								},
							},
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			name:      "Success - Empty Result",
			startDate: &startDate,
			endDate:   &endDate,
			page:      1,
			limit:     10,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"id", "city", "registration_date",
					"id", "reception_date", "status",
					"id", "product_type", "reception_date",
				})

				mock.ExpectQuery(`SELECT`).
					WithArgs(startDate, endDate, 10, 0).
					WillReturnRows(rows)
			},
			expected:    []dto.PickupPointListResponse{},
			expectedErr: false,
		},
		{
			name:      "Database Error",
			startDate: &startDate,
			endDate:   &endDate,
			page:      1,
			limit:     10,
			mock: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(startDate, endDate, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetPickupPointsWithReceptions(context.Background(), tt.startDate, tt.endDate, tt.page, tt.limit)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.expected) == 0 {
					assert.Empty(t, got)
				} else {
					assert.Equal(t, len(tt.expected), len(got))
					assert.Equal(t, tt.expected[0].PickupPoint.City, got[0].PickupPoint.City)
					assert.Equal(t, len(tt.expected[0].Receptions), len(got[0].Receptions))
					if len(tt.expected[0].Receptions) > 0 {
						assert.Equal(t, tt.expected[0].Receptions[0].Reception.Status, got[0].Receptions[0].Reception.Status)
						assert.Equal(t, len(tt.expected[0].Receptions[0].Products), len(got[0].Receptions[0].Products))
						if len(tt.expected[0].Receptions[0].Products) > 0 {
							assert.Equal(t, tt.expected[0].Receptions[0].Products[0].ProductType, got[0].Receptions[0].Products[0].ProductType)
						}
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

