package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	reception "github.com/nik-mLb/avito_task/internal/models/reception"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nik-mLb/avito_task/internal/usecase/reception"
	mocks "github.com/nik-mLb/avito_task/internal/repository/mocks"
)

func TestCreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReceptionRepository(ctrl)
	uc := usecase.NewReceptionUsecase(mockRepo)

	ctx := context.Background()
	testPvzID := uuid.New().String()
	uuidPvzID, _ := uuid.Parse(testPvzID)
	expectedReceptionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedReception := &reception.Reception{
			ID:            expectedReceptionID,
			ReceptionDate: time.Now(),
			PickupPointID: uuidPvzID,
			Status:        "in_progress",
		}

		mockRepo.EXPECT().
			CreateReception(ctx, gomock.Any(), uuidPvzID).
			DoAndReturn(func(_ context.Context, receptionID uuid.UUID, _ uuid.UUID) (*reception.Reception, error) {
				assert.NotEqual(t, uuid.Nil, receptionID)
				return expectedReception, nil
			})

		result, err := uc.CreateReception(ctx, testPvzID)

		assert.NoError(t, err)
		assert.Equal(t, expectedReception, result)
	})

	t.Run("invalid pvzId", func(t *testing.T) {
		_, err := uc.CreateReception(ctx, "invalid-uuid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository error")

		mockRepo.EXPECT().
			CreateReception(ctx, gomock.Any(), uuidPvzID).
			Return(nil, expectedErr)

		_, err := uc.CreateReception(ctx, testPvzID)

		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCloseReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReceptionRepository(ctrl)
	uc := usecase.NewReceptionUsecase(mockRepo)

	ctx := context.Background()
	testPvzID := uuid.New().String()
	uuidPvzID, _ := uuid.Parse(testPvzID)

	t.Run("success", func(t *testing.T) {
		expectedReception := &reception.Reception{
			ID:            uuid.New(),
			ReceptionDate: time.Now(),
			PickupPointID: uuidPvzID,
			Status:        "close",
		}

		mockRepo.EXPECT().
			CloseReception(ctx, uuidPvzID).
			Return(expectedReception, nil)

		result, err := uc.CloseReception(ctx, testPvzID)

		assert.NoError(t, err)
		assert.Equal(t, expectedReception, result)
	})

	t.Run("invalid pvzId", func(t *testing.T) {
		_, err := uc.CloseReception(ctx, "invalid-uuid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pvzId")
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("repository error")

		mockRepo.EXPECT().
			CloseReception(ctx, uuidPvzID).
			Return(nil, expectedErr)

		_, err := uc.CloseReception(ctx, testPvzID)

		assert.ErrorIs(t, err, expectedErr)
	})
}