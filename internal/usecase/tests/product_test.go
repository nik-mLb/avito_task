package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	"github.com/nik-mLb/avito_task/internal/usecase/product"
	"github.com/nik-mLb/avito_task/internal/repository/mocks"
)

func TestProductUsecase_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	uc := usecase.NewProductUsecase(mockRepo)

	validUUID := uuid.New().String()
	validProductType := string(product.Electronics)

	t.Run("successful product addition", func(t *testing.T) {
		expectedProduct := &product.Product{
			ID:          uuid.New(),
			ReceptionID:       uuid.MustParse(validUUID),
			ProductType: product.ProductType(validProductType),
		}

		mockRepo.EXPECT().
			AddProduct(gomock.Any(), gomock.Any(), validProductType).
			Return(expectedProduct, nil)

		result, err := uc.AddProduct(context.Background(), validUUID, validProductType)

		assert.NoError(t, err)
		assert.Equal(t, expectedProduct, result)
	})

	t.Run("invalid pvzId format", func(t *testing.T) {
		invalidUUID := "invalid-uuid"

		result, err := uc.AddProduct(context.Background(), invalidUUID, validProductType)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pvzId")
		assert.Nil(t, result)
	})

	t.Run("invalid product type", func(t *testing.T) {
		invalidType := "invalid-type"

		result, err := uc.AddProduct(context.Background(), validUUID, invalidType)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid product type")
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		repoError := errors.New("repository error")

		mockRepo.EXPECT().
			AddProduct(gomock.Any(), gomock.Any(), validProductType).
			Return(nil, repoError)

		result, err := uc.AddProduct(context.Background(), validUUID, validProductType)

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		assert.Nil(t, result)
	})
}

func TestProductUsecase_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProductRepository(ctrl)
	uc := usecase.NewProductUsecase(mockRepo)

	validUUID := uuid.New().String()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteLastProduct(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.DeleteLastProduct(context.Background(), validUUID)

		assert.NoError(t, err)
	})

	t.Run("invalid pvzId format", func(t *testing.T) {
		invalidUUID := "invalid-uuid"

		err := uc.DeleteLastProduct(context.Background(), invalidUUID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pvzId")
	})

	t.Run("repository error", func(t *testing.T) {
		repoError := errors.New("repository error")

		mockRepo.EXPECT().
			DeleteLastProduct(gomock.Any(), gomock.Any()).
			Return(repoError)

		err := uc.DeleteLastProduct(context.Background(), validUUID)

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
	})
}