package tests

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	usecase "github.com/nik-mLb/avito_task/internal/usecase/pickup_point"
	"github.com/nik-mLb/avito_task/internal/repository/mocks"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
)

func TestPickupPointUsecase_CreatePickupPoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPickupPointRepository(ctrl)
	uc := usecase.NewPickupPointUsecase(mockRepo)

	t.Run("not allowed city", func(t *testing.T) {
		city := "Новосибирск"

		result, err := uc.CreatePickupPoint(context.Background(), city)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrCityNotAllowed, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		city := "Казань"

		mockRepo.EXPECT().
			CreatePickupPoint(gomock.Any(), city).
			Return(nil, assert.AnError)

		result, err := uc.CreatePickupPoint(context.Background(), city)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, result)
	})
}

func TestPickupPointUsecase_GetPickupPointsWithReceptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPickupPointRepository(ctrl)
	uc := usecase.NewPickupPointUsecase(mockRepo)

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetPickupPointsWithReceptions(gomock.Any(), &startDate, &endDate, 2, 5).
			Return(nil, assert.AnError)

		result, err := uc.GetPickupPointsWithReceptions(context.Background(), &startDate, &endDate, 2, 5)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, result)
	})
}